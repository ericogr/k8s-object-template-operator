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
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	toolsaocv1 "github.com/ericogr/k8s-aoc/api/v1"
	aocv1 "github.com/ericogr/k8s-aoc/pkg/aoc"
)

// AutoObjectCreationReconciler reconciles a AutoObjectCreation object
type AutoObjectCreationReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=tools.aoc.github.com,resources=autoobjectcreations,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=tools.aoc.github.com,resources=autoobjectcreations/status,verbs=get;update;patch

// Reconcile k8s reconcile
func (r *AutoObjectCreationReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("autoobjectcreation", req.NamespacedName)
	var aoc toolsaocv1.AutoObjectCreation

	err := r.Get(ctx, req.NamespacedName, &aoc)

	if err != nil {
		log.Error(err, "Error when refresh AutoObjectCreation")
		return ctrl.Result{}, err
	}

	namespaces, err := r.FindNamespacesByAnnotation(aoc.Spec.Trigger.Annotations)

	var errors string
	for _, namespace := range namespaces {
		log.Info("Ready to greate [" + aoc.Kind + "(" + aoc.Name + ")] at " + namespace.ObjectMeta.Name)

		processor := aocv1.Processor{Client: r.Client}
		err := processor.CreateObject(aoc.Spec.Template, namespace)

		if err != nil {
			strErr := "Error creating object [" + aoc.Kind + " (" + aoc.Name + ")] at " + namespace.ObjectMeta.Name
			errors = errors + strErr + "\n"
			log.Error(err, strErr)
		} else {
			log.Info("Successfully created object [" + aoc.Kind + " (" + aoc.Name + ")] at " + namespace.ObjectMeta.Name)
		}
	}

	if errors != "" {
		aoc.Status.Status = errors
	} else {
		aoc.Status.Status = "OK"
	}

	err = r.Status().Update(ctx, &aoc)
	if err != nil {
		log.Error(err, "Fail to update status")
	}

	return ctrl.Result{}, nil
}

// SetupWithManager setup
func (r *AutoObjectCreationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&toolsaocv1.AutoObjectCreation{}).
		Complete(r)
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
