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

// Metadata metadata for object
type Metadata struct {
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

// Template defines template for spec creation
type Template struct {
	Name       string   `json:"name"`
	Kind       string   `json:"kind"`
	APIVersion string   `json:"apiVersion"`
	Metadata   Metadata `json:"metadata,omitempty"`
	Spec       string   `json:"spec"`
}

// AutoObjectCreationSpec defines the desired state of AutoObjectCreation
type AutoObjectCreationSpec struct {
	Template Template `json:"template"`
}

// AutoObjectCreationStatus defines the observed state of AutoObjectCreation
type AutoObjectCreationStatus struct {
	Status string `json:"status"`
}

// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="status",type=string,JSONPath=`.status.status`
// +kubebuilder:subresource:status

// AutoObjectCreation is the Schema for the autoobjectcreations API
type AutoObjectCreation struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AutoObjectCreationSpec   `json:"spec,omitempty"`
	Status AutoObjectCreationStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AutoObjectCreationList contains a list of AutoObjectCreation
type AutoObjectCreationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AutoObjectCreation `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AutoObjectCreation{}, &AutoObjectCreationList{})
}
