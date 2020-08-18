package controllers

import (
	"context"
	"errors"

	toolsaocv1 "github.com/ericogr/k8s-aoc/apis/template.totvs.app/v1"
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
func (r *Common) UpdateObjectByNamespace(aoc toolsaocv1.ObjectTemplate, namespaceName string, values map[string]string) error {
	ctx := context.Background()
	log := r.Log.WithValues("objecttemplate", aocGV)
	reference := "[" + aoc.Spec.Template.Kind + "(" + aoc.Spec.Template.Name + ")] at " + namespaceName + " namespace"
	log.Info("Ready to process " + reference)

	processor := processor.Processor{Client: r.Client}
	newObj, gvk, err := processor.ToObject(aoc.Spec.Template, values, namespaceName)

	if err != nil {
		return errors.New("Error serializing " + reference + ": " + err.Error())
	}

	log.Info("Object encoded succefully " + reference)

	findObj, err := processor.GetObject(
		*gvk,
		types.NamespacedName{
			Namespace: namespaceName,
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

// FindObjectTemplateParamsByTemplateName find all aoc params by template name
func (r *Common) FindObjectTemplateParamsByTemplateName(templateName string) ([]toolsaocv1.ObjectTemplateParams, error) {
	aocParams, err := r.FindObjectTemplateParams()
	if err != nil {
		return nil, err
	}

	var aocParamsRet []toolsaocv1.ObjectTemplateParams

	for _, aocParam := range aocParams {
		_, err := aocParam.Spec.GetParametersByTemplateName(templateName)

		if err == nil {
			aocParamsRet = append(aocParamsRet, aocParam)
		}
	}

	return aocParamsRet, nil
}

// FindObjectTemplateParams find all aoc params
func (r *Common) FindObjectTemplateParams() ([]toolsaocv1.ObjectTemplateParams, error) {
	aocParamsList := &toolsaocv1.ObjectTemplateParamsList{}
	err := r.Client.List(context.Background(), aocParamsList)
	return aocParamsList.Items, err
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

// GetAOCByName get aoc by name
func (r *Common) GetAOCByName(name string) (*toolsaocv1.ObjectTemplate, error) {
	aocs, err := r.FindAOCs()

	if err != nil {
		return nil, err
	}

	for _, aoc := range aocs {
		if aoc.Spec.Template.Name == name {
			return &aoc, nil
		}
	}

	return nil, nil
}

// FindAOCs find all AOC
func (r *Common) FindAOCs() (aoc []toolsaocv1.ObjectTemplate, err error) {
	aocList := &toolsaocv1.ObjectTemplateList{}
	err = r.Client.List(context.Background(), aocList)
	aoc = aocList.Items

	return
}
