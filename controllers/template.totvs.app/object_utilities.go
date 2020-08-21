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
	"strings"
	"text/template"

	otv1 "github.com/ericogr/k8s-object-template/apis/template.totvs.app/v1"
)

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
