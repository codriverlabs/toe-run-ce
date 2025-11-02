package e2e

import (
	"context"
	"fmt"

	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"toe/api/v1alpha1"
)

var (
	simpleK8sClient client.Client
	simpleClientset *kubernetes.Clientset
	simpleCtx       = context.Background()
)

// InitializeSimpleClients initializes clients for simple E2E tests
func InitializeSimpleClients() error {
	config, err := ctrl.GetConfig()
	if err != nil {
		return err
	}

	scheme := runtime.NewScheme()
	err = v1alpha1.AddToScheme(scheme)
	if err != nil {
		return err
	}

	simpleK8sClient, err = client.New(config, client.Options{Scheme: scheme})
	if err != nil {
		return err
	}

	simpleClientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	return nil
}

// CreateSimpleTestNamespace creates a unique test namespace
func CreateSimpleTestNamespace() *corev1.Namespace {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "toe-simple-e2e-",
		},
	}
	Expect(simpleK8sClient.Create(simpleCtx, ns)).To(Succeed())
	return ns
}

// DeleteSimpleTestNamespace deletes a test namespace
func DeleteSimpleTestNamespace(ns *corev1.Namespace) {
	Expect(simpleK8sClient.Delete(simpleCtx, ns)).To(Succeed())
}

// CreateSimpleMockTargetPod creates a mock pod for testing
func CreateSimpleMockTargetPod(namespace, name string, labels map[string]string) *corev1.Pod {
	if labels == nil {
		labels = map[string]string{
			"app": "test-app",
			"env": "testing",
		}
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{
				Name:  "app",
				Image: "nginx:latest",
			}},
		},
	}
	Expect(simpleK8sClient.Create(simpleCtx, pod)).To(Succeed())

	// Update status to simulate running pod
	pod.Status = corev1.PodStatus{
		Phase: corev1.PodRunning,
		ContainerStatuses: []corev1.ContainerStatus{{
			Name:  "app",
			Ready: true,
			State: corev1.ContainerState{
				Running: &corev1.ContainerStateRunning{},
			},
		}},
	}
	Expect(simpleK8sClient.Status().Update(simpleCtx, pod)).To(Succeed())
	return pod
}

// CreateSimpleTestPowerTool creates a PowerTool for testing
func CreateSimpleTestPowerTool(name, namespace string, spec v1alpha1.PowerToolSpec) *v1alpha1.PowerTool {
	powerTool := &v1alpha1.PowerTool{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: spec,
	}
	Expect(simpleK8sClient.Create(simpleCtx, powerTool)).To(Succeed())
	return powerTool
}

// CreateSimpleTestPowerToolConfig creates a PowerToolConfig for testing
func CreateSimpleTestPowerToolConfig(name, namespace string) *v1alpha1.PowerToolConfig {
	allowPrivileged := true
	powerToolConfig := &v1alpha1.PowerToolConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1alpha1.PowerToolConfigSpec{
			Name:  "aperf",
			Image: "ghcr.io/codriverlabs/toe-aperf:latest",
			SecurityContext: v1alpha1.SecuritySpec{
				AllowPrivileged: &allowPrivileged,
				Capabilities: &v1alpha1.Capabilities{
					Add: []string{"SYS_ADMIN", "SYS_PTRACE"},
				},
			},
		},
	}
	Expect(simpleK8sClient.Create(simpleCtx, powerToolConfig)).To(Succeed())
	return powerToolConfig
}

// WaitForSimplePowerToolPhase waits for PowerTool to reach expected phase
func WaitForSimplePowerToolPhase(powerTool *v1alpha1.PowerTool, expectedPhase string) {
	Eventually(func() string {
		updated := &v1alpha1.PowerTool{}
		err := simpleK8sClient.Get(simpleCtx, client.ObjectKeyFromObject(powerTool), updated)
		if err != nil {
			return ""
		}
		if updated.Status.Phase == nil {
			return ""
		}
		return *updated.Status.Phase
	}, "30s", "1s").Should(Equal(expectedPhase))
}

// WaitForSimplePowerToolCondition waits for PowerTool to have expected condition
func WaitForSimplePowerToolCondition(powerTool *v1alpha1.PowerTool, conditionType string, status string) {
	Eventually(func() bool {
		updated := &v1alpha1.PowerTool{}
		err := simpleK8sClient.Get(simpleCtx, client.ObjectKeyFromObject(powerTool), updated)
		if err != nil {
			return false
		}

		for _, condition := range updated.Status.Conditions {
			if condition.Type == conditionType && condition.Status == status {
				return true
			}
		}
		return false
	}, "30s", "1s").Should(BeTrue())
}

// GetSimplePowerTool retrieves the latest version of a PowerTool
func GetSimplePowerTool(powerTool *v1alpha1.PowerTool) *v1alpha1.PowerTool {
	updated := &v1alpha1.PowerTool{}
	Expect(simpleK8sClient.Get(simpleCtx, client.ObjectKeyFromObject(powerTool), updated)).To(Succeed())
	return updated
}

// GetSimplePod retrieves the latest version of a Pod
func GetSimplePod(pod *corev1.Pod) *corev1.Pod {
	updated := &corev1.Pod{}
	Expect(simpleK8sClient.Get(simpleCtx, client.ObjectKeyFromObject(pod), updated)).To(Succeed())
	return updated
}

// CreateSimpleBasicPowerToolSpec creates a basic PowerTool spec for testing
func CreateSimpleBasicPowerToolSpec(targetLabels map[string]string) v1alpha1.PowerToolSpec {
	return v1alpha1.PowerToolSpec{
		Targets: v1alpha1.TargetSpec{
			LabelSelector: &metav1.LabelSelector{
				MatchLabels: targetLabels,
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
}

// LogSimplePowerToolStatus logs the current status of a PowerTool for debugging
func LogSimplePowerToolStatus(powerTool *v1alpha1.PowerTool) {
	updated := GetSimplePowerTool(powerTool)
	fmt.Printf("PowerTool %s/%s Status:\n", updated.Namespace, updated.Name)
	if updated.Status.Phase != nil {
		fmt.Printf("  Phase: %s\n", *updated.Status.Phase)
	} else {
		fmt.Printf("  Phase: <nil>\n")
	}
	fmt.Printf("  Conditions:\n")
	for _, condition := range updated.Status.Conditions {
		fmt.Printf("    %s: %s - %s\n", condition.Type, condition.Status, condition.Message)
	}
}
