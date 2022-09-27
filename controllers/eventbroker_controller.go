/*
Copyright 2022.

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
	"reflect"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	"context"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"

	eventbrokerv1alpha1 "github.com/SolaceProducts/pubsubplus-operator/api/v1alpha1"
)

// EventBrokerReconciler reconciles a EventBroker object
type EventBrokerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=pubsubplus.solace.com,resources=eventbrokers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=pubsubplus.solace.com,resources=eventbrokers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=pubsubplus.solace.com,resources=eventbrokers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the EventBroker object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *EventBrokerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrllog.FromContext(ctx)

	// Fetch the EventBroker instance
	eventbroker := &eventbrokerv1alpha1.EventBroker{}
	err := r.Get(ctx, req.NamespacedName, eventbroker)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("EventBroker resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get EventBroker")
		return ctrl.Result{}, err
	} else {
		log.Info("Detected existing eventbroker", " eventbroker.Name", eventbroker.Name)
	}

	// Check if the Service already exists, if not create a new one
	svc := &corev1.Service{}
	err = r.Get(ctx, types.NamespacedName{Name: eventbroker.Name + "-pubsubplus", Namespace: eventbroker.Namespace}, svc)
	if err != nil && errors.IsNotFound(err) {
		// Define a new service
		svc := r.serviceForEventBroker(eventbroker)
		log.Info("Creating a new Service", "Service.Namespace", svc.Namespace, "Service.Name", svc.Name)
		err = r.Create(ctx, svc)
		if err != nil {
			log.Error(err, "Failed to create new Service", "Service.Namespace", svc.Namespace, "Service.Name", svc.Name)
			return ctrl.Result{}, err
		}
		// Service created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Service")
		return ctrl.Result{}, err
	} else {
		log.Info("Detected existing Service", " Service.Name", svc.Name)
	}

	// Check if the ConfigMap already exists, if not create a new one
	cm := &corev1.ConfigMap{}
	err = r.Get(ctx, types.NamespacedName{Name: eventbroker.Name + "-pubsubplus", Namespace: eventbroker.Namespace}, cm)
	if err != nil && errors.IsNotFound(err) {
		// Define a new configmap
		cm := r.configmapForEventBroker(eventbroker)
		log.Info("Creating a new ConfigMap", "Configmap.Namespace", cm.Namespace, "Configmap.Name", cm.Name)
		err = r.Create(ctx, cm)
		if err != nil {
			log.Error(err, "Failed to create new ConfigMap", "Configmap.Namespace", cm.Namespace, "Configmap.Name", cm.Name)
			return ctrl.Result{}, err
		}
		// ConfigMap created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get ConfigMap")
		return ctrl.Result{}, err
	} else {
		log.Info("Detected existing ConfigMap", " ConfigMap.Name", cm.Name)
	}

	// Check if the StatefulSet already exists, if not create a new one
	sts := &appsv1.StatefulSet{}
	err = r.Get(ctx, types.NamespacedName{Name: eventbroker.Name + "-pubsubplus-p", Namespace: eventbroker.Namespace}, sts)
	if err != nil && errors.IsNotFound(err) {
		// Define a new statefulset
		sts := r.statefulsetForEventBroker(eventbroker, "-p")
		log.Info("Creating a new StatefulSet", "StatefulSet.Namespace", sts.Namespace, "StatefulSet.Name", sts.Name)
		err = r.Create(ctx, sts)
		if err != nil {
			log.Error(err, "Failed to create new StatefulSet", "StatefulSet.Namespace", sts.Namespace, "StatefulSet.Name", sts.Name)
			return ctrl.Result{}, err
		}
		// StatefulSet created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get StatefulSet")
		return ctrl.Result{}, err
	} else {
		log.Info("Detected existing StatefulSet", " StatefulSet.Name", sts.Name)
	}

	// // Ensure the StatefulSet size is the same as the spec
	// size := eventbroker.Spec.Size
	// if *sts.Spec.Replicas != size {
	// 	sts.Spec.Replicas = &size
	// 	log.Info("Detected size change, new size:", "Size:", size)
	// 	err = r.Update(ctx, sts)
	// 	if err != nil {
	// 		log.Error(err, "Failed to update StatefulSet", "StatefulSet.Namespace", sts.Namespace, "StatefulSet.Name", sts.Name)
	// 		return ctrl.Result{}, err
	// 	}

	// 	// Ask to requeue after 1 minute in order to give enough time for the
	// 	// pods be created on the cluster side and the operand be able
	// 	// to do the next update step accurately.
	// 	return ctrl.Result{RequeueAfter: time.Minute}, nil
	// }

	// Update the EventBroker status with the pod names
	// List the pods for this eventbroker's StatefulSet
	podList := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(eventbroker.Namespace),
		client.MatchingLabels(baseLabels(eventbroker.Name)),
	}
	if err = r.List(ctx, podList, listOpts...); err != nil {
		log.Error(err, "Failed to list pods", "EventBroker.Namespace", eventbroker.Namespace, "EventBroker.Name", eventbroker.Name)
		return ctrl.Result{}, err
	}
	podNames := getPodNames(podList.Items)
	// Update status.BrokerPods if needed
	if !reflect.DeepEqual(podNames, eventbroker.Status.BrokerPods) {
		eventbroker.Status.BrokerPods = podNames
		err := r.Status().Update(ctx, eventbroker)
		if err != nil {
			log.Error(err, "Failed to update EventBroker status")
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}

// baseLabels returns the labels for selecting the resources
// belonging to the given eventbroker CR name.
func baseLabels(name string) map[string]string {
	return map[string]string{
		"app.kubernetes.io/instance": name,
		"app.kubernetes.io/name":     "eventbroker",
	}
}

// getPodNames returns the pod names of the array of pods passed in
func getPodNames(pods []corev1.Pod) []string {
	var podNames []string
	for _, pod := range pods {
		podNames = append(podNames, pod.Name)
	}
	return podNames
}

// SetupWithManager sets up the controller with the Manager.
func (r *EventBrokerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&eventbrokerv1alpha1.EventBroker{}).
		Owns(&appsv1.StatefulSet{}).
		Complete(r)
}
