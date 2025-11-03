//go:build e2ekind
// +build e2ekind

package e2ekind

import (
	"bytes"
	"context"
	"io"

	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"toe/api/v1alpha1"
)

// CreateTestNamespace creates a unique test namespace
func CreateTestNamespace() *corev1.Namespace {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "toe-kind-e2e-",
		},
	}
	Expect(k8sClient.Create(ctx, ns)).To(Succeed())
	return ns
}

// DeleteTestNamespace deletes a test namespace
func DeleteTestNamespace(ns *corev1.Namespace) {
	Expect(k8sClient.Delete(ctx, ns)).To(Succeed())
}

// CreateTargetPod creates a target pod for profiling
func CreateTargetPod(namespace, name string, labels map[string]string) *corev1.Pod {
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
				Command: []string{
					"sh", "-c",
					"while true; do echo 'Running...'; sleep 10; done",
				},
			}},
		},
	}
	Expect(k8sClient.Create(ctx, pod)).To(Succeed())
	return pod
}

// WaitForPodRunning waits for a pod to reach Running phase
func WaitForPodRunning(pod *corev1.Pod) {
	Eventually(func() bool {
		updated := &corev1.Pod{}
		err := k8sClient.Get(ctx, client.ObjectKeyFromObject(pod), updated)
		if err != nil {
			return false
		}
		return updated.Status.Phase == corev1.PodRunning
	}, "60s", "2s").Should(BeTrue())
}

// CreatePowerTool creates a PowerTool resource
func CreatePowerTool(namespace, name string, spec v1alpha1.PowerToolSpec) *v1alpha1.PowerTool {
	pt := &v1alpha1.PowerTool{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: spec,
	}
	Expect(k8sClient.Create(ctx, pt)).To(Succeed())
	return pt
}

// CreatePowerToolConfig creates a PowerToolConfig resource
func CreatePowerToolConfig(namespace, name string) *v1alpha1.PowerToolConfig {
	allowPrivileged := true
	ptc := &v1alpha1.PowerToolConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1alpha1.PowerToolConfigSpec{
			Name:  name,
			Image: "ghcr.io/codriverlabs/toe-aperf:latest",
			SecurityContext: v1alpha1.SecuritySpec{
				AllowPrivileged: &allowPrivileged,
				Capabilities: &v1alpha1.Capabilities{
					Add: []string{"SYS_ADMIN", "SYS_PTRACE"},
				},
			},
		},
	}
	Expect(k8sClient.Create(ctx, ptc)).To(Succeed())
	return ptc
}

// GetPodLogs retrieves logs from a pod container
func GetPodLogs(pod *corev1.Pod, containerName string) (string, error) {
	req := clientset.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &corev1.PodLogOptions{
		Container: containerName,
	})

	podLogs, err := req.Stream(context.Background())
	if err != nil {
		return "", err
	}
	defer podLogs.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, podLogs)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

// CreatePVC creates a PersistentVolumeClaim
func CreatePVC(namespace, name string, size string) *corev1.PersistentVolumeClaim {
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: parseQuantity(size),
				},
			},
		},
	}
	Expect(k8sClient.Create(ctx, pvc)).To(Succeed())
	return pvc
}

// parseQuantity is a helper to parse resource quantities
func parseQuantity(s string) resource.Quantity {
	q, err := resource.ParseQuantity(s)
	Expect(err).NotTo(HaveOccurred())
	return q
}

// WaitForPowerToolPhase waits for PowerTool to reach expected phase
func WaitForPowerToolPhase(pt *v1alpha1.PowerTool, expectedPhase string) {
	Eventually(func() string {
		updated := &v1alpha1.PowerTool{}
		err := k8sClient.Get(ctx, client.ObjectKeyFromObject(pt), updated)
		if err != nil {
			return ""
		}
		if updated.Status.Phase == nil {
			return ""
		}
		return *updated.Status.Phase
	}, "120s", "2s").Should(Equal(expectedPhase))
}

// GetPowerTool retrieves the latest version of a PowerTool
func GetPowerTool(pt *v1alpha1.PowerTool) *v1alpha1.PowerTool {
	updated := &v1alpha1.PowerTool{}
	Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(pt), updated)).To(Succeed())
	return updated
}
