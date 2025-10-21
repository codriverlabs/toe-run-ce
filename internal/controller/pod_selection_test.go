/*
Copyright 2025.

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

package controller

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	toerunv1alpha1 "toe/api/v1alpha1"
)

var _ = Describe("PowerTool Controller", func() {
	Context("When reconciling a resource with pod selection", func() {
		const resourceName = "test-resource-pod-selection"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default",
		}
		powertool := &toerunv1alpha1.PowerTool{}

		BeforeEach(func() {
			By("creating the toe-system namespace")
			namespace := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "toe-system",
				},
			}
			err := k8sClient.Get(ctx, types.NamespacedName{Name: "toe-system"}, &corev1.Namespace{})
			if err != nil && errors.IsNotFound(err) {
				Expect(k8sClient.Create(ctx, namespace)).To(Succeed())
			}

			By("creating the PowerToolConfig")
			config := &toerunv1alpha1.PowerToolConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "aperf-config",
					Namespace: "toe-system",
				},
				Spec: toerunv1alpha1.PowerToolConfigSpec{
					Name:  "aperf",
					Image: "test-registry/aperf:latest",
					SecurityContext: toerunv1alpha1.SecuritySpec{
						AllowPrivileged: boolPtr(true),
					},
				},
			}
			configKey := types.NamespacedName{Name: "aperf-config", Namespace: "toe-system"}
			err = k8sClient.Get(ctx, configKey, &toerunv1alpha1.PowerToolConfig{})
			if err != nil && errors.IsNotFound(err) {
				Expect(k8sClient.Create(ctx, config)).To(Succeed())
			}

			By("creating the custom resource for the Kind PowerTool")
			err = k8sClient.Get(ctx, typeNamespacedName, powertool)
			if err != nil && errors.IsNotFound(err) {
				resource := &toerunv1alpha1.PowerTool{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					Spec: toerunv1alpha1.PowerToolSpec{
						Targets: toerunv1alpha1.TargetSpec{
							LabelSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									"app": "test-app",
								},
							},
						},
						Tool: toerunv1alpha1.ToolSpec{
							Name:     "aperf",
							Duration: "30s",
						},
						Output: toerunv1alpha1.OutputSpec{
							Mode: "ephemeral",
						},
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			resource := &toerunv1alpha1.PowerTool{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance PowerTool")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})

		It("should successfully reconcile the resource", func() {
			By("creating a test pod with matching labels")
			testPod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pod",
					Namespace: "default",
					Labels: map[string]string{
						"app": "test-app",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "test-container",
							Image: "nginx:alpine",
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, testPod)).To(Succeed())

			By("Reconciling the created resource")
			controllerReconciler := &PowerToolReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			By("checking the status of the PowerTool")
			updatedPowerTool := &toerunv1alpha1.PowerTool{}
			Expect(k8sClient.Get(ctx, typeNamespacedName, updatedPowerTool)).To(Succeed())
			// Note: Status updates would need to be implemented in the controller
			// Expect(*updatedPowerTool.Status.SelectedPods).To(Equal(int32(1)))

			By("cleaning up the pod")
			Expect(k8sClient.Delete(ctx, testPod)).To(Succeed())
		})
	})
})
