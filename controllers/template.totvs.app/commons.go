package controllers

import (
	"context"
	"errors"
	"strings"
	"text/template"

	otv1 "github.com/ericogr/k8s-aoc/apis/template.totvs.app/v1"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Common common controllers things
type Common struct {
	client.Client
	Log logr.Logger
}

// UpdateObjectByNamespace update namespace
func (c *Common) UpdateObjectByNamespace(aoc otv1.ObjectTemplate, namespaceName string, values map[string]string) error {
	ctx := context.Background()
	log := c.Log.WithValues("objecttemplate", aocGV)
	reference := "[" + aoc.Spec.Template.Kind + "(" + aoc.Spec.Template.Name + ")] at " + namespaceName + " namespace"
	log.Info("Ready to process " + reference)

	newObj, gvk, err := c.ToObject(aoc.Spec.Template, values, namespaceName)

	if err != nil {
		return errors.New("Error serializing " + reference + ": " + err.Error())
	}

	log.Info("Object encoded succefully " + reference)

	findObj, err := c.GetObject(
		*gvk,
		types.NamespacedName{
			Namespace: namespaceName,
			Name:      aoc.Spec.Template.Name,
		},
	)

	// controllerutil.CreateOrUpdate(ctx, c.Client, &newObj, func() error {
	// 	return nil
	// })

	if err != nil && k8sErrors.IsNotFound(err) {
		log.Info("Creating new object " + reference)
		err := c.Client.Create(ctx, &newObj)

		if err == nil {
			log.Info("Create succefully " + reference)
		} else {
			return errors.New("Error creating object " + reference + ": " + err.Error())
		}
	} else {
		if err == nil {
			findObj.Object["spec"] = newObj.Object["spec"]
			err := c.Client.Update(ctx, &findObj)

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
func (c *Common) FindObjectTemplateParamsByTemplateName(templateName string) ([]otv1.ObjectTemplateParams, error) {
	aocParams, err := c.FindObjectTemplateParams()
	if err != nil {
		return nil, err
	}

	var aocParamsRet []otv1.ObjectTemplateParams

	for _, aocParam := range aocParams {
		_, err := aocParam.Spec.GetParametersByTemplateName(templateName)

		if err == nil {
			aocParamsRet = append(aocParamsRet, aocParam)
		}
	}

	return aocParamsRet, nil
}

// FindObjectTemplateParams find all aoc params
func (c *Common) FindObjectTemplateParams() ([]otv1.ObjectTemplateParams, error) {
	aocParamsList := &otv1.ObjectTemplateParamsList{}
	err := c.Client.List(context.Background(), aocParamsList)
	return aocParamsList.Items, err
}

// ValidateNamespace validate by annotations
func (c *Common) ValidateNamespace(namespace corev1.Namespace, annotations map[string]string) (found bool) {
	found = true
	for annotation := range annotations {
		if _, found = namespace.Annotations[annotation]; !found {
			break
		}
	}

	return
}

// GetAOCByName get aoc by name
func (c *Common) GetAOCByName(name string) (*otv1.ObjectTemplate, error) {
	aocs, err := c.FindAOCs()

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
func (c *Common) FindAOCs() (aoc []otv1.ObjectTemplate, err error) {
	aocList := &otv1.ObjectTemplateList{}
	err = c.Client.List(context.Background(), aocList)
	aoc = aocList.Items

	return
}

// GetObject get any object
func (c *Common) GetObject(gvk schema.GroupVersionKind, nn types.NamespacedName) (obj unstructured.Unstructured, err error) {
	ctx := context.Background()
	obj = unstructured.Unstructured{}
	obj.SetGroupVersionKind(gvk)
	err = c.Client.Get(ctx, nn, &obj)

	return
}

// GetObjectSimplified get any object
func (c *Common) GetObjectSimplified(groupversion string, kind string, namespace string, name string) (obj unstructured.Unstructured, err error) {
	return c.GetObject(
		schema.FromAPIVersionAndKind(groupversion, kind),
		types.NamespacedName{Namespace: namespace, Name: name},
	)
}

// ToObject process object from template
func (c *Common) ToObject(template otv1.Template, values map[string]string, namespaceName string) (unstructured.Unstructured, *schema.GroupVersionKind, error) {
	values["__namespace"] = namespaceName
	templateYAML := c.getStrFromTemplate(template)
	templateYAMLExecuted, err := c.executeTemplate(templateYAML, values)

	if err != nil {
		return unstructured.Unstructured{}, nil, err
	}

	object := unstructured.Unstructured{}
	dec := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	_, _, err = dec.Decode([]byte(templateYAMLExecuted), nil, &object)

	if err != nil {
		return object, nil, err
	}

	gvk := schema.FromAPIVersionAndKind(template.APIVersion, template.Kind)

	object.SetNamespace(namespaceName)
	object.SetGroupVersionKind(gvk)
	object.SetName(template.Name)

	return object, &gvk, nil
}

// getStrFromTemplate get string from template
func (c *Common) getStrFromTemplate(template otv1.Template) string {
	return `
apiVersion: ` + template.APIVersion + `
kind: ` + template.Kind + `
spec:
  ` + c.addIdentation(template.Spec)
}

func (c *Common) addIdentation(str string) string {
	return strings.ReplaceAll(str, "\n", "\n  ")
}

func (c *Common) executeTemplate(templateYAML string, values map[string]string) (string, error) {
	template, err := template.New("template").Parse(templateYAML)

	if err != nil {
		return "", err
	}

	sb := strings.Builder{}
	err = template.Execute(&sb, values)

	if err != nil {
		return "", err
	}

	return sb.String(), nil
}
