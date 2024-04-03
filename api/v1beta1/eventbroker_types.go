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

package v1beta1

import (
	"encoding/json"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// EventBrokerSpec defines the desired state of PubSubPlusEventBroker
type EventBrokerSpec struct {
	//+optional
	//+kubebuilder:validation:Type:=boolean
	//+kubebuilder:default:=false
	//+operator-sdk:csv:customresourcedefinitions:type=spec,order=1
	// Redundancy true specifies HA deployment, false specifies Non-HA.
	Redundancy bool `json:"redundancy"`
	//+optional
	//+kubebuilder:validation:Type:=boolean
	//+kubebuilder:default:=false
	//+operator-sdk:csv:customresourcedefinitions:type=spec,order=2
	// Developer true specifies a minimum footprint scaled-down deployment, not for production use.
	// If set to true it overrides SystemScaling parameters.
	Developer bool `json:"developer"`
	//+optional
	//+kubebuilder:validation:Type:=object
	//+kubebuilder:pruning:PreserveUnknownFields
	//+kubebuilder:validation:Schemaless
	//+operator-sdk:csv:customresourcedefinitions:type=spec,order=3
	// SystemScaling provides exact fine-grained specification of the event broker scaling parameters
	// and the assigned CPU / memory resources to the Pod.
	SystemScaling *SystemScaling `json:"systemScaling,omitempty"`
	//+optional
	//+nullable
	//+kubebuilder:validation:Type:=string
	// Defines the password for PubSubPlusEventBroker if provided. Random one will be generated if not provided.
	// When provided, ensure the secret key name is `username_admin_password`. For valid values refer to the Solace documentation https://docs.solace.com/Admin/Configuring-Internal-CLI-User-Accounts.htm.
	AdminCredentialsSecret string `json:"adminCredentialsSecret,omitempty"`
	//+optional
	//+nullable
	//+kubebuilder:validation:Type:=string
	// Defines the password for PubSubPlusEventBroker to be used by the Exporter for monitoring.
	// When provided, ensure the secret key name is `username_monitor_password`. For valid values refer to the Solace documentation https://docs.solace.com/Admin/Configuring-Internal-CLI-User-Accounts.htm.
	MonitoringCredentialsSecret string `json:"monitoringCredentialsSecret,omitempty"`
	//+optional
	//+nullable
	//+kubebuilder:validation:Type:=string
	// PreSharedAuthKeySecret defines the PreSharedAuthKey Secret for PubSubPlusEventBroker. Random one will be generated if not provided.
	// When provided, ensure the secret key name is `preshared_auth_key`. For valid values refer to the Solace documentation https://docs.solace.com/Features/HA-Redundancy/Pre-Shared-Keys-SMB.htm?Highlight=pre%20shared.
	PreSharedAuthKeySecret string `json:"preSharedAuthKeySecret,omitempty"`
	//+optional
	//+kubebuilder:validation:Type:=array
	// List of extra environment variables to be added to the PubSubPlusEventBroker container. Note: Do not configure Timezone or SystemScaling parameters here as it could cause unintended consequences.
	// A primary use case is to specify configuration keys, although the variables defined here will not override the ones defined in ConfigMap
	ExtraEnvVars []*ExtraEnvVar `json:"extraEnvVars"`
	//+optional
	//+kubebuilder:validation:Type:=string
	// List of extra environment variables to be added to the PubSubPlusEventBroker container from an existing ConfigMap. Note: Do not configure Timezone or SystemScaling parameters here as it could cause unintended consequences.
	ExtraEnvVarsCM string `json:"extraEnvVarsCM,omitempty"`
	//+optional
	//+kubebuilder:validation:Type:=string
	// List of extra environment variables to be added to the PubSubPlusEventBroker container from an existing Secret
	ExtraEnvVarsSecret string `json:"extraEnvVarsSecret,omitempty"`
	//+optional
	//+kubebuilder:validation:Type:=boolean
	//+kubebuilder:default:=false
	// PodDisruptionBudgetForHA enables setting up PodDisruptionBudget for the broker pods in HA deployment.
	// This parameter is ignored for non-HA deployments (if redundancy is false).
	PodDisruptionBudgetForHA bool `json:"podDisruptionBudgetForHA"`
	//+optional
	//+kubebuilder:validation:Enum=automatedRolling;manualPodRestart
	//+kubebuilder:default:=automatedRolling
	// UpdateStrategy specifies how to update an existing deployment. manualPodRestart waits for user intervention.
	UpdateStrategy PubSubPlusEventBrokerUpdateStrategy `json:"updateStrategy"`
	//+optional
	//+kubebuilder:validation:Type:=string
	//+kubebuilder:default:="UTC"
	// Defines the timezone for the event broker container, if undefined default is UTC. Valid values are tz database time zone names.
	Timezone string `json:"timezone"`
	//+optional
	//+kubebuilder:validation:Type:=object
	//+kubebuilder:default:={}
	// PodLabels allows adding provider-specific pod labels to PubSubPlusEventBroker pods
	PodLabels map[string]string `json:"podLabels"`
	//+optional
	//+kubebuilder:validation:Type:=object
	//+kubebuilder:default:={}
	// PodAnnotations allows adding provider-specific pod annotations to PubSubPlusEventBroker pods
	PodAnnotations map[string]string `json:"podAnnotations"`
	//+kubebuilder:validation:Type:=object
	// Image defines container image parameters for the event broker.
	BrokerImage BrokerImage `json:"image,omitempty"`
	//+kubebuilder:validation:Type:=array
	// NodeAssignment defines labels to constrain PubSubPlusEventBroker nodes to run on particular node(s), or to prefer to run on particular nodes.
	BrokerNodeAssignment []NodeAssignment `json:"nodeAssignment,omitempty"`
	//+kubebuilder:validation:Type:=object
	// SecurityContext defines the pod security context for the event broker.
	SecurityContext SecurityContext `json:"securityContext,omitempty"`
	//+kubebuilder:validation:Type:=object
	// ServiceAccount defines a ServiceAccount dedicated to the PubSubPlusEventBroker
	ServiceAccount BrokerServiceAccount `json:"serviceAccount,omitempty"`
	//+kubebuilder:validation:Type:=object
	// TLS provides TLS configuration for the event broker.
	BrokerTLS BrokerTLS `json:"tls,omitempty"`
	//+kubebuilder:validation:Type:=object
	// Service defines broker service details.
	Service Service `json:"service,omitempty"`
	//+kubebuilder:validation:Type:=object
	// Storage defines storage details for the broker.
	Storage Storage `json:"storage,omitempty"`
	//+kubebuilder:validation:Type:=object
	// Monitoring specifies a Prometheus monitoring endpoint for the event broker
	Monitoring Monitoring `json:"monitoring,omitempty"`
}

type PubSubPlusEventBrokerUpdateStrategy string

const (
	AutomatedRollingUpdateStrategy PubSubPlusEventBrokerUpdateStrategy = "automatedRolling"
	ManualPodRestartUpdateStrategy PubSubPlusEventBrokerUpdateStrategy = "manualPodRestart"
)

// Port defines parameters configure Service details for the Broker
type BrokerPort struct {
	//+kubebuilder:validation:MaxLength:15
	//+kubebuilder:validation:Type:=string
	// Unique name for the port that can be referred to by services.
	Name string `json:"name"`
	//+kubebuilder:validation:Enum=TCP;UDP;SCTP
	// Protocol for port. Must be UDP, TCP, or SCTP.
	Protocol corev1.Protocol `json:"protocol"`
	//+kubebuilder:validation:Type:=number
	// Port number to expose on the pod.
	ContainerPort int32 `json:"containerPort"`
	//+kubebuilder:validation:Type:=number
	// Port number to expose on the service
	ServicePort int32 `json:"servicePort"`
}

// Service defines parameters configure Service details for the Broker
type Service struct {
	//+optional
	//+kubebuilder:validation:Type:=string
	//+kubebuilder:default:=LoadBalancer
	// ServiceType specifies how to expose the broker services. Options include ClusterIP, NodePort, LoadBalancer (default).
	ServiceType corev1.ServiceType `json:"type"`
	//+optional
	//+kubebuilder:validation:Type:=object
	//+kubebuilder:default:={}
	// Annotations allows adding provider-specific service annotations
	Annotations map[string]string `json:"annotations"`
	//+optional
	//+kubebuilder:validation:Type:=array
	//+kubebuilder:default:={{name:"tcp-ssh",protocol:"TCP",containerPort:2222,servicePort:2222},{name:"tcp-semp",protocol:"TCP",containerPort:8080,servicePort:8080},{name:"tls-semp",protocol:"TCP",containerPort:1943,servicePort:1943},{name:"tcp-smf",protocol:"TCP",containerPort:55555,servicePort:55555},{name:"tcp-smfcomp",protocol:"TCP",containerPort:55003,servicePort:55003},{name:"tls-smf",protocol:"TCP",containerPort:55443,servicePort:55443},{name:"tcp-smfroute",protocol:"TCP",containerPort:55556,servicePort:55556},{name:"tcp-web",protocol:"TCP",containerPort:8008,servicePort:8008},{name:"tls-web",protocol:"TCP",containerPort:1443,servicePort:1443},{name:"tcp-rest",protocol:"TCP",containerPort:9000,servicePort:9000},{name:"tls-rest",protocol:"TCP",containerPort:9443,servicePort:9443},{name:"tcp-amqp",protocol:"TCP",containerPort:5672,servicePort:5672},{name:"tls-amqp",protocol:"TCP",containerPort:5671,servicePort:5671},{name:"tcp-mqtt",protocol:"TCP",containerPort:1883,servicePort:1883},{name:"tls-mqtt",protocol:"TCP",containerPort:8883,servicePort:8883},{name:"tcp-mqttweb",protocol:"TCP",containerPort:8000,servicePort:8000},{name:"tls-mqttweb",protocol:"TCP",containerPort:8443,servicePort:8443}}
	// Ports specifies the ports to expose PubSubPlusEventBroker services.
	Ports []*BrokerPort `json:"ports"`
}

// Storage defines parameters configure Storage details for the Broker
type Storage struct {
	//+optional
	//+kubebuilder:validation:Type:=boolean
	//+kubebuilder:default:=false
	// Slow indicate slow storage is in use, an example is NFS.
	Slow bool `json:"slow,omitempty"`
	//+kubebuilder:validation:Type:=string
	//+kubebuilder:default:="30Gi"
	// MessagingNodeStorageSize if provided will assign the minimum persistent storage to be used by the message nodes.
	MessagingNodeStorageSize string `json:"messagingNodeStorageSize,omitempty"`
	//+optional
	//+kubebuilder:validation:Type:=string
	//+kubebuilder:default:="3Gi"
	// MonitorNodeStorageSize if provided this will create and assign the minimum recommended storage to Monitor pods.
	MonitorNodeStorageSize string `json:"monitorNodeStorageSize,omitempty"`
	//+optional
	//+kubebuilder:validation:Type:=array
	//CustomVolumeMount can be used to show the data volume should be mounted instead of using a storage class.
	CustomVolumeMount []StorageCustomVolumeMount `json:"customVolumeMount,omitempty"`
	//+optional
	//+kubebuilder:validation:Type:=string
	// UseStrorageClass Name of the StorageClass to be used to request persistent storage volumes. If undefined, the "default" StorageClass will be used.
	UseStorageClass string `json:"useStorageClass,omitempty"`
}

// StorageCustomVolumeMount defines Image details and pulling configurations
type StorageCustomVolumeMount struct {
	//+kubebuilder:validation:Type:=string
	//+kubebuilder:validation:Enum=Primary;Backup;Monitor
	// Defines the name of PubSubPlusEventBroker node type that has the customVolumeMount spec defined
	Name string `json:"name,omitempty"`
	//+optional
	//+kubebuilder:validation:Type:=object
	// Defines the customVolumeMount that can be used mount the data volume instead of using a storage class
	PersistentVolumeClaim BrokerPersistentVolumeClaim `json:"persistentVolumeClaim,omitempty"`
}

// BrokerPersistentVolumeClaim defines custom PersistentVolumeClaim to be use by PubSubPlusEventBroker
type BrokerPersistentVolumeClaim struct {
	//+kubebuilder:validation:Type:=string
	// Defines the claimName of a custom PersistentVolumeClaim to be used instead
	ClaimName string `json:"claimName"`
}

// +kubebuilder:pruning:PreserveUnknownFields
type SystemScaling struct {
	// +kubebuilder:default:=100
	MaxConnections int `json:"maxConnections,omitempty"`
	// +kubebuilder:default:=100
	MaxQueueMessages int `json:"maxQueueMessages,omitempty"`
	// +kubebuilder:default:=1000
	MaxSpoolUsage int `json:"maxSpoolUsage,omitempty"`
	// +kubebuilder:default:="2"
	MessagingNodeCpu string `json:"messagingNodeCpu,omitempty"`
	// +kubebuilder:default:="4025Mi"
	MessagingNodeMemory string `json:"messagingNodeMemory,omitempty"`
	//+kubebuilder:pruning:PreserveUnknownFields
	runtime.RawExtension `json:"-"`
}

// BrokerTLS defines TLS configuration for the PubSubPlusEventBroker
type BrokerTLS struct {
	//+optional
	//+kubebuilder:validation:Type:=boolean
	//+kubebuilder:default:=false
	// Enabled true enables TLS for the broker.
	Enabled bool `json:"enabled"`
	//+optional
	//+kubebuilder:validation:Type:=string
	//+kubebuilder:default:=example-tls-secret
	// Specifies the tls configuration secret to be used for the broker
	ServerTLsConfigSecret string `json:"serverTlsConfigSecret"`
	//+optional
	//+kubebuilder:validation:Type:=string
	//+kubebuilder:default:=tls.key
	// Name of the Certificate file in the `serverCertificatesSecret`
	TLSCertName string `json:"certFilename"`
	//+optional
	//+kubebuilder:validation:Type:=string
	//+kubebuilder:default:=tls.crt
	// Name of the Key file in the `serverCertificatesSecret`
	TLSCertKeyName string `json:"certKeyFilename"`
}

// ServiceAccount defines a ServiceAccount dedicated to the PubSubPlusEventBroker
type BrokerServiceAccount struct {
	//+kubebuilder:validation:Type:=string
	// Name specifies the name of an existing ServiceAccount dedicated to the PubSubPlusEventBroker.
	// If this value is missing a new ServiceAccount will be created.
	Name string `json:"name"`
}

// ExtraEnvVar defines environment variables to be added to the PubSubPlusEventBroker container
type ExtraEnvVar struct {
	//+kubebuilder:validation:Type:=string
	// Specifies the Name of an environment variable to be added to the PubSubPlusEventBroker container
	Name string `json:"name"`
	//+kubebuilder:validation:Type:=string
	// Specifies the Value of an environment variable to be added to the PubSubPlusEventBroker container
	Value string `json:"value"`
}

// BrokerImage defines Image details and pulling configurations
type BrokerImage struct {
	//+optional
	//+kubebuilder:validation:Type:=string
	// Defines the container image repo where the event broker image is pulled from
	Repository string `json:"repository"`
	//+optional
	//+kubebuilder:validation:Type:=string
	//+kubebuilder:default:="latest"
	// Specifies the tag of the container image to be used for the event broker.
	Tag string `json:"tag"`
	//+optional
	//+kubebuilder:validation:Type:=string
	//+kubebuilder:default:="IfNotPresent"
	// Specifies ImagePullPolicy of the container image for the event broker.
	ImagePullPolicy corev1.PullPolicy `json:"pullPolicy"`
	//+optional
	//+kubebuilder:validation:Type:=array
	// pullSecrets is an optional list of references to secrets in the same namespace to use for pulling any of the images used by this PodSpec.
	ImagePullSecrets []corev1.LocalObjectReference `json:"pullSecrets,omitempty"`
}

// NodeAssignment defines labels to constrain PubSubPlusEventBroker nodes to specific nodes
type NodeAssignment struct {
	//+kubebuilder:validation:Type:=string
	//+kubebuilder:validation:Enum=Primary;Backup;Monitor
	// Defines the name of broker node type that has the nodeAssignment spec defined
	Name string `json:"name"`
	//+kubebuilder:validation:Type:=object
	// If provided defines the labels to constrain the PubSubPlusEventBroker node to specific nodes
	Spec NodeAssignmentSpec `json:"spec"`
}

// NodeAssignmentSpec defines the NodeAffinity, NodeSelector, and Tolerations details to be used for event broker nodes
type NodeAssignmentSpec struct {
	//+optional
	//+nullable
	//+kubebuilder:validation:Type:=object
	//+kubebuilder:default:={}
	// Affinity if provided defines the conditional approach to assign PubSubPlusEventBroker nodes to specific nodes to which they can be scheduled
	Affinity corev1.Affinity `json:"affinity,omitempty"`
	//+optional
	//+kubebuilder:validation:Type:=object
	//+kubebuilder:default:={}
	// NodeSelector if provided defines the exact labels of nodes to which PubSubPlusEventBroker nodes can be scheduled
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	//+optional
	//+kubebuilder:validation:Type:=array
	// Toleration if provided defines the exact properties of the PubSubPlusEventBroker nodes can be scheduled on nodes with d matching taint.
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`
}

// SecurityContext defines the pod security context for the PubSubPlusEventBroker
type SecurityContext struct {
	//+optional
	//+kubebuilder:validation:Type:=number
	// Specifies fsGroup in pod security context. 0 or unset defaults either to 1000002, or if OpenShift detected to unspecified (see documentation)
	FSGroup int64 `json:"fsGroup"`
	//+optional
	//+kubebuilder:validation:Type:=number
	// Specifies runAsUser in pod security context. 0 or unset defaults either to 1000001, or if OpenShift detected to unspecified (see documentation)
	RunAsUser int64 `json:"runAsUser"`
}

// MonitoringImage defines Image details and pulling configurations for the Prometheus Exporter for Monitoring
type MonitoringImage struct {
	//+kubebuilder:validation:Type:=string
	// Defines the container image repo where the Prometheus Exporter image is pulled from
	Repository string `json:"repository,omitempty"`
	//+kubebuilder:validation:Type:=string
	//+kubebuilder:default:=latest
	// Specifies the tag of the container image to be used for the Prometheus Exporter.
	Tag string `json:"tag,omitempty"`
	//+kubebuilder:validation:Type:=string
	//+kubebuilder:default:=IfNotPresent
	// Specifies ImagePullPolicy of the container image for the Prometheus Exporter.
	ImagePullPolicy corev1.PullPolicy `json:"pullPolicy,omitempty"`
	// pullSecrets is an optional list of references to secrets in the same namespace to use for pulling any of the images used by this PodSpec.
	// +optional
	//+kubebuilder:validation:Type:=array
	ImagePullSecrets []corev1.LocalObjectReference `json:"pullSecrets,omitempty"`
}

// Monitoring defines parameters to use Prometheus Exporter
type Monitoring struct {
	//+optional
	//+kubebuilder:validation:Type:=boolean
	//+kubebuilder:default:=false
	// Enabled true enables the setup of the Prometheus Exporter.
	Enabled bool `json:"enabled"`
	//+optional
	//+kubebuilder:validation:Type:=object
	// Image defines container image parameters for the Prometheus Exporter.
	MonitoringImage *MonitoringImage `json:"image,omitempty"`
	//+optional
	//+kubebuilder:validation:Type:=object
	// MetricsEndpoint defines parameters to configure monitoring for the Prometheus Exporter.
	MonitoringMetricsEndpoint *MonitoringMetricsEndpoint `json:"metricsEndpoint,omitempty"`
	//+optional
	//+kubebuilder:validation:Type:=number
	//+kubebuilder:default:=5
	// Timeout configuration for Prometheus Exporter scrapper
	TimeOut int32 `json:"timeOut,omitempty"`
	//+optional
	//+kubebuilder:validation:Type:=boolean
	//+kubebuilder:default:=false
	// Defines if Prometheus Exporter verifies SSL
	SSLVerify bool `json:"sslVerify,omitempty"`
	//+optional
	//+kubebuilder:validation:Type:=boolean
	//+kubebuilder:default:=false
	// Defines if Prometheus Exporter should include rates
	IncludeRates bool `json:"includeRates,omitempty"`
}

// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
// Important: Run "make" to regenerate code after modifying this file

// EventBrokerStatus defines the observed state of the PubSubPlusEventBroker
type EventBrokerStatus struct {
	// PodsList are the names of the eventbroker and optionally the monitoring pods
	PodsList []string `json:"podsList,omitempty"`
	// Conditions provide information about the observed status of the deployment
	Conditions []metav1.Condition `json:"conditions,omitempty"`
	// Broker section provides the broker status
	Broker BrokerSubStatus `json:"broker,omitempty"`
	// Monitoring sectionprovides monitoring support status
	Monitoring MonitoringSubStatus `json:"prometheusMonitoring,omitempty"`
}

type BrokerSubStatus struct {
	HADeployment           string   `json:"haDeployment,omitempty"`
	TLSSupport             string   `json:"tlsSupport,omitempty"`
	TLSSecret              string   `json:"tlsSecret,omitempty"`
	AdminCredentialsSecret string   `json:"adminCredentialsSecret,omitempty"`
	ServiceName            string   `json:"serviceName,omitempty"`
	ServiceType            string   `json:"serviceType,omitempty"`
	StatefulSets           []string `json:"statefulSets,omitempty"`
	BrokerImage            string   `json:"brokerImage,omitempty"`
}

type MonitoringSubStatus struct {
	Enabled       string `json:"enabled,omitempty"`
	ServiceName   string `json:"serviceName,omitempty"`
	ExporterImage string `json:"exporterImage,omitempty"`
}

// MonitoringMetricsEndpoint defines parameters to configure Metrics Service Endpoint
type MonitoringMetricsEndpoint struct {
	//+optional
	//+kubebuilder:validation:MaxLength:15
	//+kubebuilder:validation:Type:=string
	// Name is a unique name for the port that can be referred to by services.
	Name string `json:"name"`
	//+optional
	//+kubebuilder:validation:Type:=number
	//+kubebuilder:default:=9628
	// ContainerPort is the port number to expose on the Prometheus Exporter pod.
	ContainerPort int32 `json:"containerPort"`
	//+optional
	//+kubebuilder:validation:Type:=number
	//+kubebuilder:default:=9628
	// ServicePort is the port number to expose on the service
	ServicePort int32 `json:"servicePort"`
	//+optional
	//+kubebuilder:validation:Type:=boolean
	//+kubebuilder:default:=false
	// Defines if Metrics Service Endpoint uses TLS configuration
	ListenTLS bool `json:"listenTLS"`
	//+optional
	//+kubebuilder:validation:Enum=TCP;UDP;SCTP
	// Protocol for port. Must be UDP, TCP, or SCTP.
	Protocol corev1.Protocol `json:"protocol"`
	//+optional
	//+kubebuilder:validation:Type:=string
	// EndpointTLSConfigSecret defines TLS secret name to set up TLS configuration
	EndpointTLSConfigSecret string `json:"endpointTlsConfigSecret,omitempty"`
	//+optional
	//+kubebuilder:validation:Type:=string
	//+kubebuilder:default:=tls.crt
	// EndpointTlsConfigServerCertName is the file name of the Server Certificate used to set up TLS configuration
	EndpointTlsConfigServerCertName string `json:"endpointTlsConfigServerCertName,omitempty"`
	//+optional
	//+kubebuilder:validation:Type:=string
	//+kubebuilder:default:=tls.key
	// EndpointTlsConfigPrivateKeyName is the file name of the Private Key used to set up TLS configuration
	EndpointTlsConfigPrivateKeyName string `json:"endpointTlsConfigPrivateKeyName,omitempty"`
	//+optional
	//+kubebuilder:validation:Type:=string
	//+kubebuilder:default:=ClusterIP
	// Defines the service type for the Metrics Service Endpoint
	ServiceType corev1.ServiceType `json:"serviceType,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=pubsubpluseventbrokers,shortName=eb;eventbroker
// +operator-sdk:csv:customresourcedefinitions:displayName="PubSub+ Event Broker",resources={{StatefulSet,v1},{Pod,v1},{Service,v1},{Secret,v1},{ConfigMap,v1},{Deployment,v1}}
// PubSub+ Event Broker
type PubSubPlusEventBroker struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   EventBrokerSpec   `json:"spec,omitempty"`
	Status EventBrokerStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// PubSubPlusEventBrokerList contains a list of PubSubPlusEventBroker
type PubSubPlusEventBrokerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PubSubPlusEventBroker `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PubSubPlusEventBroker{}, &PubSubPlusEventBrokerList{})
}

func (spec EventBrokerSpec) String() string {
	specString, _ := json.Marshal(spec)
	return string(specString)
}

func (service Service) String() string {
	serviceString, _ := json.Marshal(service)
	return string(serviceString)
}

func (monitoring Monitoring) String() string {
	monitoringString, _ := json.Marshal(monitoring)
	return string(monitoringString)
}
