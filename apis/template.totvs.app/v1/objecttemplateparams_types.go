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

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Parameters values
type Parameters struct {
	Name   string            `json:"name"`
	Values map[string]string `json:"values,omitempty"`
}

// ObjectTemplateParamsSpec defines the desired state of ObjectTemplateParams
type ObjectTemplateParamsSpec struct {
	Templates []Parameters `json:"templates"`
}

// ObjectTemplateParamsStatus defines the observed state of ObjectTemplateParams
type ObjectTemplateParamsStatus struct {
}

// +kubebuilder:object:root=true

// ObjectTemplateParams is the Schema for the objecttemplateparams API
type ObjectTemplateParams struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ObjectTemplateParamsSpec   `json:"spec,omitempty"`
	Status ObjectTemplateParamsStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ObjectTemplateParamsList contains a list of ObjectTemplateParams
type ObjectTemplateParamsList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ObjectTemplateParams `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ObjectTemplateParams{}, &ObjectTemplateParamsList{})
}
