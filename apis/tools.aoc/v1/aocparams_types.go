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

// Values values
type Values struct {
	Values map[string]string `json:",inline"`
}

// AOCParamsSpec defines the desired state of AOCParams
type AOCParamsSpec struct {
	Parameters map[string]Values `json:"parameters"`
}

// AOCParamsStatus defines the observed state of AOCParams
type AOCParamsStatus struct {
}

// +kubebuilder:object:root=true

// AOCParams is the Schema for the aocparams API
type AOCParams struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AOCParamsSpec   `json:"spec,omitempty"`
	Status AOCParamsStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AOCParamsList contains a list of AOCParams
type AOCParamsList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AOCParams `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AOCParams{}, &AOCParamsList{})
}
