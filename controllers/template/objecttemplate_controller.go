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
	"reflect"

	"github.com/go-logr/logr"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	otv1 "github.com/ericogr/k8s-object-template/apis/v1"
)

// ObjectTemplateReconciler ot reconciler
type ObjectTemplateReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// SetupWithManager setup
func (r *ObjectTemplateReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&otv1.ObjectTemplate{}).
		WithEventFilter(predicate.GenerationChangedPredicate{}).
		Complete(r)
}

// +kubebuilder:rbac:groups=template.k8s.ericogr.com.br,resources=objecttemplates,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=template.k8s.ericogr.com.br,resources=objecttemplates/status,verbs=get;update;patch

// Reconcile k8s reconcile
func (r *ObjectTemplateReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("objecttemplate", otGV)
	var objectTemplate otv1.ObjectTemplate
	err := r.Get(ctx, req.NamespacedName, &objectTemplate)
	common := Common{r.Client, log}

	if err != nil {
		objectTemplate.Status.Status = err.Error()

		if k8sErrors.IsNotFound(err) {
			// Object not found, return. Created objects are automatically garbage collected
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}

		// Error reading the object - requeue the request.
		return ctrl.Result{}, err
	}

	defer common.UpdateStatus(ctx, &objectTemplate)

	otParams, err := common.FindObjectTemplateParamsByTemplateName(objectTemplate.Name)

	if err != nil {
		objectTemplate.Status.Status = err.Error()
		return ctrl.Result{}, err
	}

	lu := LogUtil{Log: log}
	for _, otParam := range otParams {
		paramNamespace := otParam.Namespace
		paramValues, err := otParam.Spec.GetParametersByTemplateName(objectTemplate.Name)

		if err != nil {
			lu.Error(err, "Error getting parameters by template name")
			continue
		}

		kind := reflect.TypeOf(otv1.ObjectTemplateParams{}).Name()
		gvk := otv1.GroupVersion.WithKind(kind)
		controllerRef := metav1.NewControllerRef(otParam.GetObjectMeta(), gvk)

		if err := common.UpdateObjectsByTemplate(objectTemplate, []metav1.OwnerReference{*controllerRef}, paramNamespace, paramValues.Values); err != nil {
			lu.Error(err, "Failed to update ObjectTemplate")
			continue
		}
	}

	objectTemplate.Status.Status = "OK"
	if lu.HasError() {
		objectTemplate.Status.Status = lu.AllErrorsMessages()
	}

	return ctrl.Result{Requeue: false}, nil
}
