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
	otv1 "github.com/ericogr/k8s-object-template/apis/template.ericogr.github.com/v1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Controller commons", func() {
	BeforeEach(func() {
	})

	Describe("Map parameters", func() {
		Context("With no params", func() {
			It("Should be filled with system params", func() {
				var obj otv1.Object
				var common = Common{}
				newmap := common.addRuntimeVariablesToMap(map[string]string{}, obj, "test")

				Expect(newmap).To(HaveLen(4))
			})
		})
	})
})
