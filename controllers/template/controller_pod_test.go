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

var _ = Describe("ObjectTemplateParams controller (POD)", func() {
	const (
		ObjectTemplateParamsNamespace = "default"
		ObjectTemplateParamsName      = "otp-pod-name"
		ObjectTemplateName            = "ot-pod-name"
		NewObjectName                 = "new-pod-name"
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
					Parameters: []otv1.Parameter{
						{
							Name:    "imageName",
							Default: "latest",
						},
						{
							Name:    "containerName",
							Default: "{{ .__namespace }}",
						},
					},
					Objects: []otv1.Object{
						{
							Kind: "Pod",
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
							TemplateBody: `spec:
  containers:
  - name: {{ .containerName }}
    image: {{ .imageName }}
  activeDeadlineSeconds: 60
`,
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
								"imageName": "nginx",
							},
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, objectTemplateParams)).Should(Succeed())

			By("By checking object in template was created")
			var pod corev1.Pod
			Eventually(func() bool {
				err := k8sClient.Get(
					ctx,
					types.NamespacedName{Name: NewObjectName, Namespace: ObjectTemplateParamsNamespace},
					&pod)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			Expect(pod.Spec.Containers[0].Name).Should(BeIdenticalTo(ObjectTemplateParamsNamespace))
			Expect(pod.Spec.Containers[0].Image).Should(BeIdenticalTo("nginx"))
			Expect(*pod.Spec.ActiveDeadlineSeconds == 60).Should(BeTrue())
			Expect(pod.ObjectMeta.Annotations["annotation1"]).Should(BeIdenticalTo("value_annotation1"))
			Expect(pod.ObjectMeta.Annotations["annotation2"]).Should(BeIdenticalTo("value_annotation2"))
			Expect(pod.ObjectMeta.Labels["label1"]).Should(BeIdenticalTo("value_label1"))
			Expect(pod.ObjectMeta.Labels["label2"]).Should(BeIdenticalTo("value_label2"))

			By("By updating object template")
			Expect(k8sClient.Get(ctx, types.NamespacedName{Name: ObjectTemplateName}, createdObjectTemplate)).Should(Succeed())

			createdObjectTemplate.Spec.Objects[0].Metadata.Annotations["annotation2"] = "value_annotation2_new"
			createdObjectTemplate.Spec.Objects[0].Metadata.Annotations["annotation3"] = "value_annotation3"
			createdObjectTemplate.Spec.Objects[0].Metadata.Labels["label1"] = "value_label1_new"
			createdObjectTemplate.Spec.Objects[0].Metadata.Labels["label3"] = "value_label3"
			createdObjectTemplate.Spec.Objects[0].TemplateBody = `spec:
  containers:
  - name: {{ .containerName }}
    image: {{ .imageName }}
  activeDeadlineSeconds: 30
`
			Expect(k8sClient.Update(ctx, createdObjectTemplate)).Should(Succeed())
			Eventually(func() bool {
				err := k8sClient.Get(ctx, types.NamespacedName{Name: NewObjectName, Namespace: ObjectTemplateParamsNamespace}, &pod)

				if err != nil {
					return false
				}

				return pod.Spec.Containers[0].Name == ObjectTemplateParamsNamespace &&
					pod.Spec.Containers[0].Image == "nginx" &&
					*pod.Spec.ActiveDeadlineSeconds == 30 &&
					pod.Annotations["annotation1"] == "value_annotation1" &&
					pod.Annotations["annotation2"] == "value_annotation2_new" &&
					pod.Annotations["annotation3"] == "value_annotation3" &&
					pod.Labels["label1"] == "value_label1_new" &&
					pod.Labels["label2"] == "value_label2" &&
					pod.Labels["label3"] == "value_label3"
			}, timeout, interval).Should(BeTrue())
		})
	})
})
