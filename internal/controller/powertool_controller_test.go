/*
Copyright 2025.

*/

package controller

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	toev1alpha1 "toe/api/v1alpha1"
)

var _ = Describe("PowerTool Controller", func() {
	Context("When reconciling a PowerTool resource", func() {
		const resourceName = "test-powertool"
		const configName = "aperf-config"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default",
		}

		configNamespacedName := types.NamespacedName{
			Name:      configName,
			Namespace: "toe-system",
		}

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
			config := &toev1alpha1.PowerToolConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      configName,
					Namespace: "toe-system",
				},
				Spec: toev1alpha1.PowerToolConfigSpec{
					Name:  "aperf",
					Image: "test-registry/aperf:latest",
					SecurityContext: toev1alpha1.SecuritySpec{
						AllowPrivileged: boolPtr(true),
					},
				},
			}
			err = k8sClient.Get(ctx, configNamespacedName, &toev1alpha1.PowerToolConfig{})
			if err != nil && errors.IsNotFound(err) {
				Expect(k8sClient.Create(ctx, config)).To(Succeed())
			}

			By("creating the PowerTool resource")
			powerTool := &toev1alpha1.PowerTool{}
			err = k8sClient.Get(ctx, typeNamespacedName, powerTool)
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
									"app": "test-app",
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
			By("cleaning up the PowerTool resource")
			resource := &toev1alpha1.PowerTool{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			if err == nil {
				Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
			}

			By("cleaning up the PowerToolConfig")
			config := &toev1alpha1.PowerToolConfig{}
			err = k8sClient.Get(ctx, configNamespacedName, config)
			if err == nil {
				Expect(k8sClient.Delete(ctx, config)).To(Succeed())
			}
		})

		It("should successfully reconcile and initialize status", func() {
			By("reconciling the created resource")
			controllerReconciler := &PowerToolReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			By("checking that status is initialized")
			powerTool := &toev1alpha1.PowerTool{}
			err = k8sClient.Get(ctx, typeNamespacedName, powerTool)
			Expect(err).NotTo(HaveOccurred())

			// Verify status initialization
			Expect(powerTool.Status.Phase).NotTo(BeNil())
			Expect(*powerTool.Status.Phase).To(Equal("Pending"))
			Expect(powerTool.Status.StartedAt).NotTo(BeNil())
			Expect(powerTool.Status.SelectedPods).NotTo(BeNil())
			Expect(*powerTool.Status.SelectedPods).To(Equal(int32(0))) // No matching pods

			// Verify conditions
			Expect(powerTool.Status.Conditions).NotTo(BeEmpty())
			readyCondition := findCondition(powerTool.Status.Conditions, toev1alpha1.PowerToolConditionReady)
			Expect(readyCondition).NotTo(BeNil())
			Expect(readyCondition.Status).To(Equal("False"))
			Expect(readyCondition.Reason).To(Equal(toev1alpha1.ReasonTargetsSelected))
		})

		It("should handle missing PowerToolConfig gracefully", func() {
			By("deleting the PowerToolConfig")
			config := &toev1alpha1.PowerToolConfig{}
			err := k8sClient.Get(ctx, configNamespacedName, config)
			Expect(err).NotTo(HaveOccurred())
			Expect(k8sClient.Delete(ctx, config)).To(Succeed())

			By("reconciling without config")
			controllerReconciler := &PowerToolReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("PowerToolConfig not found"))
		})

		It("should detect conflicts with other PowerTools", func() {
			By("creating a target pod")
			pod := &corev1.Pod{
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
							Image: "nginx:latest",
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, pod)).To(Succeed())

			By("creating a second PowerTool targeting the same pod")
			conflictingPowerTool := &toev1alpha1.PowerTool{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "conflicting-powertool",
					Namespace: "default",
				},
				Spec: toev1alpha1.PowerToolSpec{
					Targets: toev1alpha1.TargetSpec{
						LabelSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"app": "test-app",
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
			Expect(k8sClient.Create(ctx, conflictingPowerTool)).To(Succeed())

			By("reconciling both PowerTools")
			controllerReconciler := &PowerToolReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			// First PowerTool should succeed
			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			// Second PowerTool should detect conflict
			conflictingNamespacedName := types.NamespacedName{
				Name:      "conflicting-powertool",
				Namespace: "default",
			}
			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: conflictingNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			By("checking conflict detection")
			conflictingResource := &toev1alpha1.PowerTool{}
			err = k8sClient.Get(ctx, conflictingNamespacedName, conflictingResource)
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() string {
				err := k8sClient.Get(ctx, conflictingNamespacedName, conflictingResource)
				if err != nil {
					return ""
				}
				if conflictingResource.Status.Phase != nil {
					return *conflictingResource.Status.Phase
				}
				return ""
			}, time.Second*5, time.Millisecond*100).Should(Equal("Conflicted"))

			// Cleanup
			Expect(k8sClient.Delete(ctx, conflictingPowerTool)).To(Succeed())
			Expect(k8sClient.Delete(ctx, pod)).To(Succeed())
		})
	})
})

// Helper functions
func findCondition(conditions []toev1alpha1.PowerToolCondition, conditionType string) *toev1alpha1.PowerToolCondition {
	for _, condition := range conditions {
		if condition.Type == conditionType {
			return &condition
		}
	}
	return nil
}

func boolPtr(b bool) *bool {
	return &b
}
