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

package v1

import "errors"

// GetParametersByTemplateName get parameters values by template name
func (a *AOCParamsSpec) GetParametersByTemplateName(templateName string) (Parameters, error) {
	for _, parameter := range a.Templates {
		if parameter.Name == templateName {
			return parameter, nil
		}
	}

	return Parameters{}, errors.New("parameter " + templateName + " not found")
}

// SetValuesByName set values for specific parameter template
func (a *AOCParamsSpec) SetValuesByName(parameterName string, values map[string]string) bool {
	for _, parameter := range a.Templates {
		if parameter.Name == parameterName {
			parameter.Values = values
			return true
		}
	}

	return false
}
