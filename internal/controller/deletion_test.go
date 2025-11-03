package controller

import (
	"context"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	toev1alpha1 "toe/api/v1alpha1"
)

func TestHandleDeletion(t *testing.T) {
	r := &PowerToolReconciler{}
	ctx := context.Background()

	tests := []struct {
		name      string
		powerTool *toev1alpha1.PowerTool
		wantErr   bool
	}{
		{
			name: "no finalizer - returns nil",
			powerTool: &toev1alpha1.PowerTool{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "test-tool",
					Namespace:  "default",
					Finalizers: []string{},
				},
			},
			wantErr: false,
		},
		{
			name: "with finalizer - cleanup executed",
			powerTool: &toev1alpha1.PowerTool{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "test-tool",
					Namespace:  "default",
					Finalizers: []string{"toe.run/finalizer"},
				},
			},
			wantErr: false,
		},
		{
			name: "with other finalizer - returns nil",
			powerTool: &toev1alpha1.PowerTool{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "test-tool",
					Namespace:  "default",
					Finalizers: []string{"other.io/finalizer"},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := r.handleDeletion(ctx, tt.powerTool)
			if (err != nil) != tt.wantErr {
				t.Errorf("handleDeletion() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
