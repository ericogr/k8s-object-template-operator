package controllers

import (
	"strings"
	"text/template"

	otv1 "github.com/ericogr/k8s-aoc/apis/template.totvs.app/v1"
)

// getStrFromTemplate get string from template
func getStrFromTemplate(template otv1.Template) string {
	return `
apiVersion: ` + template.APIVersion + `
kind: ` + template.Kind + `
spec:
  ` + addIdentation(template.Spec)
}

func addIdentation(str string) string {
	return strings.ReplaceAll(str, "\n", "\n  ")
}

func executeTemplate(templateYAML string, values map[string]string) (string, error) {
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
