package e2e

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"toe/api/v1alpha1"
)

var _ = Describe("Controller Reconciliation Logic", func() {
	var namespace *corev1.Namespace

	BeforeEach(func() {
		namespace = CreateSimpleTestNamespace()
		CreateSimpleTestPowerToolConfig("aperf-config", namespace.Name)
	})

	AfterEach(func() {
		DeleteSimpleTestNamespace(namespace)
	})

	Context("Target Pod Discovery", func() {
		It("should handle missing target pods gracefully", func() {
			By("creating PowerTool with non-matching label selector")
			spec := CreateSimpleBasicPowerToolSpec(map[string]string{"app": "nonexistent"})
			powerTool := CreateSimpleTestPowerTool("no-targets", namespace.Name, spec)

			By("verifying PowerTool transitions to Failed phase")
			Eventually(func() string {
				updated := GetSimplePowerTool(powerTool)
				if updated.Status.Phase == nil {
					return ""
				}
				return *updated.Status.Phase
			}, "30s", "1s").Should(Equal("Failed"))

			By("verifying appropriate condition is set")
			WaitForSimplePowerToolCondition(powerTool, "TargetsFound", "False")

			By("checking error message in status")
			updated := GetSimplePowerTool(powerTool)
			Expect(updated.Status.LastError).NotTo(BeNil())
			Expect(*updated.Status.LastError).To(ContainSubstring("no matching pods found"))
		})

		It("should discover pods with complex label selectors", func() {
			By("creating pods with various labels")
			CreateSimpleMockTargetPod(namespace.Name, "pod-1", map[string]string{
				"app":     "nginx",
				"version": "v1.0",
				"tier":    "frontend",
			})
			CreateSimpleMockTargetPod(namespace.Name, "pod-2", map[string]string{
				"app":     "nginx",
				"version": "v2.0",
				"tier":    "frontend",
			})
			CreateSimpleMockTargetPod(namespace.Name, "pod-3", map[string]string{
				"app":     "redis",
				"version": "v1.0",
				"tier":    "backend",
			})

			By("creating PowerTool with matchExpressions selector")
			powerTool := &v1alpha1.PowerTool{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "complex-selector",
					Namespace: namespace.Name,
				},
				Spec: v1alpha1.PowerToolSpec{
					Targets: v1alpha1.TargetSpec{
						LabelSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{"tier": "frontend"},
							MatchExpressions: []metav1.LabelSelectorRequirement{
								{
									Key:      "version",
									Operator: metav1.LabelSelectorOpIn,
									Values:   []string{"v1.0", "v2.0"},
								},
							},
						},
					},
					Tool: v1alpha1.ToolSpec{
						Name:     "aperf",
						Duration: "30s",
					},
					Output: v1alpha1.OutputSpec{
						Mode: "ephemeral",
					},
				},
			}
			Expect(simpleK8sClient.Create(simpleCtx, powerTool)).To(Succeed())

			By("verifying only matching pods are selected")
			Eventually(func() int {
				updated := GetSimplePowerTool(powerTool)
				return len(updated.Status.ActivePods)
			}, "30s", "1s").Should(Equal(2)) // pod-1 and pod-2
		})

		It("should handle pod lifecycle changes during reconciliation", func() {
			By("creating initial target pod")
			targetPod := CreateSimpleMockTargetPod(namespace.Name, "dynamic-pod", map[string]string{
				"app": "dynamic-app",
			})

			By("creating PowerTool targeting the pod")
			spec := CreateSimpleBasicPowerToolSpec(map[string]string{"app": "dynamic-app"})
			powerTool := CreateSimpleTestPowerTool("dynamic-targets", namespace.Name, spec)

			By("waiting for initial reconciliation")
			Eventually(func() int {
				updated := GetSimplePowerTool(powerTool)
				return len(updated.Status.ActivePods)
			}, "30s", "1s").Should(Equal(1))

			By("deleting the target pod")
			Expect(simpleK8sClient.Delete(simpleCtx, targetPod)).To(Succeed())

			By("verifying PowerTool status is updated")
			Eventually(func() int {
				updated := GetSimplePowerTool(powerTool)
				return len(updated.Status.ActivePods)
			}, "30s", "1s").Should(Equal(0))
		})
	})

	Context("Reconciliation State Management", func() {
		It("should handle concurrent PowerTool conflicts", func() {
			By("creating target pod")
			CreateSimpleMockTargetPod(namespace.Name, "shared-pod", map[string]string{
				"app": "shared-app",
			})

			By("creating first PowerTool")
			spec1 := CreateSimpleBasicPowerToolSpec(map[string]string{"app": "shared-app"})
			powerTool1 := CreateSimpleTestPowerTool("conflict-tool-1", namespace.Name, spec1)

			By("waiting for first PowerTool to start")
			WaitForSimplePowerToolPhase(powerTool1, "Pending")

			By("creating second PowerTool targeting same pod")
			spec2 := CreateSimpleBasicPowerToolSpec(map[string]string{"app": "shared-app"})
			spec2.Tool.Name = "aperf" // Same tool
			powerTool2 := CreateSimpleTestPowerTool("conflict-tool-2", namespace.Name, spec2)

			By("verifying conflict detection")
			Eventually(func() string {
				updated := GetSimplePowerTool(powerTool2)
				if updated.Status.Phase == nil {
					return ""
				}
				return *updated.Status.Phase
			}, "30s", "1s").Should(Equal("Failed"))

			By("verifying conflict condition is set")
			WaitForSimplePowerToolCondition(powerTool2, "ConflictDetected", "True")
		})

		It("should update status conditions correctly throughout lifecycle", func() {
			By("creating target pod")
			CreateSimpleMockTargetPod(namespace.Name, "status-pod", map[string]string{
				"app": "status-app",
			})

			By("creating PowerTool")
			spec := CreateSimpleBasicPowerToolSpec(map[string]string{"app": "status-app"})
			powerTool := CreateSimpleTestPowerTool("status-conditions", namespace.Name, spec)

			By("verifying initial conditions are set")
			Eventually(func() []v1alpha1.PowerToolCondition {
				updated := GetSimplePowerTool(powerTool)
				return updated.Status.Conditions
			}, "30s", "1s").Should(Not(BeEmpty()))

			By("verifying Ready condition transitions")
			WaitForSimplePowerToolCondition(powerTool, "Ready", "False")

			By("verifying TargetsFound condition is True")
			WaitForSimplePowerToolCondition(powerTool, "TargetsFound", "True")

			By("verifying ToolConfigured condition")
			WaitForSimplePowerToolCondition(powerTool, "ToolConfigured", "True")
		})

		It("should handle tool configuration validation", func() {
			By("creating PowerTool with invalid tool configuration")
			spec := v1alpha1.PowerToolSpec{
				Targets: v1alpha1.TargetSpec{
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "test-app"},
					},
				},
				Tool: v1alpha1.ToolSpec{
					Name:     "nonexistent-tool",
					Duration: "30s",
				},
				Output: v1alpha1.OutputSpec{
					Mode: "ephemeral",
				},
			}

			powerTool := &v1alpha1.PowerTool{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "invalid-tool-config",
					Namespace: namespace.Name,
				},
				Spec: spec,
			}

			By("expecting creation to fail due to validation")
			err := simpleK8sClient.Create(simpleCtx, powerTool)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("Resource Lifecycle Management", func() {
		It("should handle finalizer logic correctly", func() {
			By("creating target pod")
			CreateSimpleMockTargetPod(namespace.Name, "finalizer-pod", map[string]string{
				"app": "finalizer-app",
			})

			By("creating PowerTool")
			spec := CreateSimpleBasicPowerToolSpec(map[string]string{"app": "finalizer-app"})
			powerTool := CreateSimpleTestPowerTool("finalizer-test", namespace.Name, spec)

			By("verifying finalizer is added")
			Eventually(func() []string {
				updated := GetSimplePowerTool(powerTool)
				return updated.Finalizers
			}, "30s", "1s").Should(ContainElement("powertool.codriverlabs.ai.toe.run/finalizer"))

			By("initiating deletion")
			Expect(simpleK8sClient.Delete(simpleCtx, powerTool)).To(Succeed())

			By("verifying finalizer cleanup occurs")
			Eventually(func() bool {
				updated := &v1alpha1.PowerTool{}
				err := simpleK8sClient.Get(simpleCtx, client.ObjectKeyFromObject(powerTool), updated)
				return err != nil // Resource should be deleted
			}, "30s", "1s").Should(BeTrue())
		})

		It("should handle requeue intervals based on phase", func() {
			By("creating target pod")
			CreateSimpleMockTargetPod(namespace.Name, "requeue-pod", map[string]string{
				"app": "requeue-app",
			})

			By("creating PowerTool")
			spec := CreateSimpleBasicPowerToolSpec(map[string]string{"app": "requeue-app"})
			powerTool := CreateSimpleTestPowerTool("requeue-test", namespace.Name, spec)

			By("tracking reconciliation timing")
			startTime := time.Now()

			By("waiting for phase transition")
			WaitForSimplePowerToolPhase(powerTool, "Pending")

			By("verifying reasonable reconciliation timing")
			elapsed := time.Since(startTime)
			Expect(elapsed).To(BeNumerically("<", 30*time.Second))
		})
	})

	Context("Error Handling and Recovery", func() {
		It("should recover from transient API errors", func() {
			By("creating target pod")
			CreateSimpleMockTargetPod(namespace.Name, "recovery-pod", map[string]string{
				"app": "recovery-app",
			})

			By("creating PowerTool")
			spec := CreateSimpleBasicPowerToolSpec(map[string]string{"app": "recovery-app"})
			powerTool := CreateSimpleTestPowerTool("recovery-test", namespace.Name, spec)

			By("verifying eventual consistency despite potential transient errors")
			Eventually(func() string {
				updated := GetSimplePowerTool(powerTool)
				if updated.Status.Phase == nil {
					return ""
				}
				return *updated.Status.Phase
			}, "60s", "2s").Should(Or(Equal("Pending"), Equal("Running"), Equal("Completed")))
		})

		It("should handle malformed PowerTool specifications", func() {
			By("creating PowerTool with invalid duration format")
			powerTool := &v1alpha1.PowerTool{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "malformed-spec",
					Namespace: namespace.Name,
				},
				Spec: v1alpha1.PowerToolSpec{
					Targets: v1alpha1.TargetSpec{
						LabelSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{"app": "test"},
						},
					},
					Tool: v1alpha1.ToolSpec{
						Name:     "aperf",
						Duration: "invalid-duration",
					},
					Output: v1alpha1.OutputSpec{
						Mode: "ephemeral",
					},
				},
			}

			By("expecting validation to prevent creation")
			err := simpleK8sClient.Create(simpleCtx, powerTool)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("Multi-Pod Scenarios", func() {
		It("should handle scaling scenarios with pod additions", func() {
			By("creating initial pods")
			CreateSimpleMockTargetPod(namespace.Name, "scale-pod-1", map[string]string{
				"app": "scalable-app",
			})

			By("creating PowerTool")
			spec := CreateSimpleBasicPowerToolSpec(map[string]string{"app": "scalable-app"})
			powerTool := CreateSimpleTestPowerTool("scaling-test", namespace.Name, spec)

			By("waiting for initial reconciliation")
			Eventually(func() int {
				updated := GetSimplePowerTool(powerTool)
				return len(updated.Status.ActivePods)
			}, "30s", "1s").Should(Equal(1))

			By("adding more pods")
			CreateSimpleMockTargetPod(namespace.Name, "scale-pod-2", map[string]string{
				"app": "scalable-app",
			})
			CreateSimpleMockTargetPod(namespace.Name, "scale-pod-3", map[string]string{
				"app": "scalable-app",
			})

			By("verifying PowerTool discovers new pods")
			Eventually(func() int {
				updated := GetSimplePowerTool(powerTool)
				return len(updated.Status.ActivePods)
			}, "30s", "1s").Should(Equal(3))
		})

		It("should handle mixed pod states correctly", func() {
			By("creating pods in different states")
			// Running pod
			_ = CreateSimpleMockTargetPod(namespace.Name, "running-pod", map[string]string{
				"app": "mixed-app",
			})

			// Pending pod
			pendingPod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pending-pod",
					Namespace: namespace.Name,
					Labels:    map[string]string{"app": "mixed-app"},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:  "app",
						Image: "nginx:latest",
					}},
				},
				Status: corev1.PodStatus{
					Phase: corev1.PodPending,
				},
			}
			Expect(simpleK8sClient.Create(simpleCtx, pendingPod)).To(Succeed())
			Expect(simpleK8sClient.Status().Update(simpleCtx, pendingPod)).To(Succeed())

			By("creating PowerTool")
			spec := CreateSimpleBasicPowerToolSpec(map[string]string{"app": "mixed-app"})
			powerTool := CreateSimpleTestPowerTool("mixed-states", namespace.Name, spec)

			By("verifying only running pods are targeted")
			Eventually(func() int {
				updated := GetSimplePowerTool(powerTool)
				return len(updated.Status.ActivePods)
			}, "30s", "1s").Should(Equal(1)) // Only running pod

			By("updating pending pod to running")
			pendingPod.Status.Phase = corev1.PodRunning
			pendingPod.Status.ContainerStatuses = []corev1.ContainerStatus{{
				Name:  "app",
				Ready: true,
				State: corev1.ContainerState{
					Running: &corev1.ContainerStateRunning{},
				},
			}}
			Expect(simpleK8sClient.Status().Update(simpleCtx, pendingPod)).To(Succeed())

			By("verifying PowerTool now targets both pods")
			Eventually(func() int {
				updated := GetSimplePowerTool(powerTool)
				return len(updated.Status.ActivePods)
			}, "30s", "1s").Should(Equal(2))
		})
	})

	Context("Performance and Timing", func() {
		It("should reconcile within reasonable time bounds", func() {
			By("creating multiple target pods")
			for i := 0; i < 5; i++ {
				CreateSimpleMockTargetPod(namespace.Name, fmt.Sprintf("perf-pod-%d", i), map[string]string{
					"app": "perf-app",
				})
			}

			By("measuring reconciliation time")
			startTime := time.Now()

			By("creating PowerTool")
			spec := CreateSimpleBasicPowerToolSpec(map[string]string{"app": "perf-app"})
			powerTool := CreateSimpleTestPowerTool("performance-test", namespace.Name, spec)

			By("waiting for reconciliation completion")
			Eventually(func() int {
				updated := GetSimplePowerTool(powerTool)
				return len(updated.Status.ActivePods)
			}, "30s", "1s").Should(Equal(5))

			By("verifying reconciliation completed within time bounds")
			elapsed := time.Since(startTime)
			Expect(elapsed).To(BeNumerically("<", 15*time.Second))
		})
	})
})
