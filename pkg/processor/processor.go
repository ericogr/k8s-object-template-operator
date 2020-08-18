package processor

import (
	"context"
	"strings"
	"text/template"

	toolsaocv1 "github.com/ericogr/k8s-aoc/apis/template.totvs.app/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Processor process template parameters to create object
type Processor struct {
	client.Client
}

// GetObjectSimplified get any object
func (aoc *Processor) GetObjectSimplified(groupversion string, kind string, namespace string, name string) (obj unstructured.Unstructured, err error) {
	return aoc.GetObject(
		schema.FromAPIVersionAndKind(groupversion, kind),
		types.NamespacedName{Namespace: namespace, Name: name},
	)
}

// GetObject get any object
func (aoc *Processor) GetObject(gvk schema.GroupVersionKind, nn types.NamespacedName) (obj unstructured.Unstructured, err error) {
	ctx := context.Background()
	obj = unstructured.Unstructured{}
	obj.SetGroupVersionKind(gvk)
	err = aoc.Client.Get(ctx, nn, &obj)

	return
}

// ToObject process object from template
func (aoc *Processor) ToObject(template toolsaocv1.Template, values map[string]string, namespaceName string) (unstructured.Unstructured, *schema.GroupVersionKind, error) {
	values["__namespace"] = namespaceName
	templateYAML := aoc.getStrFromTemplate(template)
	templateYAMLExecuted, err := aoc.executeTemplate(templateYAML, values)

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
func (aoc *Processor) getStrFromTemplate(template toolsaocv1.Template) string {
	return `
apiVersion: ` + template.APIVersion + `
kind: ` + template.Kind + `
spec:
  ` + aoc.addIdentation(template.Spec)
}

func (aoc *Processor) addIdentation(str string) string {
	return strings.ReplaceAll(str, "\n", "\n  ")
}

func (aoc *Processor) executeTemplate(templateYAML string, values map[string]string) (string, error) {
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
