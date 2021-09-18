/*
Copyright 2021.

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
	"fmt"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// PodReconciler reconciles a Pod object
type PodReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=labeler.uthark.dev,resources=pods,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=labeler.uthark.dev,resources=pods/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=labeler.uthark.dev,resources=pods/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Pod object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *PodReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	pod := &v1.Pod{}
	err := r.Get(ctx, req.NamespacedName, pod)
	if err != nil {
		fmt.Println(err)
		return ctrl.Result{}, err
	}

	namespace := &v1.Namespace{}
	err = r.Get(ctx, client.ObjectKey{Name: req.Namespace}, namespace)
	if err != nil {
		fmt.Println(err)
		return ctrl.Result{}, err
	}

	if namespace.Annotations["labeler.uthark.dev/enabled"] != "true" {
		fmt.Println("Skipping namespace", namespace)
		return ctrl.Result{}, nil
	}

	_, err = ctrl.CreateOrUpdate(ctx, r.Client, pod, func() error {
		for _, label := range []string{"team", "product", "project"} {
			r.copyLabel(pod, namespace, label)
		}
		fmt.Println("Processed", req.NamespacedName)
		return nil
	})

	return ctrl.Result{}, err
}

// SetupWithManager sets up the controller with the Manager.
func (r *PodReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		// Uncomment the following line adding a pointer to an instance of the controlled resource as an argument
		For(&v1.Pod{}).
		Complete(r)
}

func (r *PodReconciler) copyLabel(pod *v1.Pod, ns *v1.Namespace, labelKey string) {
	if ns.Labels[labelKey] == "" {
		return
	}

	//  TODO: Add feature to keep label if it was present originally.
	// Track which labels were added by labeler.
	if pod.Labels[labelKey] == "" || pod.Labels[labelKey] != ns.Labels[labelKey] {
		metav1.SetMetaDataLabel(&pod.ObjectMeta, labelKey, ns.Labels[labelKey])
	}
}
