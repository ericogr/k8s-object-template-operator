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
	"errors"

	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	toolsaocv1 "github.com/ericogr/k8s-aoc/apis/tools/v1"
	"github.com/ericogr/k8s-aoc/pkg/processor"
)

var (
	aocGV = toolsaocv1.GroupVersion.String()
)

// AutoObjectCreationReconciler aoc reconciler
type AutoObjectCreationReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// SetupWithManager setup
func (r *AutoObjectCreationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&toolsaocv1.AutoObjectCreation{}).
		Complete(r)
}

// +kubebuilder:rbac:groups=tools.aoc.github.com,resources=autoobjectcreations,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=tools.aoc.github.com,resources=autoobjectcreations/status,verbs=get;update;patch

// Reconcile k8s reconcile
func (r *AutoObjectCreationReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("autoobjectcreation", aocGV)
	var aoc toolsaocv1.AutoObjectCreation
	err := r.Get(ctx, req.NamespacedName, &aoc)

	if err != nil {
		if k8sErrors.IsNotFound(err) {
			// Object not found, return. Created objects are automatically garbage collected
			return ctrl.Result{}, nil
		}

		// Error reading the object - requeue the request.
		return ctrl.Result{}, err
	}

	namespaces, err := r.FindNamespacesByAnnotation(aoc.Spec.Trigger.Annotations)
	var listErrors string
	for _, namespace := range namespaces {
		if err := r.UpdateByNamespace(aoc, namespace); err != nil {
			listErrors += err.Error() + "\n"
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

// FindNamespaces find all namespaces
func (r *AutoObjectCreationReconciler) FindNamespaces() (namespaces []corev1.Namespace, err error) {
	namespacesList := &corev1.NamespaceList{}
	err = r.Client.List(context.Background(), namespacesList)
	namespaces = namespacesList.Items

	return
}

// FindNamespacesByAnnotation find namespaces by annotation map
func (r *AutoObjectCreationReconciler) FindNamespacesByAnnotation(annotations map[string]string) ([]corev1.Namespace, error) {
	namespaces, err := r.FindNamespaces()
	if err != nil {
		return nil, err
	}

	var foundedNamespaces []corev1.Namespace

	for _, namespace := range namespaces {
		var found = true
		for annotation := range annotations {
			if _, found = namespace.Annotations[annotation]; found {
				break
			}
		}

		if found {
			foundedNamespaces = append(foundedNamespaces, namespace)
		}
	}

	return foundedNamespaces, nil
}

// UpdateByNamespace update namespace
func (r *AutoObjectCreationReconciler) UpdateByNamespace(aoc toolsaocv1.AutoObjectCreation, namespace corev1.Namespace) error {
	ctx := context.Background()
	log := r.Log.WithValues("autoobjectcreation", aocGV)
	reference := "[" + aoc.Spec.Template.Kind + "(" + aoc.Spec.Template.Name + ")] at " + namespace.ObjectMeta.Name + " namespace"
	log.Info("Ready to process " + reference)

	processor := processor.Processor{Client: r.Client}
	newObj, gvk, err := processor.ToObject(aoc.Spec.Template, namespace)

	if err != nil {
		return errors.New("Error serializing " + reference + ": " + err.Error())
	}

	log.Info("Object encoded succefully " + reference)

	findObj, err := processor.GetObject(
		*gvk,
		types.NamespacedName{
			Namespace: namespace.Name,
			Name:      aoc.Spec.Template.Name,
		},
	)

	// controllerutil.CreateOrUpdate(ctx, r.Client, &newObj, func() error {
	// 	return nil
	// })

	if err != nil && k8sErrors.IsNotFound(err) {
		log.Info("Creating new object " + reference)
		err := r.Client.Create(ctx, &newObj)

		if err == nil {
			log.Info("Create succefully " + reference)
		} else {
			return errors.New("Error creating object " + reference + ": " + err.Error())
		}
	} else {
		if err == nil {
			findObj.Object["spec"] = newObj.Object["spec"]
			err := r.Client.Update(ctx, &findObj)

			if err == nil {
				log.Info("Update succefully " + reference)
			} else {
				return errors.New("Error updating object " + reference + ": " + err.Error())
			}
		} else {
			return errors.New("Error getting object " + reference + ": " + err.Error())
		}
	}

	return nil
}
