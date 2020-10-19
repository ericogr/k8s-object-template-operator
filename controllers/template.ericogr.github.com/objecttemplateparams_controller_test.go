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
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	otv1 "github.com/ericogr/k8s-object-template/apis/template.ericogr.github.com/v1"
)

var _ = Describe("ObjectTemplateParams controller", func() {
	const (
		ObjectTemplateParamsNamespace = "default"
		ObjectTemplateParamsName      = "otp-name"
		ObjectTemplateName            = "ot-name"
		NewObjectName                 = "new-config-map"
		timeout                       = time.Second * 5
		interval                      = time.Second * 1
	)
	Context("When updating parameters from ObjectTemplateParams", func() {
		It("Should update templated object.", func() {
			By("By creating a new ObjectTemplate")
			ctx := context.Background()
			objectTemplate := &otv1.ObjectTemplate{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "template.ericogr.github.com/v1",
					Kind:       "ObjectTemplate",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: ObjectTemplateName,
				},
				Spec: otv1.ObjectTemplateSpec{
					Description: "namespace-template",
					Objects: []otv1.Object{
						{
							Kind:       "ConfigMap",
							APIVersion: "v1",
							Name:       NewObjectName,
							TemplateBody: `data:
  player_initial_lives: "{{ .lives }}"
  ui_properties_file_name: "{{ .properties_file }}"`,
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, objectTemplate)).Should(Succeed())

			createdObjectTemplate := &otv1.ObjectTemplate{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, types.NamespacedName{Name: ObjectTemplateName}, createdObjectTemplate)

				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())
			Expect(createdObjectTemplate.Spec.Objects).Should(HaveLen(1))

			By("Creating a new ObjectTemplateParam")
			objectTemplateParams := &otv1.ObjectTemplateParams{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "template.ericogr.github.com/v1",
					Kind:       "ObjectTemplateParam",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      ObjectTemplateParamsName,
					Namespace: ObjectTemplateParamsNamespace,
				},
				Spec: otv1.ObjectTemplateParamsSpec{
					Templates: []otv1.Parameters{
						{
							Name: ObjectTemplateName,
							Values: map[string]string{
								"lives":           "3",
								"properties_file": "user-interface.properties",
							},
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, objectTemplateParams)).Should(Succeed())

			By("By checking object in template was created")
			var configmap corev1.ConfigMap
			Eventually(func() bool {
				err := k8sClient.Get(
					ctx,
					types.NamespacedName{Name: NewObjectName, Namespace: ObjectTemplateParamsNamespace},
					&configmap)

				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())
			Expect(configmap.Data["player_initial_lives"]).Should(BeIdenticalTo("3"))
			Expect(configmap.Data["ui_properties_file_name"]).Should(BeIdenticalTo("user-interface.properties"))

			By("By updating object template")
			Expect(k8sClient.Get(ctx, types.NamespacedName{Name: ObjectTemplateName}, createdObjectTemplate)).Should(Succeed())
			createdObjectTemplate.Spec.Objects[0].TemplateBody = `data:
  new_player_initial_lives: "{{ .lives }}"
  new_ui_properties_file_name: "{{ .properties_file }}"`
			Expect(k8sClient.Update(ctx, createdObjectTemplate)).Should(Succeed())

			Eventually(func() bool {
				err := k8sClient.Get(ctx, types.NamespacedName{Name: NewObjectName, Namespace: ObjectTemplateParamsNamespace}, &configmap)

				if err != nil {
					return false
				}

				return configmap.Data["new_player_initial_lives"] == "3" && configmap.Data["new_ui_properties_file_name"] == "user-interface.properties"
			}, timeout, interval).Should(BeTrue())
		})
	})
})
