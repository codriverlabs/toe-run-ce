package controller

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	toev1alpha1 "toe/api/v1alpha1"
)

func TestCheckForConflicts_MultipleScenarios(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = toev1alpha1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	tests := []struct {
		name           string
		currentTool    *toev1alpha1.PowerTool
		otherTools     []toev1alpha1.PowerTool
		targetPods     []corev1.Pod
		expectConflict bool
		conflictMsg    string
	}{
		{
			name: "no other tools",
			currentTool: &toev1alpha1.PowerTool{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "tool1",
					Namespace: "default",
				},
			},
			otherTools: []toev1alpha1.PowerTool{},
			targetPods: []corev1.Pod{
				{ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "default"}},
			},
			expectConflict: false,
		},
		{
			name: "other tool with different pods",
			currentTool: &toev1alpha1.PowerTool{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "tool1",
					Namespace: "default",
				},
			},
			otherTools: []toev1alpha1.PowerTool{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "tool2",
						Namespace: "default",
					},
					Status: toev1alpha1.PowerToolStatus{
						ActivePods: map[string]string{"pod2": "default"},
					},
				},
			},
			targetPods: []corev1.Pod{
				{ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "default"}},
			},
			expectConflict: false,
		},
		{
			name: "conflict with same pod",
			currentTool: &toev1alpha1.PowerTool{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "tool1",
					Namespace: "default",
				},
			},
			otherTools: []toev1alpha1.PowerTool{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "tool2",
						Namespace: "default",
					},
					Status: toev1alpha1.PowerToolStatus{
						ActivePods: map[string]string{"pod1": "default"},
					},
				},
			},
			targetPods: []corev1.Pod{
				{ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "default"}},
			},
			expectConflict: true,
			conflictMsg:    "already being profiled",
		},
		{
			name: "other tool with nil target pods",
			currentTool: &toev1alpha1.PowerTool{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "tool1",
					Namespace: "default",
				},
			},
			otherTools: []toev1alpha1.PowerTool{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "tool2",
						Namespace: "default",
					},
					Status: toev1alpha1.PowerToolStatus{
						ActivePods: nil,
					},
				},
			},
			targetPods: []corev1.Pod{
				{ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "default"}},
			},
			expectConflict: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			objs := []runtime.Object{tt.currentTool}
			for i := range tt.otherTools {
				objs = append(objs, &tt.otherTools[i])
			}

			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithRuntimeObjects(objs...).
				Build()

			reconciler := &PowerToolReconciler{
				Client: fakeClient,
				Scheme: scheme,
			}

			hasConflict, msg := reconciler.checkForConflicts(context.Background(), tt.currentTool, tt.targetPods)

			assert.Equal(t, tt.expectConflict, hasConflict)
			if tt.expectConflict && tt.conflictMsg != "" {
				assert.Contains(t, msg, tt.conflictMsg)
			}
		})
	}
}
