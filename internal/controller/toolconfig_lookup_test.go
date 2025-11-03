package controller

import (
	"context"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	toev1alpha1 "toe/api/v1alpha1"
)

func TestGetToolConfig(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = toev1alpha1.AddToScheme(scheme)

	tests := []struct {
		name        string
		toolName    string
		configs     []toev1alpha1.PowerToolConfig
		expectFound bool
		expectError bool
	}{
		{
			name:     "config found in toe-system",
			toolName: "perf",
			configs: []toev1alpha1.PowerToolConfig{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "perf-config",
						Namespace: "toe-system",
					},
					Spec: toev1alpha1.PowerToolConfigSpec{
						Image: "test:latest",
					},
				},
			},
			expectFound: true,
			expectError: false,
		},
		{
			name:     "config found in default",
			toolName: "strace",
			configs: []toev1alpha1.PowerToolConfig{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "strace-config",
						Namespace: "default",
					},
					Spec: toev1alpha1.PowerToolConfigSpec{
						Image: "test:latest",
					},
				},
			},
			expectFound: true,
			expectError: false,
		},
		{
			name:        "config not found",
			toolName:    "nonexistent",
			configs:     []toev1alpha1.PowerToolConfig{},
			expectFound: false,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			objects := []runtime.Object{}
			for i := range tt.configs {
				objects = append(objects, &tt.configs[i])
			}

			client := fake.NewClientBuilder().
				WithScheme(scheme).
				WithRuntimeObjects(objects...).
				Build()

			r := &PowerToolReconciler{
				Client: client,
				Scheme: scheme,
			}

			config, err := r.getToolConfig(context.Background(), tt.toolName)

			if tt.expectError && err == nil {
				t.Error("expected error, got nil")
			}

			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if tt.expectFound && config == nil {
				t.Error("expected config to be found, got nil")
			}

			if !tt.expectFound && config != nil {
				t.Error("expected config to be nil, got non-nil")
			}
		})
	}
}
