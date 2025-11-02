package controller

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	toev1alpha1 "toe/api/v1alpha1"
)

func TestReconcile_PowerToolNotFoundReturnsEmpty(t *testing.T) {
	scheme := runtime.NewScheme()
	require.NoError(t, toev1alpha1.AddToScheme(scheme))
	require.NoError(t, corev1.AddToScheme(scheme))

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		Build()

	k8sClient := k8sfake.NewSimpleClientset()
	reconciler := NewPowerToolReconciler(fakeClient, scheme, k8sClient)

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "nonexistent-tool",
			Namespace: "default",
		},
	}

	result, err := reconciler.Reconcile(context.Background(), req)

	assert.NoError(t, err)
	assert.Equal(t, ctrl.Result{}, result)
}

func TestReconcile_DeletionHandling(t *testing.T) {
	scheme := runtime.NewScheme()
	require.NoError(t, toev1alpha1.AddToScheme(scheme))
	require.NoError(t, corev1.AddToScheme(scheme))

	powerTool := &toev1alpha1.PowerTool{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-tool",
			Namespace:         "default",
			DeletionTimestamp: &metav1.Time{Time: time.Now()},
			Finalizers:        []string{"toe.run/powertool-cleanup"},
		},
		Spec: toev1alpha1.PowerToolSpec{
			Tool: toev1alpha1.ToolSpec{
				Name: "aperf",
			},
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(powerTool).
		Build()

	k8sClient := k8sfake.NewSimpleClientset()
	reconciler := NewPowerToolReconciler(fakeClient, scheme, k8sClient)

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      powerTool.Name,
			Namespace: powerTool.Namespace,
		},
	}

	result, err := reconciler.Reconcile(context.Background(), req)

	assert.NoError(t, err)
	assert.Equal(t, ctrl.Result{}, result)
}
