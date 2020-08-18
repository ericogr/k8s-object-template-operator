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

	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/go-logr/logr"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	toolsaocv1 "github.com/ericogr/k8s-aoc/apis/template.totvs.app/v1"
)

var (
	aocGV = toolsaocv1.GroupVersion.String()
)

// ObjectTemplateReconciler aoc reconciler
type ObjectTemplateReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// SetupWithManager setup
func (r *ObjectTemplateReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&toolsaocv1.ObjectTemplate{}).
		Complete(r)
}

// +kubebuilder:rbac:groups=template.totvs.app.github.com,resources=objecttemplates,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=template.totvs.app.github.com,resources=objecttemplates/status,verbs=get;update;patch

// Reconcile k8s reconcile
func (r *ObjectTemplateReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	// if 1 > 0 {
	// 	return ctrl.Result{}, nil
	// }

	ctx := context.Background()
	log := r.Log.WithValues("objecttemplate", aocGV)
	var aoc toolsaocv1.ObjectTemplate
	err := r.Get(ctx, req.NamespacedName, &aoc)

	if err != nil {
		if k8sErrors.IsNotFound(err) {
			// Object not found, return. Created objects are automatically garbage collected
			return ctrl.Result{}, nil
		}

		// Error reading the object - requeue the request.
		return ctrl.Result{}, err
	}

	common := Common{r.Client, r.Log}
	aocParams, err := common.FindObjectTemplateParamsByTemplateName(aoc.Spec.Template.Name)
	listErrors := ""
	if err != nil {
		listErrors = err.Error()
	} else {
		for _, aocParam := range aocParams {
			paramNamespace := aocParam.Namespace
			paramValues, err := aocParam.Spec.GetParametersByTemplateName(aoc.Spec.Template.Name)

			if err != nil {
				listErrors += err.Error() + "\n"
				continue
			}

			if err := common.UpdateObjectByNamespace(aoc, paramNamespace, paramValues.Values); err != nil {
				listErrors += err.Error() + "\n"
				continue
			}
		}
	}

	// //https://godoc.org/sigs.k8s.io/controller-runtime/pkg/predicate#GenerationChangedPredicate
	if listErrors != "" {
		aoc.Status.Status = listErrors
	} else {
		aoc.Status.Status = "OK"
	}

	if err := r.Status().Update(ctx, &aoc); err != nil {
		log.Error(err, "Unable to update status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}
