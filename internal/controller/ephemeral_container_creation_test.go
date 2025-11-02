package controller

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	toev1alpha1 "toe/api/v1alpha1"
)

func TestCreateEphemeralContainerForPod(t *testing.T) {
	scheme := runtime.NewScheme()
	require.NoError(t, toev1alpha1.AddToScheme(scheme))
	require.NoError(t, corev1.AddToScheme(scheme))

	tests := []struct {
		name          string
		powerTool     *toev1alpha1.PowerTool
		toolConfig    *toev1alpha1.PowerToolConfig
		pod           corev1.Pod
		containerName string
		expectError   bool
	}{
		{
			name: "successful ephemeral container creation",
			powerTool: &toev1alpha1.PowerTool{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-tool",
					Namespace: "default",
				},
				Spec: toev1alpha1.PowerToolSpec{
					Tool: toev1alpha1.ToolSpec{
						Name:     "aperf",
						Duration: "30s",
					},
					Output: toev1alpha1.OutputSpec{
						Mode: "ephemeral",
					},
				},
			},
			toolConfig: &toev1alpha1.PowerToolConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "aperf",
					Namespace: "toe-system",
				},
				Spec: toev1alpha1.PowerToolConfigSpec{
					Image: "test/aperf:latest",
				},
			},
			pod: corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pod",
					Namespace: "default",
					Labels:    map[string]string{"app": "test"},
				},
				Status: corev1.PodStatus{
					Phase: corev1.PodRunning,
				},
			},
			containerName: "test-container",
			expectError:   false,
		},
		{
			name: "ephemeral container with PVC output mode",
			powerTool: &toev1alpha1.PowerTool{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-tool",
					Namespace: "default",
				},
				Spec: toev1alpha1.PowerToolSpec{
					Tool: toev1alpha1.ToolSpec{
						Name:     "aperf",
						Duration: "30s",
					},
					Output: toev1alpha1.OutputSpec{
						Mode: "pvc",
						PVC: &toev1alpha1.PVCSpec{
							ClaimName: "test-pvc",
						},
					},
				},
			},
			toolConfig: &toev1alpha1.PowerToolConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "aperf",
					Namespace: "toe-system",
				},
				Spec: toev1alpha1.PowerToolConfigSpec{
					Image: "test/aperf:latest",
				},
			},
			pod: corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pod",
					Namespace: "default",
					Labels:    map[string]string{"app": "test"},
				},
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{
						{
							Name: "pvc-volume",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: "test-pvc",
								},
							},
						},
					},
				},
				Status: corev1.PodStatus{
					Phase: corev1.PodRunning,
				},
			},
			containerName: "test-container-pvc",
			expectError:   false,
		},
		{
			name: "ephemeral container with ephemeral output mode",
			powerTool: &toev1alpha1.PowerTool{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-tool",
					Namespace: "default",
				},
				Spec: toev1alpha1.PowerToolSpec{
					Tool: toev1alpha1.ToolSpec{
						Name:     "aperf",
						Duration: "30s",
					},
					Output: toev1alpha1.OutputSpec{
						Mode: "ephemeral",
					},
				},
			},
			toolConfig: &toev1alpha1.PowerToolConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "aperf",
					Namespace: "toe-system",
				},
				Spec: toev1alpha1.PowerToolConfigSpec{
					Image: "test/aperf:latest",
				},
			},
			pod: corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pod",
					Namespace: "default",
					Labels:    map[string]string{"app": "test"},
				},
				Status: corev1.PodStatus{
					Phase: corev1.PodRunning,
				},
			},
			containerName: "test-container-ephemeral",
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(tt.toolConfig, &tt.pod).
				Build()

			k8sClient := k8sfake.NewSimpleClientset(&tt.pod)
			reconciler := NewPowerToolReconciler(fakeClient, scheme, k8sClient)

			err := reconciler.createEphemeralContainerForPod(
				context.Background(),
				tt.powerTool,
				tt.toolConfig,
				tt.pod,
				tt.containerName,
			)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCreateEphemeralContainer_WithToolArgs(t *testing.T) {
	scheme := runtime.NewScheme()
	require.NoError(t, toev1alpha1.AddToScheme(scheme))
	require.NoError(t, corev1.AddToScheme(scheme))

	powerTool := &toev1alpha1.PowerTool{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-tool",
			Namespace: "default",
		},
		Spec: toev1alpha1.PowerToolSpec{
			Tool: toev1alpha1.ToolSpec{
				Name:     "aperf",
				Duration: "30s",
				Args:     []string{"--verbose", "--output=/tmp/profile.out"},
			},
			Output: toev1alpha1.OutputSpec{
				Mode: "ephemeral",
			},
		},
	}

	toolConfig := &toev1alpha1.PowerToolConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "aperf",
			Namespace: "toe-system",
		},
		Spec: toev1alpha1.PowerToolConfigSpec{
			Image: "test/aperf:latest",
		},
	}

	pod := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
			Labels:    map[string]string{"app": "test", "tier": "backend"},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(toolConfig, &pod).
		Build()

	k8sClient := k8sfake.NewSimpleClientset(&pod)
	reconciler := NewPowerToolReconciler(fakeClient, scheme, k8sClient)

	err := reconciler.createEphemeralContainerForPod(
		context.Background(),
		powerTool,
		toolConfig,
		pod,
		"test-container-args",
	)

	assert.NoError(t, err)
}

func TestCreateEphemeralContainer_ErrorCases(t *testing.T) {
	scheme := runtime.NewScheme()
	require.NoError(t, toev1alpha1.AddToScheme(scheme))
	require.NoError(t, corev1.AddToScheme(scheme))

	tests := []struct {
		name        string
		setupClient func() (client.Client, *k8sfake.Clientset)
		expectError bool
	}{
		{
			name: "pod not found in k8s client",
			setupClient: func() (client.Client, *k8sfake.Clientset) {
				toolConfig := &toev1alpha1.PowerToolConfig{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "aperf",
						Namespace: "toe-system",
					},
					Spec: toev1alpha1.PowerToolConfigSpec{
						Image: "test/aperf:latest",
					},
				}

				fakeClient := fake.NewClientBuilder().
					WithScheme(scheme).
					WithObjects(toolConfig).
					Build()

				// Empty k8s client - pod won't be found
				k8sClient := k8sfake.NewSimpleClientset()

				return fakeClient, k8sClient
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient, k8sClient := tt.setupClient()
			reconciler := NewPowerToolReconciler(fakeClient, scheme, k8sClient)

			powerTool := &toev1alpha1.PowerTool{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-tool",
					Namespace: "default",
				},
				Spec: toev1alpha1.PowerToolSpec{
					Tool: toev1alpha1.ToolSpec{
						Name: "aperf",
					},
				},
			}

			toolConfig := &toev1alpha1.PowerToolConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "aperf",
					Namespace: "toe-system",
				},
				Spec: toev1alpha1.PowerToolConfigSpec{
					Image: "test/aperf:latest",
				},
			}

			pod := corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pod",
					Namespace: "default",
				},
			}

			err := reconciler.createEphemeralContainerForPod(
				context.Background(),
				powerTool,
				toolConfig,
				pod,
				"test-container",
			)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
