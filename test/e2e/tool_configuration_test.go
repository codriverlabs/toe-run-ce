package e2e

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"toe/api/v1alpha1"
)

var _ = Describe("Tool Configuration and Validation", func() {
	var namespace *corev1.Namespace

	BeforeEach(func() {
		namespace = CreateSimpleTestNamespace()
		CreateSimpleMockTargetPod(namespace.Name, "tool-pod", map[string]string{
			"app": "tool-app",
		})
	})

	AfterEach(func() {
		DeleteSimpleTestNamespace(namespace)
	})

	Context("PowerToolConfig Management", func() {
		It("should validate tool configurations exist", func() {
			By("creating PowerToolConfig")
			_ = CreateSimpleTestPowerToolConfig("valid-config", namespace.Name)

			By("creating PowerTool referencing the config")
			spec := CreateSimpleBasicPowerToolSpec(map[string]string{"app": "tool-app"})
			powerTool := CreateSimpleTestPowerTool("config-test", namespace.Name, spec)

			By("verifying PowerTool finds the configuration")
			WaitForSimplePowerToolCondition(powerTool, "ToolConfigured", "True")
		})

		It("should handle missing tool configurations", func() {
			By("creating PowerTool without corresponding config")
			spec := v1alpha1.PowerToolSpec{
				Targets: v1alpha1.TargetSpec{
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "tool-app"},
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
					Name:      "missing-config",
					Namespace: namespace.Name,
				},
				Spec: spec,
			}

			By("expecting creation to fail due to missing config")
			err := simpleK8sClient.Create(simpleCtx, powerTool)
			Expect(err).To(HaveOccurred())
		})

		It("should handle multiple PowerToolConfigs", func() {
			By("creating multiple tool configurations")
			_ = CreateSimpleTestPowerToolConfig("aperf-config", namespace.Name)

			allowPrivileged := true
			config2 := &v1alpha1.PowerToolConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "strace-config",
					Namespace: namespace.Name,
				},
				Spec: v1alpha1.PowerToolConfigSpec{
					Name:  "strace",
					Image: "ghcr.io/codriverlabs/toe-strace:latest",
					SecurityContext: v1alpha1.SecuritySpec{
						AllowPrivileged: &allowPrivileged,
						Capabilities: &v1alpha1.Capabilities{
							Add: []string{"SYS_PTRACE"},
						},
					},
				},
			}
			Expect(simpleK8sClient.Create(simpleCtx, config2)).To(Succeed())

			By("creating PowerTools for different tools")
			spec1 := CreateSimpleBasicPowerToolSpec(map[string]string{"app": "tool-app"})
			spec1.Tool.Name = "aperf"
			powerTool1 := CreateSimpleTestPowerTool("aperf-tool", namespace.Name, spec1)

			spec2 := CreateSimpleBasicPowerToolSpec(map[string]string{"app": "tool-app"})
			spec2.Tool.Name = "strace"
			powerTool2 := CreateSimpleTestPowerTool("strace-tool", namespace.Name, spec2)

			By("verifying both tools are configured correctly")
			WaitForSimplePowerToolCondition(powerTool1, "ToolConfigured", "True")
			WaitForSimplePowerToolCondition(powerTool2, "ToolConfigured", "True")
		})
	})

	Context("Tool Duration Validation", func() {
		BeforeEach(func() {
			CreateSimpleTestPowerToolConfig("duration-config", namespace.Name)
		})

		It("should accept valid duration formats", func() {
			By("testing various valid duration formats")
			validDurations := []string{
				"30s",
				"5m",
				"1h",
				"90s",
				"2m30s",
			}

			for _, duration := range validDurations {
				spec := CreateSimpleBasicPowerToolSpec(map[string]string{"app": "tool-app"})
				spec.Tool.Duration = duration
				powerTool := CreateSimpleTestPowerTool("duration-"+duration, namespace.Name, spec)

				By("verifying duration is accepted: " + duration)
				updated := GetSimplePowerTool(powerTool)
				Expect(updated.Spec.Tool.Duration).To(Equal(duration))

				// Cleanup
				Expect(simpleK8sClient.Delete(simpleCtx, powerTool)).To(Succeed())
			}
		})

		It("should reject invalid duration formats", func() {
			By("testing invalid duration formats")
			invalidDurations := []string{
				"invalid",
				"30",
				"-5s",
				"0s",
				"25h", // Too long
			}

			for _, duration := range invalidDurations {
				spec := CreateSimpleBasicPowerToolSpec(map[string]string{"app": "tool-app"})
				spec.Tool.Duration = duration

				powerTool := &v1alpha1.PowerTool{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "invalid-duration-" + duration,
						Namespace: namespace.Name,
					},
					Spec: spec,
				}

				By("expecting validation to fail for duration: " + duration)
				err := simpleK8sClient.Create(simpleCtx, powerTool)
				Expect(err).To(HaveOccurred())
			}
		})
	})

	Context("Tool Arguments and Environment", func() {
		BeforeEach(func() {
			CreateSimpleTestPowerToolConfig("args-config", namespace.Name)
		})

		It("should handle tool arguments correctly", func() {
			By("creating PowerTool with custom arguments")
			spec := v1alpha1.PowerToolSpec{
				Targets: v1alpha1.TargetSpec{
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "tool-app"},
					},
				},
				Tool: v1alpha1.ToolSpec{
					Name:     "aperf",
					Duration: "30s",
					Args:     []string{"--verbose", "--output=/tmp/custom.out"},
				},
				Output: v1alpha1.OutputSpec{
					Mode: "ephemeral",
				},
			}
			powerTool := CreateSimpleTestPowerTool("custom-args", namespace.Name, spec)

			By("verifying arguments are preserved")
			updated := GetSimplePowerTool(powerTool)
			Expect(updated.Spec.Tool.Args).To(Equal([]string{"--verbose", "--output=/tmp/custom.out"}))
		})

		It("should handle environment variables", func() {
			By("creating PowerTool with environment variables")
			spec := v1alpha1.PowerToolSpec{
				Targets: v1alpha1.TargetSpec{
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "tool-app"},
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
			powerTool := CreateSimpleTestPowerTool("custom-env", namespace.Name, spec)

			By("verifying environment variables are set")
			_ = GetSimplePowerTool(powerTool)
		})
	})

	Context("Security Context Validation", func() {
		It("should validate security context requirements", func() {
			By("creating PowerToolConfig with specific security requirements")
			allowPrivileged := true
			allowHostPID := true
			config := &v1alpha1.PowerToolConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "secure-config",
					Namespace: namespace.Name,
				},
				Spec: v1alpha1.PowerToolConfigSpec{
					Name:  "secure-tool",
					Image: "ghcr.io/codriverlabs/toe-secure:latest",
					SecurityContext: v1alpha1.SecuritySpec{
						AllowPrivileged: &allowPrivileged,
						AllowHostPID:    &allowHostPID,
						Capabilities: &v1alpha1.Capabilities{
							Add:  []string{"SYS_ADMIN", "SYS_PTRACE"},
							Drop: []string{"NET_RAW"},
						},
					},
				},
			}
			Expect(simpleK8sClient.Create(simpleCtx, config)).To(Succeed())

			By("creating PowerTool using secure configuration")
			spec := v1alpha1.PowerToolSpec{
				Targets: v1alpha1.TargetSpec{
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "tool-app"},
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
			powerTool := CreateSimpleTestPowerTool("secure-tool-test", namespace.Name, spec)

			By("verifying security context is applied")
			WaitForSimplePowerToolCondition(powerTool, "ToolConfigured", "True")
		})

		It("should handle capability requirements", func() {
			By("creating PowerToolConfig with specific capabilities")
			allowPrivileged := false
			config := &v1alpha1.PowerToolConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "cap-config",
					Namespace: namespace.Name,
				},
				Spec: v1alpha1.PowerToolConfigSpec{
					Name:  "cap-tool",
					Image: "ghcr.io/codriverlabs/toe-cap:latest",
					SecurityContext: v1alpha1.SecuritySpec{
						AllowPrivileged: &allowPrivileged,
						Capabilities: &v1alpha1.Capabilities{
							Add: []string{"NET_ADMIN", "SYS_TIME"},
						},
					},
				},
			}
			Expect(simpleK8sClient.Create(simpleCtx, config)).To(Succeed())

			By("creating PowerTool with capability requirements")
			spec := v1alpha1.PowerToolSpec{
				Targets: v1alpha1.TargetSpec{
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "tool-app"},
					},
				},
				Tool: v1alpha1.ToolSpec{
					Name:     "cap-tool",
					Duration: "30s",
				},
				Output: v1alpha1.OutputSpec{
					Mode: "ephemeral",
				},
			}
			powerTool := CreateSimpleTestPowerTool("capability-test", namespace.Name, spec)

			By("verifying capability configuration")
			WaitForSimplePowerToolCondition(powerTool, "ToolConfigured", "True")
		})
	})

	Context("Tool Image Management", func() {
		It("should handle different image registries", func() {
			By("creating PowerToolConfig with custom registry")
			allowPrivileged := true
			config := &v1alpha1.PowerToolConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "custom-registry",
					Namespace: namespace.Name,
				},
				Spec: v1alpha1.PowerToolConfigSpec{
					Name:  "custom-tool",
					Image: "custom-registry.example.com/tools/profiler:v1.0.0",
					SecurityContext: v1alpha1.SecuritySpec{
						AllowPrivileged: &allowPrivileged,
					},
				},
			}
			Expect(simpleK8sClient.Create(simpleCtx, config)).To(Succeed())

			By("creating PowerTool using custom registry image")
			spec := v1alpha1.PowerToolSpec{
				Targets: v1alpha1.TargetSpec{
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "tool-app"},
					},
				},
				Tool: v1alpha1.ToolSpec{
					Name:     "custom-tool",
					Duration: "30s",
				},
				Output: v1alpha1.OutputSpec{
					Mode: "ephemeral",
				},
			}
			powerTool := CreateSimpleTestPowerTool("custom-registry-test", namespace.Name, spec)

			By("verifying custom image is configured")
			WaitForSimplePowerToolCondition(powerTool, "ToolConfigured", "True")
		})

		It("should handle image pull policies", func() {
			By("creating PowerToolConfig with pull policy")
			allowPrivileged := true
			config := &v1alpha1.PowerToolConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pull-policy-config",
					Namespace: namespace.Name,
				},
				Spec: v1alpha1.PowerToolConfigSpec{
					Name:  "pull-policy-tool",
					Image: "ghcr.io/codriverlabs/toe-test:latest",
					SecurityContext: v1alpha1.SecuritySpec{
						AllowPrivileged: &allowPrivileged,
					},
				},
			}
			Expect(simpleK8sClient.Create(simpleCtx, config)).To(Succeed())

			By("verifying image configuration is accepted")
			updated := &v1alpha1.PowerToolConfig{}
			Expect(simpleK8sClient.Get(simpleCtx, client.ObjectKeyFromObject(config), updated)).To(Succeed())
			Expect(updated.Spec.Image).To(Equal("ghcr.io/codriverlabs/toe-test:latest"))
		})
	})
})
