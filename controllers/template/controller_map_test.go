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

	otv1 "github.com/ericogr/k8s-object-template/apis/v1"
)

var _ = Describe("ObjectTemplateParams controller (ConfigMap)", func() {
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
			By("By creating a new ObjectTemplate as a kubernetes map")
			ctx := context.Background()
			objectTemplate := &otv1.ObjectTemplate{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "template.k8s.ericogr.com.br/v1",
					Kind:       "ObjectTemplate",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: ObjectTemplateName,
				},
				Spec: otv1.ObjectTemplateSpec{
					Description: "namespace-template",
					Objects: []otv1.Object{
						{
							Kind: "ConfigMap",
							Metadata: otv1.Metadata{
								Annotations: map[string]string{
									"annotation1": "value_annotation1",
									"annotation2": "value_annotation2",
								},
								Labels: map[string]string{
									"label1": "value_label1",
									"label2": "value_label2",
								},
							},
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
				return err == nil
			}, timeout, interval).Should(BeTrue())
			Expect(createdObjectTemplate.Spec.Objects).Should(HaveLen(1))

			By("Creating a new ObjectTemplateParam")
			objectTemplateParams := &otv1.ObjectTemplateParams{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "template.k8s.ericogr.com.br/v1",
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
				return err == nil
			}, timeout, interval).Should(BeTrue())
			Expect(configmap.Data["player_initial_lives"]).Should(BeIdenticalTo("3"))
			Expect(configmap.Data["ui_properties_file_name"]).Should(BeIdenticalTo("user-interface.properties"))
			Expect(configmap.ObjectMeta.Annotations["annotation1"]).Should(BeIdenticalTo("value_annotation1"))
			Expect(configmap.ObjectMeta.Annotations["annotation2"]).Should(BeIdenticalTo("value_annotation2"))
			Expect(configmap.ObjectMeta.Labels["label1"]).Should(BeIdenticalTo("value_label1"))
			Expect(configmap.ObjectMeta.Labels["label2"]).Should(BeIdenticalTo("value_label2"))

			By("By updating object template")
			Expect(k8sClient.Get(ctx, types.NamespacedName{Name: ObjectTemplateName}, createdObjectTemplate)).Should(Succeed())

			createdObjectTemplate.Spec.Objects[0].Metadata.Annotations["annotation2"] = "value_annotation2_new"
			createdObjectTemplate.Spec.Objects[0].Metadata.Annotations["annotation3"] = "value_annotation3"
			createdObjectTemplate.Spec.Objects[0].Metadata.Labels["label1"] = "value_label1_new"
			createdObjectTemplate.Spec.Objects[0].Metadata.Labels["label3"] = "value_label3"
			createdObjectTemplate.Spec.Objects[0].TemplateBody = `data:
  new_player_initial_lives: "{{ .lives }}"
  new_ui_properties_file_name: "{{ .properties_file }}"`
			Expect(k8sClient.Update(ctx, createdObjectTemplate)).Should(Succeed())
			Eventually(func() bool {
				err := k8sClient.Get(ctx, types.NamespacedName{Name: NewObjectName, Namespace: ObjectTemplateParamsNamespace}, &configmap)

				if err != nil {
					return false
				}

				return configmap.Data["new_player_initial_lives"] == "3" &&
					configmap.Data["new_ui_properties_file_name"] == "user-interface.properties" &&
					configmap.Annotations["annotation1"] == "value_annotation1" &&
					configmap.Annotations["annotation2"] == "value_annotation2_new" &&
					configmap.Annotations["annotation3"] == "value_annotation3" &&
					configmap.Labels["label1"] == "value_label1_new" &&
					configmap.Labels["label2"] == "value_label2" &&
					configmap.Labels["label3"] == "value_label3"
			}, timeout, interval).Should(BeTrue())
		})
	})
})
