package controllers

import (
	"context"
	"errors"

	toolsaocv1 "github.com/ericogr/k8s-aoc/apis/tools/v1"
	"github.com/ericogr/k8s-aoc/pkg/processor"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Common common controllers things
type Common struct {
	client.Client
	Log logr.Logger
}

// UpdateObjectByNamespace update namespace
func (r *Common) UpdateObjectByNamespace(aoc toolsaocv1.AutoObjectCreation, namespace corev1.Namespace) error {
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

// FindNamespacesByAnnotation find namespaces by annotation map
func (r *Common) FindNamespacesByAnnotation(annotations map[string]string) ([]corev1.Namespace, error) {
	namespaces, err := r.FindNamespaces()
	if err != nil {
		return nil, err
	}

	var foundedNamespaces []corev1.Namespace

	for _, namespace := range namespaces {
		if r.ValidateNamespace(namespace, annotations) {
			foundedNamespaces = append(foundedNamespaces, namespace)
		}
	}

	return foundedNamespaces, nil
}

// ValidateNamespace validate by annotations
func (r *Common) ValidateNamespace(namespace corev1.Namespace, annotations map[string]string) (found bool) {
	found = true
	for annotation := range annotations {
		if _, found = namespace.Annotations[annotation]; !found {
			break
		}
	}

	return
}

// FindNamespaces find all namespaces
func (r *Common) FindNamespaces() (namespaces []corev1.Namespace, err error) {
	namespacesList := &corev1.NamespaceList{}
	err = r.Client.List(context.Background(), namespacesList)
	namespaces = namespacesList.Items

	return
}

// FindAOCs find all AOC
func (r *Common) FindAOCs() (aoc []toolsaocv1.AutoObjectCreation, err error) {
	aocList := &toolsaocv1.AutoObjectCreationList{}
	err = r.Client.List(context.Background(), aocList)
	aoc = aocList.Items

	return
}
