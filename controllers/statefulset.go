/*
Copyright 2023 Solace Corporation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	ctrl "sigs.k8s.io/controller-runtime"

	eventbrokerv1beta1 "github.com/SolaceProducts/pubsubplus-operator/api/v1beta1"
)

// statefulsetForEventBroker returns a new pubsubpluseventbroker StatefulSet object
func (r *PubSubPlusEventBrokerReconciler) createStatefulsetForEventBroker(stsName string, ctx context.Context, m *eventbrokerv1beta1.PubSubPlusEventBroker, sa *corev1.ServiceAccount, adminSecret *corev1.Secret, preSharedAuthKeySecret *corev1.Secret, monitoringSecret *corev1.Secret) *appsv1.StatefulSet {
	nodeType := getBrokerNodeType(stsName)

	// Determine broker sizing
	var storageSize string
	if nodeType == "monitor" {
		if len(strings.TrimSpace(m.Spec.Storage.MonitorNodeStorageSize)) == 0 || m.Spec.Storage.MonitorNodeStorageSize == "0" {
			storageSize = "3Gi"
		} else {
			storageSize = m.Spec.Storage.MonitorNodeStorageSize
		}
	} else {
		storageSize = (map[bool]string{true: "7Gi", false: getBrokerMessageNodeStorageSize(&m.Spec.Storage)})[m.Spec.Developer]
	}
	dep := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      stsName,
			Namespace: m.Namespace,
			Labels:    getObjectLabels(m.Name),
		},
		// Followings are immutable fields of the StatefulSet - cannot be part of the update
		Spec: appsv1.StatefulSetSpec{
			Selector: &metav1.LabelSelector{ // Refers to the broker Pod labels - see template below
				MatchLabels: getPodLabels(m.Name, nodeType),
			},
			Replicas:                             &[]int32{1}[0], // Set to 1
			ServiceName:                          getObjectName("DiscoveryService", m.Name),
			PodManagementPolicy:                  "",
			UpdateStrategy:                       appsv1.StatefulSetUpdateStrategy{},
			RevisionHistoryLimit:                 new(int32),
			MinReadySeconds:                      0,
			PersistentVolumeClaimRetentionPolicy: &appsv1.StatefulSetPersistentVolumeClaimRetentionPolicy{},
		},
	}

	if len(m.Spec.Storage.CustomVolumeMount) == 0 && !usesEphemeralStorageForMonitoringNode(&m.Spec.Storage, nodeType) && !usesEphemeralStorageForMessageNode(&m.Spec.Storage, nodeType) {
		dep.Spec.VolumeClaimTemplates = []corev1.PersistentVolumeClaim{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "data",
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
					Resources: corev1.VolumeResourceRequirements{
						Requests: map[corev1.ResourceName]resource.Quantity{
							corev1.ResourceStorage: resource.MustParse(storageSize),
						},
					},
				},
			},
		}

		//Set StorageClass
		if len(strings.TrimSpace(m.Spec.Storage.UseStorageClass)) > 0 {
			dep.Spec.VolumeClaimTemplates[0].Spec.StorageClassName = &m.Spec.Storage.UseStorageClass
		}
	}

	r.updateStatefulsetForEventBroker(dep, ctx, m, sa, adminSecret, preSharedAuthKeySecret, monitoringSecret)
	// Set PubSubPlusEventBroker instance as the owner and controller
	ctrl.SetControllerReference(m, dep, r.Scheme)
	return dep
}

// statefulsetForEventBroker returns an updated pubsubpluseventbroker StatefulSet object
func (r *PubSubPlusEventBrokerReconciler) updateStatefulsetForEventBroker(sts *appsv1.StatefulSet, ctx context.Context, m *eventbrokerv1beta1.PubSubPlusEventBroker, sa *corev1.ServiceAccount, adminSecret *corev1.Secret, preSharedAuthKeySecret *corev1.Secret, monitoringSecret *corev1.Secret) {
	DefaultServiceConfig, _ := scripts.ReadFile("configs/default-service.json")
	brokerServicesName := getObjectName("BrokerService", m.Name)
	adminSecretName := adminSecret.Name
	configmapName := getObjectName("ConfigMap", m.Name)
	haDeployment := m.Spec.Redundancy
	stsName := sts.ObjectMeta.Name
	nodeType := getBrokerNodeType(stsName)

	// Determine broker sizing
	var cpuRequests, cpuLimits string
	var memRequests, memLimits string
	var maxConnections, maxQueueMessages, maxSpoolUsage int
	// TODO: _types.go has already defaults. Review if those indeed need to be duplicated here.
	// TODO: remove any hardcoded values - at a minimum move it to the beginning of this code file
	if nodeType == "monitor" {
		cpuRequests = "1"
		cpuLimits = "1"
		memRequests = "2Gi"
		memLimits = "2Gi"
		maxConnections = 100
		maxQueueMessages = 100
		maxSpoolUsage = 1000
	} else {
		// First determine default settings for the message routing broker nodes, depending on developer mode set
		// refer to https://docs.solace.com/Admin-Ref/Resource-Calculator/pubsubplus-resource-calculator.html
		cpuRequests = (map[bool]string{true: "1", false: "2"})[m.Spec.Developer]
		cpuLimits = (map[bool]string{true: "2", false: "2"})[m.Spec.Developer]
		memRequests = (map[bool]string{true: "3410Mi", false: "4025Mi"})[m.Spec.Developer]
		memLimits = (map[bool]string{true: "3410Mi", false: "4025Mi"})[m.Spec.Developer]
		maxConnections = (map[bool]int{true: 100, false: 100})[m.Spec.Developer]
		maxQueueMessages = (map[bool]int{true: 100, false: 100})[m.Spec.Developer]
		maxSpoolUsage = (map[bool]int{true: 1000, false: 10000})[m.Spec.Developer]
		// Overwrite for any values defined in spec.systemScaling
		if m.Spec.SystemScaling != nil && !m.Spec.Developer {
			if m.Spec.SystemScaling.MessagingNodeCpu != "" {
				cpuRequests = m.Spec.SystemScaling.MessagingNodeCpu
				cpuLimits = cpuRequests
			}
			if m.Spec.SystemScaling.MessagingNodeMemory != "" {
				memRequests = m.Spec.SystemScaling.MessagingNodeMemory
				memLimits = memRequests
			}
			if m.Spec.SystemScaling.MaxConnections > 0 {
				maxConnections = m.Spec.SystemScaling.MaxConnections
			}
			if m.Spec.SystemScaling.MaxQueueMessages > 0 {
				maxQueueMessages = m.Spec.SystemScaling.MaxQueueMessages
			}
			if m.Spec.SystemScaling.MaxSpoolUsage > 0 {
				maxSpoolUsage = m.Spec.SystemScaling.MaxSpoolUsage
			}
		}
	}

	// Update fields

	podLabels := getPodLabels(m.Name, nodeType)
	configPodLabels := m.Spec.PodLabels
	if len(configPodLabels) > 0 {
		for k, v := range m.Spec.PodLabels {
			_, exists := podLabels[k]
			if !exists {
				podLabels[k] = v
			}
		}
	}

	podAnnotations := map[string]string{
		brokerSpecSignatureAnnotationName: brokerSpecHash(m.Spec),
		tlsSecretSignatureAnnotationName:  r.tlsSecretHash(ctx, m),
	}
	if len(m.Spec.PodAnnotations) > 0 {
		for k, v := range m.Spec.PodAnnotations {
			_, exists := podAnnotations[k]
			if !exists {
				podAnnotations[k] = v
			}
		}
	}

	sts.Spec.UpdateStrategy = appsv1.StatefulSetUpdateStrategy{
		Type: appsv1.OnDeleteStatefulSetStrategyType,
	}
	sts.Spec.Template = corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: podLabels,
			// Place to note the resource version of upstream objects
			Annotations: podAnnotations,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:            "pubsubplus",
					Image:           r.getBrokerImageDetails(&m.Spec.BrokerImage),
					ImagePullPolicy: m.Spec.BrokerImage.ImagePullPolicy,
					Resources: corev1.ResourceRequirements{
						Limits: map[corev1.ResourceName]resource.Quantity{
							corev1.ResourceCPU:    resource.MustParse(cpuLimits),
							corev1.ResourceMemory: resource.MustParse(memLimits),
						},
						Requests: map[corev1.ResourceName]resource.Quantity{
							corev1.ResourceCPU:    resource.MustParse(cpuRequests),
							corev1.ResourceMemory: resource.MustParse(memRequests),
						},
					},
					Command: []string{
						"bash",
						"-ec",
						"source /mnt/disks/solace/init.sh\nnohup /mnt/disks/solace/startup-broker.sh &\n/usr/sbin/boot.sh",
					},
					Env: []corev1.EnvVar{
						{
							Name:  "STATEFULSET_NAME",
							Value: stsName,
						},
						{
							Name: "STATEFULSET_NAMESPACE",
							ValueFrom: &corev1.EnvVarSource{
								FieldRef: &corev1.ObjectFieldSelector{
									FieldPath: "metadata.namespace",
								},
							},
						},
						{
							Name:  "BROKERSERVICES_NAME",
							Value: brokerServicesName,
						},
						{
							Name:  "BROKER_MAXCONNECTIONCOUNT",
							Value: strconv.Itoa(maxConnections),
						},
						{
							Name:  "BROKER_MAXQUEUEMESSAGECOUNT",
							Value: strconv.Itoa(maxQueueMessages),
						},
						{
							Name:  "BROKER_MAXSPOOLUSAGE",
							Value: strconv.Itoa(maxSpoolUsage),
						},
						{
							Name:  "BROKER_TLS_ENABLED",
							Value: strconv.FormatBool(m.Spec.BrokerTLS.Enabled),
						},
						{
							Name:  "BROKER_CERT_FILENAME",
							Value: m.Spec.BrokerTLS.TLSCertName,
						},
						{
							Name:  "BROKER_CERTKEY_FILENAME",
							Value: m.Spec.BrokerTLS.TLSCertKeyName,
						},
						{
							Name:  "BROKER_REDUNDANCY",
							Value: strconv.FormatBool(haDeployment),
						},
						{
							Name:  "TZ",
							Value: ":/usr/share/zoneinfo/" + getTimezone(m.Spec.Timezone),
						},
						{
							Name:  "UMASK",
							Value: "0022",
						},
					},
					LivenessProbe: &corev1.Probe{
						ProbeHandler: corev1.ProbeHandler{
							TCPSocket: &corev1.TCPSocketAction{
								Port: intstr.IntOrString{Type: intstr.Int, IntVal: int32(8080)},
							},
						},
						InitialDelaySeconds: 300,
						TimeoutSeconds:      5,
						PeriodSeconds:       10,
						SuccessThreshold:    1,
						FailureThreshold:    3,
					},
					ReadinessProbe: &corev1.Probe{
						ProbeHandler: corev1.ProbeHandler{
							Exec: &corev1.ExecAction{
								Command: []string{
									"/mnt/disks/solace/readiness_check.sh",
								},
							},
						},
						InitialDelaySeconds: 30,
						TimeoutSeconds:      1,
						PeriodSeconds:       5,
						SuccessThreshold:    1,
						FailureThreshold:    3,
					},
					Lifecycle: &corev1.Lifecycle{
						PreStop: &corev1.LifecycleHandler{
							Exec: &corev1.ExecAction{
								Command: []string{
									"bash",
									"-ec",
									"while ! pgrep solacedaemon ; do sleep 1; done\nkillall solacedaemon;\nwhile [ ! -d /usr/sw/var/db.upgrade ]; do sleep 1; done;",
								},
							},
						},
					},
					SecurityContext: &corev1.SecurityContext{
						Privileged: &[]bool{false}[0], // Set to false
						Capabilities: &corev1.Capabilities{
							Drop: []corev1.Capability{
								corev1.Capability("ALL"),
							},
						},
						RunAsNonRoot:             &[]bool{true}[0],  // Set to true
						AllowPrivilegeEscalation: &[]bool{false}[0], // Set to false
						SeccompProfile: &corev1.SeccompProfile{
							Type: corev1.SeccompProfileTypeRuntimeDefault,
						},
					},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "podinfo",
							MountPath: "/etc/podinfo",
						},
						{
							Name:      "config-map",
							MountPath: "/mnt/disks/solace",
						},
						{
							Name:      "secrets",
							ReadOnly:  true,
							MountPath: "/mnt/disks/secrets/admin",
						},
						{
							Name:      "monitoring-secrets",
							ReadOnly:  true,
							MountPath: "/mnt/disks/secrets/monitoring",
						},
						{
							Name:      "dshm",
							MountPath: "/dev/shm",
						},
						{
							Name:      "data",
							MountPath: "/var/lib/solace",
						},
					},
				},
			},
			RestartPolicy:                 corev1.RestartPolicyAlways,
			TerminationGracePeriodSeconds: &[]int64{1200}[0], // 1200
			ServiceAccountName:            sa.Name,
			// NodeName:                      "",
			Volumes: []corev1.Volume{
				{
					Name: "podinfo",
					VolumeSource: corev1.VolumeSource{
						DownwardAPI: &corev1.DownwardAPIVolumeSource{
							Items: []corev1.DownwardAPIVolumeFile{
								{
									Path: "labels",
									FieldRef: &corev1.ObjectFieldSelector{
										APIVersion: "v1",
										FieldPath:  "metadata.labels",
									},
								},
							},
							DefaultMode: &[]int32{420}[0], // 420
						},
					},
				},
				{
					Name: "config-map",
					VolumeSource: corev1.VolumeSource{
						ConfigMap: &corev1.ConfigMapVolumeSource{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: configmapName,
							},
							DefaultMode: &[]int32{493}[0], // 493
						},
					},
				},
				{
					Name: "secrets",
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName:  adminSecretName,
							DefaultMode: &[]int32{256}[0], // 256
						},
					},
				}, {
					Name: "monitoring-secrets",
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName:  monitoringSecret.Name,
							DefaultMode: &[]int32{256}[0], // 256
						},
					},
				},
				{
					Name: "dshm",
					VolumeSource: corev1.VolumeSource{
						EmptyDir: &corev1.EmptyDirVolumeSource{
							Medium: corev1.StorageMediumMemory,
						},
					},
				},
			},
			ImagePullSecrets: m.Spec.BrokerImage.ImagePullSecrets,
			// SchedulerName:                 "",
			// Tolerations:                   []corev1.Toleration{},
			// TopologySpreadConstraints:     []corev1.TopologySpreadConstraint{},
		},
	}

	//Set custom volume
	if len(m.Spec.Storage.CustomVolumeMount) > 0 {
		allVolumes := sts.Spec.Template.Spec.Volumes
		for _, customVolume := range m.Spec.Storage.CustomVolumeMount {
			if strings.Contains(
				strings.ToLower(nodeType),
				strings.ToLower(customVolume.Name),
			) {
				allVolumes = append(allVolumes, corev1.Volume{
					Name: "data",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: customVolume.PersistentVolumeClaim.ClaimName,
							ReadOnly:  false,
						},
					},
				})
			}
		}
		sts.Spec.Template.Spec.Volumes = allVolumes
	}

	// Set pod security context
	// Following cases are distinguished for RunAsUser and FSGroup: 1) if value not specified AND in OpenShift env AND using non-default namespace, then leave it to unspecified
	// 2) value not specified or using default namespace: set to default 3) value specified, then set to value.
	sts.Spec.Template.Spec.SecurityContext = &corev1.PodSecurityContext{}
	// Set RunAsUser
	if m.Spec.SecurityContext.RunAsUser == 0 {
		// not specified case
		if !r.IsOpenShift || sts.ObjectMeta.Namespace == corev1.NamespaceDefault {
			sts.Spec.Template.Spec.SecurityContext.RunAsUser = &[]int64{1000001}[0]
		} // else in OpenShift env AND using non-default namespace so leave it undefined
	} else {
		sts.Spec.Template.Spec.SecurityContext.RunAsUser = &m.Spec.SecurityContext.RunAsUser
	}
	// Set FSGroup
	if m.Spec.SecurityContext.FSGroup == 0 {
		// not specified case
		if !r.IsOpenShift || sts.ObjectMeta.Namespace == corev1.NamespaceDefault {
			sts.Spec.Template.Spec.SecurityContext.FSGroup = &[]int64{1000002}[0]
		} // else in OpenShift env AND using non-default namespace so leave it undefined
	} else {
		sts.Spec.Template.Spec.SecurityContext.FSGroup = &m.Spec.SecurityContext.FSGroup
	}

	//Set TLS configuration
	if m.Spec.BrokerTLS.Enabled {
		allVolumes := sts.Spec.Template.Spec.Volumes
		allVolumes = append(allVolumes, corev1.Volume{
			Name: "server-certs",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName:  m.Spec.BrokerTLS.ServerTLsConfigSecret,
					DefaultMode: &[]int32{0400}[0],
				},
			},
		})
		allContainerVolumeMounts := sts.Spec.Template.Spec.Containers[0].VolumeMounts
		allContainerVolumeMounts = append(allContainerVolumeMounts, corev1.VolumeMount{
			Name:      "server-certs",
			MountPath: "/mnt/disks/certs/server",
			ReadOnly:  true,
		})
		sts.Spec.Template.Spec.Volumes = allVolumes
		sts.Spec.Template.Spec.Containers[0].VolumeMounts = allContainerVolumeMounts
	}

	//Mount PreSharedAuthSecret in HA mode
	if m.Spec.Redundancy {
		allVolumes := sts.Spec.Template.Spec.Volumes
		allVolumes = append(allVolumes, corev1.Volume{
			Name: "presharedauthkey-secret",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName:  preSharedAuthKeySecret.Name,
					DefaultMode: &[]int32{256}[0], // 256
				},
			},
		})
		allContainerVolumeMounts := sts.Spec.Template.Spec.Containers[0].VolumeMounts
		allContainerVolumeMounts = append(allContainerVolumeMounts, corev1.VolumeMount{
			Name:      "presharedauthkey-secret",
			MountPath: "/mnt/disks/secrets/presharedauthkey",
			ReadOnly:  true,
		})
		sts.Spec.Template.Spec.Volumes = allVolumes
		sts.Spec.Template.Spec.Containers[0].VolumeMounts = allContainerVolumeMounts
	}

	//Set Service Port configuration
	if len(m.Spec.Service.Ports) > 0 {
		ports := make([]corev1.ContainerPort, len(m.Spec.Service.Ports))
		for idx, pbPort := range m.Spec.Service.Ports {
			ports[idx] = corev1.ContainerPort{
				Name:          pbPort.Name,
				Protocol:      pbPort.Protocol,
				ContainerPort: pbPort.ContainerPort,
			}
		}
		sts.Spec.Template.Spec.Containers[0].Ports = ports
	} else {
		portConfig := eventbrokerv1beta1.Service{}
		err := json.Unmarshal([]byte(DefaultServiceConfig), &portConfig)
		if err == nil {
			ports := make([]corev1.ContainerPort, len(portConfig.Ports))
			for idx, pbPort := range portConfig.Ports {
				ports[idx] = corev1.ContainerPort{
					Name:          pbPort.Name,
					Protocol:      pbPort.Protocol,
					ContainerPort: pbPort.ContainerPort,
				}
			}
			sts.Spec.Template.Spec.Containers[0].Ports = ports
		}
	}

	//Set Extra environment variables
	if len(m.Spec.ExtraEnvVars) > 0 {
		allEnv := sts.Spec.Template.Spec.Containers[0].Env
		for _, envV := range m.Spec.ExtraEnvVars {
			allEnv = append(allEnv, corev1.EnvVar{
				Name:  envV.Name,
				Value: envV.Value,
			})
		}
		sts.Spec.Template.Spec.Containers[0].Env = allEnv
	}

	//Set unknown scaling parameter values
	if m.Spec.SystemScaling != nil {
		var scalingParamMap map[string]interface{}
		inrec, _ := json.Marshal(m.Spec.SystemScaling)
		json.Unmarshal(inrec, &scalingParamMap)
		allEnv := sts.Spec.Template.Spec.Containers[0].Env
		for key, val := range scalingParamMap {
			println("Adding extra scaling param key " + key)
			println("Adding extra scaling param value " + fmt.Sprint(val))
			allEnv = append(allEnv, corev1.EnvVar{
				Name:  key,
				Value: fmt.Sprint(val),
			})
		}
		sts.Spec.Template.Spec.Containers[0].Env = allEnv
	}

	allEnvFrom := []corev1.EnvFromSource{}
	//Set Extra adminSecret environment variables
	if len(strings.TrimSpace(m.Spec.ExtraEnvVarsSecret)) > 0 {
		allEnvFrom = append(allEnvFrom, corev1.EnvFromSource{
			SecretRef: &corev1.SecretEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: m.Spec.ExtraEnvVarsSecret,
				},
			},
		})
	}

	//Set Extra configmap environment variables
	if len(strings.TrimSpace(m.Spec.ExtraEnvVarsCM)) > 0 {
		allEnvFrom = append(allEnvFrom, corev1.EnvFromSource{
			ConfigMapRef: &corev1.ConfigMapEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: m.Spec.ExtraEnvVarsCM,
				},
			},
		})
	}
	sts.Spec.Template.Spec.Containers[0].EnvFrom = allEnvFrom

	//Set volume configuration for when storage is slow
	if m.Spec.Storage.Slow {
		allVolumes := sts.Spec.Template.Spec.Volumes
		allVolumes = append(allVolumes, corev1.Volume{
			Name: "soft-adb-ephemeral",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		})
		allContainerVolumeMounts := sts.Spec.Template.Spec.Containers[0].VolumeMounts
		allContainerVolumeMounts = append(allContainerVolumeMounts, corev1.VolumeMount{
			Name:      "soft-adb-ephemeral",
			MountPath: "/var/lib/solace/spool-cache",
		})
		sts.Spec.Template.Spec.Volumes = allVolumes
		sts.Spec.Template.Spec.Containers[0].VolumeMounts = allContainerVolumeMounts
	}

	//determine storage type is ephemeral
	var useEphemeralStorageForMonitoringNode = usesEphemeralStorageForMonitoringNode(&m.Spec.Storage, nodeType)
	var useEphemeralStorageForMessageNode = usesEphemeralStorageForMessageNode(&m.Spec.Storage, nodeType)

	if useEphemeralStorageForMessageNode || useEphemeralStorageForMonitoringNode {
		allVolumes := sts.Spec.Template.Spec.Volumes
		if useEphemeralStorageForMonitoringNode && nodeType == "monitor" {
			allVolumes = append(allVolumes, corev1.Volume{
				Name: "data",
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{},
				},
			})
		} else if useEphemeralStorageForMessageNode && nodeType != "monitor" {
			allVolumes = append(allVolumes, corev1.Volume{
				Name: "data",
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{},
				},
			})
		}
		sts.Spec.Template.Spec.Volumes = allVolumes
	}

	nodeSelectorConfiguration := getNodeSelectorDetails(m.Spec.BrokerNodeAssignment, nodeType)
	if nodeSelectorConfiguration != nil {
		sts.Spec.Template.Spec.NodeSelector = nodeSelectorConfiguration
	}

	affinityConfiguration := getNodeAffinityDetails(m.Spec.BrokerNodeAssignment, nodeType)
	if affinityConfiguration != nil {
		sts.Spec.Template.Spec.Affinity = affinityConfiguration
	}

	tolerationConfiguration := getNodeTolerationDetails(m.Spec.BrokerNodeAssignment, nodeType)
	if tolerationConfiguration != nil {
		sts.Spec.Template.Spec.Tolerations = tolerationConfiguration
	}

}

func (r *PubSubPlusEventBrokerReconciler) getBrokerImageDetails(bm *eventbrokerv1beta1.BrokerImage) string {
	imageRepo := bm.Repository
	imageTag := bm.Tag
	if len(strings.TrimSpace(bm.Repository)) == 0 {
		imageRepo = (map[bool]string{true: DefaultBrokerImageRepoOpenShift, false: DefaultBrokerImageRepoK8s})[r.IsOpenShift]
	}
	if len(strings.TrimSpace(bm.Tag)) == 0 {
		imageTag = (map[bool]string{true: DefaultBrokerImageTagOpenShift, false: DefaultBrokerImageTagK8s})[r.IsOpenShift]
	}
	return imageRepo + ":" + imageTag
}

func getTimezone(tz string) string {
	if len(strings.TrimSpace(tz)) == 0 {
		return "UTC"
	}
	return tz
}

func getBrokerMessageNodeStorageSize(st *eventbrokerv1beta1.Storage) string {
	if st == nil || len(strings.TrimSpace(st.MessagingNodeStorageSize)) == 0 || st.MessagingNodeStorageSize == "0" {
		return "30Gi"
	}
	return st.MessagingNodeStorageSize
}

func getNodeAffinityDetails(na []eventbrokerv1beta1.NodeAssignment, nodeType string) *corev1.Affinity {
	for _, nodeAssignment := range na {
		if strings.Contains(
			strings.ToLower(nodeType),
			strings.ToLower(nodeAssignment.Name),
		) {
			if (corev1.Affinity{}) != nodeAssignment.Spec.Affinity {
				return &nodeAssignment.Spec.Affinity
			}
		}
	}
	return nil
}

func getNodeTolerationDetails(na []eventbrokerv1beta1.NodeAssignment, nodeType string) []corev1.Toleration {
	for _, nodeAssignment := range na {
		if strings.Contains(
			strings.ToLower(nodeType),
			strings.ToLower(nodeAssignment.Name),
		) {
			if len(nodeAssignment.Spec.Tolerations) > 0 {
				return nodeAssignment.Spec.Tolerations
			}
		}
	}
	return nil
}

func getNodeSelectorDetails(na []eventbrokerv1beta1.NodeAssignment, nodeType string) map[string]string {
	for _, nodeAssignment := range na {
		if strings.Contains(
			strings.ToLower(nodeType),
			strings.ToLower(nodeAssignment.Name),
		) {
			if len(nodeAssignment.Spec.NodeSelector) > 0 {
				return nodeAssignment.Spec.NodeSelector
			}
		}
	}
	return nil
}

func usesEphemeralStorageForMonitoringNode(st *eventbrokerv1beta1.Storage, nodeType string) bool {
	var useEphemeralStorageForMonitoringNode = false
	if st == nil && nodeType == "monitor" {
		useEphemeralStorageForMonitoringNode = false
	} else if len(strings.TrimSpace(st.MonitorNodeStorageSize)) == 0 && nodeType == "monitor" {
		useEphemeralStorageForMonitoringNode = false
	} else if st.MonitorNodeStorageSize == "0" && nodeType == "monitor" {
		useEphemeralStorageForMonitoringNode = true
	}
	return useEphemeralStorageForMonitoringNode
}

func usesEphemeralStorageForMessageNode(st *eventbrokerv1beta1.Storage, nodeType string) bool {
	var useEphemeralStorageForMessageNode = false
	if st == nil && nodeType != "monitor" {
		useEphemeralStorageForMessageNode = false
	} else if len(strings.TrimSpace(st.MessagingNodeStorageSize)) == 0 && nodeType != "monitor" {
		useEphemeralStorageForMessageNode = false
	} else if st.MessagingNodeStorageSize == "0" && nodeType != "monitor" {
		useEphemeralStorageForMessageNode = true
	}
	return useEphemeralStorageForMessageNode
}
