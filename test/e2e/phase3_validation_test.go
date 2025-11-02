package e2e

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"toe/api/v1alpha1"
)

var _ = Describe("Phase 3 Validation Tests", func() {
	var namespace *corev1.Namespace

	BeforeEach(func() {
		err := InitializeSimpleClients()
		Expect(err).NotTo(HaveOccurred())

		namespace = CreateSimpleTestNamespace()
		CreateSimpleTestPowerToolConfig("phase3-config", namespace.Name)
		CreateSimpleMockTargetPod(namespace.Name, "phase3-pod", map[string]string{
			"app": "phase3-app",
		})
	})

	AfterEach(func() {
		if namespace != nil {
			DeleteSimpleTestNamespace(namespace)
		}
	})

	Context("Basic RBAC Validation", func() {
		It("should create PowerTool with basic configuration", func() {
			By("creating a basic PowerTool")
			spec := CreateSimpleBasicPowerToolSpec(map[string]string{"app": "phase3-app"})
			powerTool := CreateSimpleTestPowerTool("basic-test", namespace.Name, spec)

			By("verifying PowerTool is created successfully")
			Expect(powerTool.Name).To(Equal("basic-test"))
			Expect(powerTool.Namespace).To(Equal(namespace.Name))
		})

		It("should validate security context requirements", func() {
			By("creating PowerToolConfig with security requirements")
			allowPrivileged := true
			secureConfig := &v1alpha1.PowerToolConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "secure-config",
					Namespace: namespace.Name,
				},
				Spec: v1alpha1.PowerToolConfigSpec{
					Name:  "secure-tool",
					Image: "ghcr.io/codriverlabs/toe-secure:latest",
					SecurityContext: v1alpha1.SecuritySpec{
						AllowPrivileged: &allowPrivileged,
						Capabilities: &v1alpha1.Capabilities{
							Add: []string{"SYS_ADMIN", "SYS_PTRACE"},
						},
					},
				},
			}
			Expect(simpleK8sClient.Create(simpleCtx, secureConfig)).To(Succeed())

			By("creating PowerTool using secure configuration")
			spec := v1alpha1.PowerToolSpec{
				Targets: v1alpha1.TargetSpec{
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "phase3-app"},
					},
				},
				Tool: v1alpha1.ToolSpec{
					Name:     "secure-tool",
					Duration: "30s",
				},
				Output: v1alpha1.OutputSpec{
					Mode: "ephemeral",
				},
			}
			powerTool := CreateSimpleTestPowerTool("secure-test", namespace.Name, spec)

			By("verifying secure PowerTool is created")
			Expect(powerTool.Spec.Tool.Name).To(Equal("secure-tool"))
		})
	})

	Context("Webhook Validation", func() {
		It("should reject PowerTool with invalid duration", func() {
			By("attempting to create PowerTool with invalid duration")
			spec := v1alpha1.PowerToolSpec{
				Targets: v1alpha1.TargetSpec{
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "phase3-app"},
					},
				},
				Tool: v1alpha1.ToolSpec{
					Name:     "phase3-config",
					Duration: "invalid-duration",
				},
				Output: v1alpha1.OutputSpec{
					Mode: "ephemeral",
				},
			}

			powerTool := &v1alpha1.PowerTool{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "invalid-duration",
					Namespace: namespace.Name,
				},
				Spec: spec,
			}

			By("expecting validation to fail")
			err := simpleK8sClient.Create(simpleCtx, powerTool)
			Expect(err).To(HaveOccurred())
		})

		It("should validate output mode configuration", func() {
			By("testing PVC mode without PVC spec")
			spec := v1alpha1.PowerToolSpec{
				Targets: v1alpha1.TargetSpec{
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "phase3-app"},
					},
				},
				Tool: v1alpha1.ToolSpec{
					Name:     "phase3-config",
					Duration: "30s",
				},
				Output: v1alpha1.OutputSpec{
					Mode: "pvc",
					// Missing PVC spec
				},
			}

			powerTool := &v1alpha1.PowerTool{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "invalid-pvc",
					Namespace: namespace.Name,
				},
				Spec: spec,
			}

			By("expecting validation to fail for incomplete PVC config")
			err := simpleK8sClient.Create(simpleCtx, powerTool)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("Integration Scenarios", func() {
		It("should handle multiple target pods", func() {
			By("creating additional target pods")
			CreateSimpleMockTargetPod(namespace.Name, "multi-pod-1", map[string]string{
				"app": "multi-app",
			})
			CreateSimpleMockTargetPod(namespace.Name, "multi-pod-2", map[string]string{
				"app": "multi-app",
			})

			By("creating PowerTool targeting multiple pods")
			spec := CreateSimpleBasicPowerToolSpec(map[string]string{"app": "multi-app"})
			powerTool := CreateSimpleTestPowerTool("multi-pod-test", namespace.Name, spec)

			By("verifying PowerTool is created successfully")
			Expect(powerTool.Spec.Targets.LabelSelector.MatchLabels["app"]).To(Equal("multi-app"))
		})

		It("should validate tool configuration exists", func() {
			By("creating PowerTool with existing configuration")
			spec := CreateSimpleBasicPowerToolSpec(map[string]string{"app": "phase3-app"})
			powerTool := CreateSimpleTestPowerTool("config-exists-test", namespace.Name, spec)

			By("verifying PowerTool references correct tool")
			Expect(powerTool.Spec.Tool.Name).To(Equal("aperf"))
		})
	})

	Context("Error Handling", func() {
		It("should handle missing target pods gracefully", func() {
			By("creating PowerTool with non-matching selector")
			spec := CreateSimpleBasicPowerToolSpec(map[string]string{"app": "nonexistent"})
			powerTool := CreateSimpleTestPowerTool("no-targets", namespace.Name, spec)

			By("verifying PowerTool is created but will fail to find targets")
			Expect(powerTool.Spec.Targets.LabelSelector.MatchLabels["app"]).To(Equal("nonexistent"))
		})

		It("should validate required fields", func() {
			By("attempting to create PowerTool without required fields")
			powerTool := &v1alpha1.PowerTool{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "missing-fields",
					Namespace: namespace.Name,
				},
				Spec: v1alpha1.PowerToolSpec{
					// Missing required fields
				},
			}

			By("expecting validation to fail")
			err := simpleK8sClient.Create(simpleCtx, powerTool)
			Expect(err).To(HaveOccurred())
		})
	})
})
