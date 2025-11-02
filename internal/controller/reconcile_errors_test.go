package controller

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	toev1alpha1 "toe/api/v1alpha1"
)

func TestReconcile_PowerToolNotFound(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = toev1alpha1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	client := fake.NewClientBuilder().WithScheme(scheme).Build()

	r := &PowerToolReconciler{
		Client: client,
		Scheme: scheme,
	}

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "nonexistent",
			Namespace: "default",
		},
	}

	result, err := r.Reconcile(context.Background(), req)

	if err != nil {
		t.Errorf("expected no error for nonexistent PowerTool, got %v", err)
	}

	if result.Requeue {
		t.Error("expected no requeue for nonexistent PowerTool")
	}
}

func TestReconcile_ToolConfigNotFound(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = toev1alpha1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

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
				Name:     "nonexistent-tool",
				Duration: "30s",
			},
			Output: toev1alpha1.OutputSpec{
				Mode: "ephemeral",
			},
		},
	}

	client := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(powerTool).
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

	_, err := r.Reconcile(context.Background(), req)

	if err == nil {
		t.Error("expected error for nonexistent ToolConfig")
	}
}

func TestReconcile_NoMatchingPods(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = toev1alpha1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	powerTool := &toev1alpha1.PowerTool{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-tool",
			Namespace: "default",
		},
		Spec: toev1alpha1.PowerToolSpec{
			Targets: toev1alpha1.TargetSpec{
				LabelSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"app": "nonexistent"},
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
	}

	toolConfig := &toev1alpha1.PowerToolConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "perf-config", // Must match {toolName}-config pattern
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
		t.Errorf("expected no error for no matching pods, got %v", err)
	}

	// Should requeue to check again later
	if !result.Requeue && result.RequeueAfter == 0 {
		t.Error("expected requeue for no matching pods")
	}
}
