package controllers

import (
	"context"
	"fmt"

	otv1 "github.com/ericogr/k8s-object-template/apis/template.ericogr.github.com/v1"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	otGV = otv1.GroupVersion.String()
)

// Common common controllers things
type Common struct {
	client.Client
	Log logr.Logger
}

// UpdateObjectsByTemplate update object
func (c *Common) UpdateObjectsByTemplate(ot otv1.ObjectTemplate, owners []metav1.OwnerReference, namespaceName string, values map[string]string) error {
	for _, obj := range ot.Spec.Objects {
		err := c.UpdateSingleObjectByTemplate(obj, owners, namespaceName, values)

		if err != nil {
			return err
		}
	}

	return nil
}

// UpdateSingleObjectByTemplate update object
func (c *Common) UpdateSingleObjectByTemplate(obj otv1.Object, owners []metav1.OwnerReference, namespaceName string, values map[string]string) error {
	ctx := context.Background()
	log := c.Log.WithValues("objecttemplate", otGV)
	reference := fmt.Sprintf("[%v(%v)] at %v namespace", obj.Kind, obj.Name, namespaceName)
	log.Info(fmt.Sprintf("Ready to process %v", reference))

	newObj, gvk, err := c.ToObject(obj, owners, values, namespaceName)

	if err != nil {
		return fmt.Errorf("Error serializing %v: %v", reference, err.Error())
	}
	log.Info(fmt.Sprintf("Object encoded succefully %v", reference))

	findObj, err := c.GetObject(
		*gvk,
		types.NamespacedName{
			Namespace: namespaceName,
			Name:      obj.Name,
		},
	)

	// controllerutil.CreateOrUpdate(ctx, c.Client, &newObj, func() error {
	// 	return nil
	// })

	if err != nil && k8sErrors.IsNotFound(err) {
		log.Info(fmt.Sprintf("Creating new object %v", reference))
		err := c.Client.Create(ctx, &newObj)

		if err == nil {
			log.Info(fmt.Sprintf("Create succefully %v", reference))
		} else {
			return fmt.Errorf("Error creating object %v: %v", reference, err.Error())
		}
	} else {
		if err == nil {
			findObj.Object["spec"] = newObj.Object["spec"]
			findObj.SetLabels(newObj.GetLabels())
			findObj.SetAnnotations(newObj.GetAnnotations())
			err := c.Client.Update(ctx, &findObj)

			if err == nil {
				log.Info(fmt.Sprintf("Update succefully %v", reference))
			} else {
				return fmt.Errorf("Error updating object %v: %v", reference, err.Error())
			}
		} else {
			return fmt.Errorf("Error getting object %v: %v", reference, err.Error())
		}
	}

	return nil
}

// FindObjectTemplateParamsByTemplateName find all ot params by template name
func (c *Common) FindObjectTemplateParamsByTemplateName(templateName string) ([]otv1.ObjectTemplateParams, error) {
	otParams, err := c.FindObjectTemplateParams()
	if err != nil {
		return nil, err
	}

	var otParamsRet []otv1.ObjectTemplateParams

	for _, otParam := range otParams {
		_, err := otParam.Spec.GetParametersByTemplateName(templateName)

		if err == nil {
			otParamsRet = append(otParamsRet, otParam)
		}
	}

	return otParamsRet, nil
}

// FindObjectTemplateParams find all ot params
func (c *Common) FindObjectTemplateParams() ([]otv1.ObjectTemplateParams, error) {
	otParamsList := &otv1.ObjectTemplateParamsList{}
	err := c.Client.List(context.Background(), otParamsList)
	return otParamsList.Items, err
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

// GetObjectTemplateByName get object template by name
func (c *Common) GetObjectTemplateByName(name string) (*otv1.ObjectTemplate, error) {
	ots, err := c.FindObjectTemplates()

	if err != nil {
		return nil, err
	}

	for _, ot := range ots {
		if ot.Name == name {
			return &ot, nil
		}
	}

	return nil, nil
}

// FindObjectTemplates find all object templates
func (c *Common) FindObjectTemplates() (ots []otv1.ObjectTemplate, err error) {
	otList := &otv1.ObjectTemplateList{}
	err = c.Client.List(context.Background(), otList)
	ots = otList.Items

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
func (c *Common) ToObject(obj otv1.Object, owners []metav1.OwnerReference, values map[string]string, namespaceName string) (unstructured.Unstructured, *schema.GroupVersionKind, error) {
	templateValues := c.addRuntimeVariablesToMap(values, obj, namespaceName)
	templateYAML := getStringObject(obj.APIVersion, obj.Kind, obj.Spec)
	templateYAMLExecuted, err := executeTemplate(templateYAML, templateValues)

	if err != nil {
		return unstructured.Unstructured{}, nil, err
	}

	object := unstructured.Unstructured{}
	dec := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	_, _, err = dec.Decode([]byte(templateYAMLExecuted), nil, &object)

	if err != nil {
		return object, nil, err
	}

	gvk := schema.FromAPIVersionAndKind(obj.APIVersion, obj.Kind)

	object.SetNamespace(namespaceName)
	object.SetGroupVersionKind(gvk)
	object.SetName(obj.Name)
	object.SetLabels(obj.Metadata.Labels)
	object.SetAnnotations(obj.Metadata.Annotations)
	object.SetOwnerReferences(owners)

	return object, &gvk, nil
}

// UpdateStatus update object status
func (c *Common) UpdateStatus(ctx context.Context, obj runtime.Object) {
	if err := c.Status().Update(ctx, obj); err != nil {
		c.Log.Error(err, fmt.Sprintf("Unable to update %v status", obj.GetObjectKind().GroupVersionKind()))
	}
}

func (c *Common) addRuntimeVariablesToMap(values map[string]string, obj otv1.Object, namespaceName string) map[string]string {
	newMap := copyMap(values)

	newMap["__namespace"] = namespaceName
	newMap["__apiVersion"] = obj.APIVersion
	newMap["__kind"] = obj.Kind
	newMap["__name"] = obj.Name

	return newMap
}

func copyMap(values map[string]string) map[string]string {
	newMap := make(map[string]string)

	for k, v := range values {
		newMap[k] = v
	}

	return newMap
}
