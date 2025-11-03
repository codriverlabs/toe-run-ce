package controller

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	toev1alpha1 "toe/api/v1alpha1"
)

func TestPowerToolConfigReconciler_Reconcile(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = toev1alpha1.AddToScheme(scheme)

	config := &toev1alpha1.PowerToolConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-config",
			Namespace: "default",
		},
		Spec: toev1alpha1.PowerToolConfigSpec{
			Name:  "test-tool",
			Image: "test-image:latest",
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(config).
		WithStatusSubresource(config).
		Build()

	reconciler := &PowerToolConfigReconciler{
		Client: fakeClient,
		Scheme: scheme,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-config",
			Namespace: "default",
		},
	}

	result, err := reconciler.Reconcile(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Verify status was updated
	var updated toev1alpha1.PowerToolConfig
	err = fakeClient.Get(context.Background(), req.NamespacedName, &updated)
	assert.NoError(t, err)
	assert.NotNil(t, updated.Status.LastValidated)
	assert.NotNil(t, updated.Status.Phase)
	assert.Equal(t, "Ready", *updated.Status.Phase)
	assert.Len(t, updated.Status.Conditions, 1)
	assert.Equal(t, "Ready", updated.Status.Conditions[0].Type)
}

func TestPowerToolConfigReconciler_NotFound(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = toev1alpha1.AddToScheme(scheme)

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()

	reconciler := &PowerToolConfigReconciler{
		Client: fakeClient,
		Scheme: scheme,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "nonexistent",
			Namespace: "default",
		},
	}

	result, err := reconciler.Reconcile(context.Background(), req)

	assert.NoError(t, err)
	assert.Equal(t, ctrl.Result{}, result)
}

func TestUpdateCondition(t *testing.T) {
	now := metav1.Now()

	tests := []struct {
		name       string
		existing   []toev1alpha1.PowerToolConfigCondition
		new        toev1alpha1.PowerToolConfigCondition
		expectLen  int
		expectType string
	}{
		{
			name:     "add new condition",
			existing: []toev1alpha1.PowerToolConfigCondition{},
			new: toev1alpha1.PowerToolConfigCondition{
				Type:               "Ready",
				Status:             "True",
				LastTransitionTime: now,
			},
			expectLen:  1,
			expectType: "Ready",
		},
		{
			name: "update existing condition",
			existing: []toev1alpha1.PowerToolConfigCondition{
				{
					Type:               "Ready",
					Status:             "False",
					LastTransitionTime: now,
				},
			},
			new: toev1alpha1.PowerToolConfigCondition{
				Type:               "Ready",
				Status:             "True",
				LastTransitionTime: now,
			},
			expectLen:  1,
			expectType: "Ready",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := updateCondition(tt.existing, tt.new)
			assert.Len(t, result, tt.expectLen)
			assert.Equal(t, tt.expectType, result[0].Type)
		})
	}
}

func TestStringPtr(t *testing.T) {
	s := "test"
	ptr := stringPtr(s)
	assert.NotNil(t, ptr)
	assert.Equal(t, "test", *ptr)
}
