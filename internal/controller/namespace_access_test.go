package controller

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	toev1alpha1 "toe/api/v1alpha1"
)

func TestValidateNamespaceAccess(t *testing.T) {
	r := &PowerToolReconciler{}

	tests := []struct {
		name       string
		powerTool  *toev1alpha1.PowerTool
		toolConfig *toev1alpha1.PowerToolConfig
		wantErr    bool
	}{
		{
			name: "no namespace restrictions - allow all",
			powerTool: &toev1alpha1.PowerTool{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "any-namespace",
				},
			},
			toolConfig: &toev1alpha1.PowerToolConfig{
				Spec: toev1alpha1.PowerToolConfigSpec{
					AllowedNamespaces: []string{},
				},
			},
			wantErr: false,
		},
		{
			name: "allowed namespace",
			powerTool: &toev1alpha1.PowerTool{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "production",
				},
			},
			toolConfig: &toev1alpha1.PowerToolConfig{
				Spec: toev1alpha1.PowerToolConfigSpec{
					AllowedNamespaces: []string{"production", "staging"},
				},
			},
			wantErr: false,
		},
		{
			name: "disallowed namespace",
			powerTool: &toev1alpha1.PowerTool{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "development",
				},
			},
			toolConfig: &toev1alpha1.PowerToolConfig{
				Spec: toev1alpha1.PowerToolConfigSpec{
					AllowedNamespaces: []string{"production", "staging"},
				},
			},
			wantErr: true,
		},
		{
			name: "nil allowed namespaces - allow all",
			powerTool: &toev1alpha1.PowerTool{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "any-namespace",
				},
			},
			toolConfig: &toev1alpha1.PowerToolConfig{
				Spec: toev1alpha1.PowerToolConfigSpec{
					AllowedNamespaces: nil,
				},
			},
			wantErr: false,
		},
		{
			name: "single allowed namespace - match",
			powerTool: &toev1alpha1.PowerTool{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "production",
				},
			},
			toolConfig: &toev1alpha1.PowerToolConfig{
				Spec: toev1alpha1.PowerToolConfigSpec{
					AllowedNamespaces: []string{"production"},
				},
			},
			wantErr: false,
		},
		{
			name: "single allowed namespace - no match",
			powerTool: &toev1alpha1.PowerTool{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "staging",
				},
			},
			toolConfig: &toev1alpha1.PowerToolConfig{
				Spec: toev1alpha1.PowerToolConfigSpec{
					AllowedNamespaces: []string{"production"},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := r.validateNamespaceAccess(tt.powerTool, tt.toolConfig)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateNamespaceAccess() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
