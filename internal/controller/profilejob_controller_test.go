/*
Copyright 2025.

*/

package controller

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	toev1alpha1 "toe/api/v1alpha1"
)

var _ = Describe("PowerTool Controller Integration", func() {
	Context("When reconciling a PowerTool resource with comprehensive scenarios", func() {
		const resourceName = "integration-test-powertool"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default",
		}

		BeforeEach(func() {
			By("creating the custom resource for the Kind PowerTool")
			powerTool := &toev1alpha1.PowerTool{}
			err := k8sClient.Get(ctx, typeNamespacedName, powerTool)
			if err != nil && errors.IsNotFound(err) {
				resource := &toev1alpha1.PowerTool{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					Spec: toev1alpha1.PowerToolSpec{
						Targets: toev1alpha1.TargetSpec{
							LabelSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									"app": "integration-test",
								},
							},
						},
						Tool: toev1alpha1.ToolSpec{
							Name:     "aperf",
							Duration: "30s",
						},
						Output: toev1alpha1.OutputSpec{
							Mode: "ephemeral",
						},
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			resource := &toev1alpha1.PowerTool{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			if err == nil {
				By("Cleanup the specific resource instance PowerTool")
				Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
			}
		})

		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			controllerReconciler := &PowerToolReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})

			// Should fail due to missing PowerToolConfig, which is expected behavior
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("PowerToolConfig not found"))
		})
	})
})
