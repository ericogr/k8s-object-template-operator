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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Object Utilities", func() {
	var (
		noIdent string
		ident   string
		k8sObj  string
	)

	BeforeEach(func() {
		noIdent = `image: image-name
tag: \"v1.2\"
params:
  param1: value1
  param2: value2`

		ident = `  image: image-name
  tag: \"v1.2\"
  params:
    param1: value1
    param2: value2`

		k8sObj = `---
apiVersion: xpto/v1
kind: Test
spec:
  image: image-name
  tag: \"v1.2\"
  params:
    param1: value1
    param2: value2`
	})

	Describe("Identation", func() {
		Context("Text with no identation", func() {
			It("Should be idented", func() {
				Expect(ident).To(Equal(addIdentation(noIdent)))
			})
		})
	})

	Describe("Object", func() {
		Context("From parameters", func() {
			It("Should be kubernetes object as string", func() {
				Expect(k8sObj).To(Equal(getStringObject("xpto/v1", "Test", noIdent)))
			})
		})
	})

})
