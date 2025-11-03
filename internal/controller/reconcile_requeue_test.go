package controller

import (
	"context"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	toev1alpha1 "toe/api/v1alpha1"
)

func TestReconcile_RequeueByPhase(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = toev1alpha1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	tests := []struct {
		name       string
		phase      string
		minRequeue time.Duration
		maxRequeue time.Duration
	}{
		{
			name:       "running phase - short interval",
			phase:      "Running",
			minRequeue: 4 * time.Second,
			maxRequeue: 6 * time.Second,
		},
		{
			name:       "completed phase - long interval",
			phase:      "Completed",
			minRequeue: 4 * time.Minute,
			maxRequeue: 6 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			powerTool := &toev1alpha1.PowerTool{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-tool",
					Namespace: "default",
				},
				Spec: toev1alpha1.PowerToolSpec{
					Targets: toev1alpha1.TargetSpec{
						LabelSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{"app": "test"},
						},
					},
					Tool: toev1alpha1.ToolSpec{
						Name:     "perf",
						Duration: "30s",
					},
					Output: toev1alpha1.OutputSpec{
						Mode: "ephemeral",
					},
				},
				Status: toev1alpha1.PowerToolStatus{
					Phase: &tt.phase,
				},
			}

			toolConfig := &toev1alpha1.PowerToolConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "perf-config",
					Namespace: "toe-system",
				},
				Spec: toev1alpha1.PowerToolConfigSpec{
					Image:           "test-image:latest",
					SecurityContext: toev1alpha1.SecuritySpec{},
				},
			}

			client := fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(powerTool, toolConfig).
				WithStatusSubresource(powerTool).
				Build()

			r := &PowerToolReconciler{
				Client: client,
				Scheme: scheme,
			}

			req := reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      "test-tool",
					Namespace: "default",
				},
			}

			result, err := r.Reconcile(context.Background(), req)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if result.RequeueAfter < tt.minRequeue || result.RequeueAfter > tt.maxRequeue {
				t.Errorf("expected RequeueAfter between %v and %v, got %v", tt.minRequeue, tt.maxRequeue, result.RequeueAfter)
			}
		})
	}
}
