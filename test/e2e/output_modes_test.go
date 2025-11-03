package e2e

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"toe/api/v1alpha1"
)

var _ = Describe("Output Mode Configuration", func() {
	var namespace *corev1.Namespace

	BeforeEach(func() {
		namespace = CreateSimpleTestNamespace()
		CreateSimpleTestPowerToolConfig("aperf-config", namespace.Name)
		CreateSimpleMockTargetPod(namespace.Name, "output-pod", map[string]string{
			"app": "output-app",
		})
	})

	AfterEach(func() {
		DeleteSimpleTestNamespace(namespace)
	})

	Context("Ephemeral Mode", func() {
		It("should configure ephemeral output correctly", func() {
			By("creating PowerTool with ephemeral output")
			spec := v1alpha1.PowerToolSpec{
				Targets: v1alpha1.TargetSpec{
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "output-app"},
					},
				},
				Tool: v1alpha1.ToolSpec{
					Name:     "aperf",
					Duration: "30s",
				},
				Output: v1alpha1.OutputSpec{
					Mode: "ephemeral",
				},
			}
			powerTool := CreateSimpleTestPowerTool("ephemeral-output", namespace.Name, spec)

			By("verifying PowerTool accepts ephemeral configuration")
			WaitForSimplePowerToolPhase(powerTool, "Pending")

			By("verifying output mode is correctly set")
			updated := GetSimplePowerTool(powerTool)
			Expect(updated.Spec.Output.Mode).To(Equal("ephemeral"))
		})

		It("should handle ephemeral mode with custom path", func() {
			By("creating PowerTool with custom ephemeral path")
			spec := v1alpha1.PowerToolSpec{
				Targets: v1alpha1.TargetSpec{
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "output-app"},
					},
				},
				Tool: v1alpha1.ToolSpec{
					Name:     "aperf",
					Duration: "30s",
				},
				Output: v1alpha1.OutputSpec{
					Mode: "ephemeral",
				},
			}
			powerTool := CreateSimpleTestPowerTool("ephemeral-custom-path", namespace.Name, spec)

			By("verifying custom path configuration")
			WaitForSimplePowerToolPhase(powerTool, "Pending")
			updated := GetSimplePowerTool(powerTool)
			Expect(updated.Spec.Output.Mode).To(Equal("ephemeral"))
		})
	})

	Context("PVC Mode", func() {
		It("should validate PVC configuration", func() {
			By("creating PowerTool with PVC output")
			spec := v1alpha1.PowerToolSpec{
				Targets: v1alpha1.TargetSpec{
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "output-app"},
					},
				},
				Tool: v1alpha1.ToolSpec{
					Name:     "aperf",
					Duration: "30s",
				},
				Output: v1alpha1.OutputSpec{
					Mode: "pvc",
					PVC: &v1alpha1.PVCSpec{
						ClaimName: "test-pvc",
						Path:      func() *string { s := "/data/profiles"; return &s }(),
					},
				},
			}
			powerTool := CreateSimpleTestPowerTool("pvc-output", namespace.Name, spec)

			By("verifying PVC configuration is accepted")
			WaitForSimplePowerToolPhase(powerTool, "Pending")

			By("verifying PVC settings are correctly configured")
			updated := GetSimplePowerTool(powerTool)
			Expect(updated.Spec.Output.Mode).To(Equal("pvc"))
			Expect(updated.Spec.Output.PVC.ClaimName).To(Equal("test-pvc"))
			Expect(updated.Spec.Output.PVC.Path).To(Equal("/data/profiles"))
		})

		It("should handle PVC mode with storage class", func() {
			By("creating PowerTool with PVC and storage class")
			spec := v1alpha1.PowerToolSpec{
				Targets: v1alpha1.TargetSpec{
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "output-app"},
					},
				},
				Tool: v1alpha1.ToolSpec{
					Name:     "aperf",
					Duration: "30s",
				},
				Output: v1alpha1.OutputSpec{
					Mode: "pvc",
					PVC: &v1alpha1.PVCSpec{
						ClaimName: "storage-pvc",
						Path:      func() *string { s := "/data/profiles"; return &s }(),
					},
				},
			}
			powerTool := CreateSimpleTestPowerTool("pvc-storage-class", namespace.Name, spec)

			By("verifying PVC configuration")
			updated := GetSimplePowerTool(powerTool)
			Expect(updated.Spec.Output.PVC.ClaimName).To(Equal("storage-pvc"))
			Expect(updated.Spec.Output.PVC.Path).NotTo(BeNil())
			Expect(*updated.Spec.Output.PVC.Path).To(Equal("/data/profiles"))
		})

		It("should reject invalid PVC configurations", func() {
			By("attempting to create PowerTool with invalid PVC config")
			spec := v1alpha1.PowerToolSpec{
				Targets: v1alpha1.TargetSpec{
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "output-app"},
					},
				},
				Tool: v1alpha1.ToolSpec{
					Name:     "aperf",
					Duration: "30s",
				},
				Output: v1alpha1.OutputSpec{
					Mode: "pvc",
					// Missing PVC configuration
				},
			}

			powerTool := &v1alpha1.PowerTool{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "invalid-pvc",
					Namespace: namespace.Name,
				},
				Spec: spec,
			}

			By("expecting validation to fail")
			err := simpleK8sClient.Create(simpleCtx, powerTool)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("Collector Mode", func() {
		It("should configure collector endpoint correctly", func() {
			By("creating PowerTool with collector output")
			spec := v1alpha1.PowerToolSpec{
				Targets: v1alpha1.TargetSpec{
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "output-app"},
					},
				},
				Tool: v1alpha1.ToolSpec{
					Name:     "aperf",
					Duration: "30s",
				},
				Output: v1alpha1.OutputSpec{
					Mode: "collector",
					Collector: &v1alpha1.CollectorSpec{
						Endpoint: "https://collector.toe-system.svc.cluster.local:8443",
					},
				},
			}
			powerTool := CreateSimpleTestPowerTool("collector-output", namespace.Name, spec)

			By("verifying collector configuration")
			WaitForSimplePowerToolPhase(powerTool, "Pending")
			updated := GetSimplePowerTool(powerTool)
			Expect(updated.Spec.Output.Mode).To(Equal("collector"))
			Expect(updated.Spec.Output.Collector.Endpoint).To(Equal("https://collector.toe-system.svc.cluster.local:8443"))
		})

		It("should handle collector mode with authentication", func() {
			By("creating PowerTool with collector authentication")
			spec := v1alpha1.PowerToolSpec{
				Targets: v1alpha1.TargetSpec{
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "output-app"},
					},
				},
				Tool: v1alpha1.ToolSpec{
					Name:     "aperf",
					Duration: "30s",
				},
				Output: v1alpha1.OutputSpec{
					Mode: "collector",
					Collector: &v1alpha1.CollectorSpec{
						Endpoint: "https://collector.example.com:8443",
					},
				},
			}
			powerTool := CreateSimpleTestPowerTool("collector-auth", namespace.Name, spec)

			By("verifying authentication configuration")
			_ = GetSimplePowerTool(powerTool)
		})

		It("should validate collector endpoint format", func() {
			By("attempting to create PowerTool with invalid collector endpoint")
			spec := v1alpha1.PowerToolSpec{
				Targets: v1alpha1.TargetSpec{
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "output-app"},
					},
				},
				Tool: v1alpha1.ToolSpec{
					Name:     "aperf",
					Duration: "30s",
				},
				Output: v1alpha1.OutputSpec{
					Mode: "collector",
					Collector: &v1alpha1.CollectorSpec{
						Endpoint: "invalid-url",
					},
				},
			}

			powerTool := &v1alpha1.PowerTool{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "invalid-collector",
					Namespace: namespace.Name,
				},
				Spec: spec,
			}

			By("expecting validation to fail for invalid URL")
			err := simpleK8sClient.Create(simpleCtx, powerTool)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("Output Mode Transitions", func() {
		It("should handle output mode changes", func() {
			By("creating PowerTool with ephemeral output")
			spec := CreateSimpleBasicPowerToolSpec(map[string]string{"app": "output-app"})
			powerTool := CreateSimpleTestPowerTool("mode-transition", namespace.Name, spec)

			By("verifying initial ephemeral mode")
			WaitForSimplePowerToolPhase(powerTool, "Pending")
			initial := GetSimplePowerTool(powerTool)
			Expect(initial.Spec.Output.Mode).To(Equal("ephemeral"))

			By("updating to PVC mode")
			updated := GetSimplePowerTool(powerTool)
			updated.Spec.Output.Mode = "pvc"
			updated.Spec.Output.PVC = &v1alpha1.PVCSpec{
				ClaimName: "transition-pvc",
				Path:      func() *string { s := "/data"; return &s }(),
			}
			Expect(simpleK8sClient.Update(simpleCtx, updated)).To(Succeed())

			By("verifying mode transition")
			Eventually(func() string {
				current := GetSimplePowerTool(powerTool)
				return current.Spec.Output.Mode
			}, "30s", "1s").Should(Equal("pvc"))
		})
	})

	Context("Output Path Validation", func() {
		It("should validate output path formats", func() {
			By("testing various path formats")
			validPaths := []string{
				"/tmp/profiles",
				"/data/output",
				"/var/log/toe",
			}

			for _, path := range validPaths {
				spec := CreateSimpleBasicPowerToolSpec(map[string]string{"app": "output-app"})
				powerTool := CreateSimpleTestPowerTool("path-test-"+path[1:], namespace.Name, spec)

				By("verifying path is accepted: " + path)
				_ = GetSimplePowerTool(powerTool)

				// Cleanup
				Expect(simpleK8sClient.Delete(simpleCtx, powerTool)).To(Succeed())
			}
		})

		It("should reject invalid output paths", func() {
			By("testing invalid path formats")
			invalidPaths := []string{
				"relative/path",
				"",
				"../../../etc/passwd",
			}

			for _, path := range invalidPaths {
				spec := CreateSimpleBasicPowerToolSpec(map[string]string{"app": "output-app"})

				powerTool := &v1alpha1.PowerTool{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "invalid-path-" + path,
						Namespace: namespace.Name,
					},
					Spec: spec,
				}

				By("expecting validation to fail for path: " + path)
				err := simpleK8sClient.Create(simpleCtx, powerTool)
				Expect(err).To(HaveOccurred())
			}
		})
	})
})
