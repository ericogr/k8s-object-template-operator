package aoc

import (
	"context"
	"fmt"

	aocv1 "github.com/ericogr/k8s-aoc/api/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// https://ymmt2005.hatenablog.com/entry/2020/04/14/An_example_of_using_dynamic_client_of_k8s.io/client-go

// Processor process template parameters to create object
type Processor struct {
	Client client.Client
}

// CreateObject create object from template
func (aoc *Processor) CreateObject(template aocv1.Template, namespace corev1.Namespace) error {
	templateYAML := aoc.getStrFromTemplate(template)
	object := &unstructured.Unstructured{}
	dec := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	_, _, err := dec.Decode([]byte(templateYAML), nil, object)

	if err != nil {
		fmt.Println("-------------unknown error!", err)
		return err
	}

	object.SetName(template.Name)
	object.SetNamespace(namespace.GetName())

	return aoc.Client.Create(context.Background(), object)
}

// getStrFromTemplate get string from template
func (aoc *Processor) getStrFromTemplate(template aocv1.Template) string {
	return `
apiVersion: ` + template.APIVersion + `
kind: ` + template.Kind + `
spec:
  ` + template.Spec
}
