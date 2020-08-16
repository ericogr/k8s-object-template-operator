/*


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

	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"context"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	namespaceGV = corev1.Namespace{}.APIVersion
)

// NamespaceReconciler reconciles a namespace object
type NamespaceReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// SetupWithManager setup
func (r *NamespaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Namespace{}).
		Complete(r)
}

// +kubebuilder:rbac:groups=core,resources=namespace,verbs=get;list
// +kubebuilder:rbac:groups=core,resources=namespace/status,verbs=get

// Reconcile namespace
func (r *NamespaceReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("namespace", namespaceGV)
	var namespace corev1.Namespace
	err := r.Get(ctx, req.NamespacedName, &namespace)

	if err != nil {
		if k8sErrors.IsNotFound(err) {
			// Object not found, return. Created objects are automatically garbage collected
			return ctrl.Result{}, nil
		}

		// Error reading the object - requeue the request.
		return ctrl.Result{}, err
	}

	common := Common{r.Client, r.Log}
	aocs, err := common.FindAOCs()
	for _, aoc := range aocs {
		if common.ValidateNamespace(namespace, aoc.Spec.Trigger.Annotations) {
			if err := common.UpdateObjectByNamespace(aoc, namespace); err != nil {
				aoc.Status.Status = err.Error()

				if err := r.Status().Update(ctx, &aoc); err != nil {
					log.Error(err, "Unable to update status")
					return ctrl.Result{}, err
				}
			}
		}
	}

	return ctrl.Result{}, nil
}
