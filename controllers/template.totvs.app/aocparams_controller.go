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
	"context"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	toolsaocv1 "github.com/ericogr/k8s-aoc/apis/template.totvs.app/v1"
)

var (
	namespaceGV = corev1.Namespace{}.APIVersion
)

// AOCParamsReconciler reconciles a AOCParams object
type AOCParamsReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=template.totvs.app.github.com,resources=aocparams,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=template.totvs.app.github.com,resources=aocparams/status,verbs=get;update;patch

// Reconcile reconcile
func (r *AOCParamsReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("aocparams", req.NamespacedName)
	var params toolsaocv1.AOCParams

	err := r.Get(ctx, req.NamespacedName, &params)

	if err != nil {
		if k8sErrors.IsNotFound(err) {
			// Object not found, return. Created objects are automatically garbage collected
			return ctrl.Result{}, nil
		}

		// Error reading the object - requeue the request.
		return ctrl.Result{}, err
	}

	common := Common{r.Client, r.Log}

	for _, parameter := range params.Spec.Templates {
		aoc, err := common.GetAOCByName(parameter.Name)

		if err != nil {
			log.Error(err, "Failed to get aoc template")
			return ctrl.Result{}, err
		}

		if aoc != nil {
			err = common.UpdateObjectByNamespace(*aoc, req.NamespacedName.Namespace, parameter.Values)

			if err != nil {
				log.Error(err, "Failed to update aoc template")
				return ctrl.Result{}, err
			}
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager setup
func (r *AOCParamsReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&toolsaocv1.AOCParams{}).
		Complete(r)
}
