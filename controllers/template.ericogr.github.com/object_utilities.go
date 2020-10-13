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
)

func getStringObject(apiVersion string, kind string, templateBody string) string {
	sb := strings.Builder{}

	sb.WriteString("---\n")
	sb.WriteString("apiVersion: ")
	sb.WriteString(apiVersion)
	sb.WriteRune('\n')
	sb.WriteString("kind: ")
	sb.WriteString(kind)
	sb.WriteRune('\n')
	sb.WriteString(templateBody)

	return sb.String()
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
