package e2e

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"toe/api/v1alpha1"
)

var _ = Describe("PowerTool Lifecycle", func() {
	var namespace *corev1.Namespace

	BeforeEach(func() {
		namespace = CreateSimpleTestNamespace()
		CreateSimpleMockTargetPod(namespace.Name, "target-pod", map[string]string{
			"app": "test-app",
			"env": "testing",
		})
		CreateSimpleTestPowerToolConfig("aperf-config", namespace.Name)
	})

	AfterEach(func() {
		DeleteSimpleTestNamespace(namespace)
	})

	Context("PowerTool Creation", func() {
		It("should create PowerTool with valid spec", func() {
			By("creating a PowerTool with basic configuration")
			spec := CreateSimpleBasicPowerToolSpec(map[string]string{"app": "test-app"})
			powerTool := CreateSimpleTestPowerTool("test-powertool", namespace.Name, spec)

			By("verifying PowerTool is created successfully")
			Expect(powerTool.Name).To(Equal("test-powertool"))
			Expect(powerTool.Namespace).To(Equal(namespace.Name))
			Expect(powerTool.Spec.Tool.Name).To(Equal("aperf"))

			By("waiting for PowerTool to be processed")
			WaitForSimplePowerToolPhase(powerTool, "Pending")

			By("verifying status conditions are set")
			updated := GetSimplePowerTool(powerTool)
			Expect(updated.Status.Conditions).NotTo(BeEmpty())
		})

		It("should handle PowerTool with multiple target pods", func() {
			By("creating additional target pods")
			CreateSimpleMockTargetPod(namespace.Name, "target-pod-2", map[string]string{
				"app": "test-app",
				"env": "testing",
			})
			CreateSimpleMockTargetPod(namespace.Name, "target-pod-3", map[string]string{
				"app": "test-app",
				"env": "production",
			})

			By("creating PowerTool targeting multiple pods")
			spec := CreateSimpleBasicPowerToolSpec(map[string]string{"app": "test-app"})
			powerTool := CreateSimpleTestPowerTool("multi-target", namespace.Name, spec)

			By("verifying PowerTool processes multiple targets")
			WaitForSimplePowerToolPhase(powerTool, "Pending")

			updated := GetSimplePowerTool(powerTool)
			Expect(updated.Status.ActivePods).NotTo(BeEmpty())
		})

		It("should set appropriate finalizers", func() {
			By("creating a PowerTool")
			spec := CreateSimpleBasicPowerToolSpec(map[string]string{"app": "test-app"})
			powerTool := CreateSimpleTestPowerTool("finalizer-test", namespace.Name, spec)

			By("verifying finalizer is set")
			updated := GetSimplePowerTool(powerTool)
			Expect(updated.Finalizers).To(ContainElement("powertool.codriverlabs.ai.toe.run/finalizer"))
		})
	})

	Context("PowerTool Validation", func() {
		It("should reject PowerTool with invalid tool name", func() {
			By("attempting to create PowerTool with nonexistent tool")
			spec := v1alpha1.PowerToolSpec{
				Targets: v1alpha1.TargetSpec{
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "test-app"},
					},
				},
				Tool: v1alpha1.ToolSpec{
					Name: "nonexistent-tool",
				},
			}

			powerTool := &v1alpha1.PowerTool{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "invalid-tool",
					Namespace: namespace.Name,
				},
				Spec: spec,
			}

			By("expecting creation to fail")
			err := simpleK8sClient.Create(simpleCtx, powerTool)
			Expect(err).To(HaveOccurred())
		})

		It("should reject PowerTool with invalid duration", func() {
			By("attempting to create PowerTool with invalid duration")
			spec := CreateSimpleBasicPowerToolSpec(map[string]string{"app": "test-app"})
			spec.Tool.Duration = "invalid-duration"

			powerTool := &v1alpha1.PowerTool{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "invalid-duration",
					Namespace: namespace.Name,
				},
				Spec: spec,
			}

			By("expecting creation to fail")
			err := simpleK8sClient.Create(simpleCtx, powerTool)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("PowerTool Status Updates", func() {
		It("should update status phase correctly", func() {
			By("creating a PowerTool")
			spec := CreateSimpleBasicPowerToolSpec(map[string]string{"app": "test-app"})
			powerTool := CreateSimpleTestPowerTool("status-test", namespace.Name, spec)

			By("verifying initial status")
			Eventually(func() string {
				updated := GetSimplePowerTool(powerTool)
				if updated.Status.Phase == nil {
					return ""
				}
				return *updated.Status.Phase
			}).Should(Equal("Pending"))

			By("verifying status conditions are populated")
			updated := GetSimplePowerTool(powerTool)
			Expect(updated.Status.Conditions).NotTo(BeEmpty())
		})

		It("should track target pods in status", func() {
			By("creating a PowerTool")
			spec := CreateSimpleBasicPowerToolSpec(map[string]string{"app": "test-app"})
			powerTool := CreateSimpleTestPowerTool("target-tracking", namespace.Name, spec)

			By("verifying target pods are tracked")
			Eventually(func() map[string]string {
				updated := GetSimplePowerTool(powerTool)
				return updated.Status.ActivePods
			}).Should(Not(BeEmpty()))
		})
	})

	Context("PowerTool Deletion", func() {
		It("should handle deletion gracefully", func() {
			By("creating a PowerTool")
			spec := CreateSimpleBasicPowerToolSpec(map[string]string{"app": "test-app"})
			powerTool := CreateSimpleTestPowerTool("deletion-test", namespace.Name, spec)

			By("waiting for PowerTool to be processed")
			WaitForSimplePowerToolPhase(powerTool, "Pending")

			By("deleting the PowerTool")
			Expect(simpleK8sClient.Delete(simpleCtx, powerTool)).To(Succeed())

			By("verifying PowerTool is deleted")
			Eventually(func() bool {
				updated := &v1alpha1.PowerTool{}
				err := simpleK8sClient.Get(simpleCtx, client.ObjectKeyFromObject(powerTool), updated)
				return err != nil
			}).Should(BeTrue())
		})
	})
})
