package controller

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	toev1alpha1 "toe/api/v1alpha1"
)

func TestCheckForConflicts(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = toev1alpha1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	tests := []struct {
		name           string
		currentTool    *toev1alpha1.PowerTool
		existingTools  []toev1alpha1.PowerTool
		targetPods     []corev1.Pod
		expectConflict bool
	}{
		{
			name: "no conflicts - different pods",
			currentTool: &toev1alpha1.PowerTool{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "tool1",
					Namespace: "default",
				},
				Spec: toev1alpha1.PowerToolSpec{
					Targets: toev1alpha1.TargetSpec{
						LabelSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{"app": "app1"},
						},
					},
				},
			},
			existingTools: []toev1alpha1.PowerTool{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "tool2",
						Namespace: "default",
					},
					Spec: toev1alpha1.PowerToolSpec{
						Targets: toev1alpha1.TargetSpec{
							LabelSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{"app": "app2"},
							},
						},
					},
					Status: toev1alpha1.PowerToolStatus{
						ActivePods: map[string]string{"pod2": "container2"},
					},
				},
			},
			targetPods: []corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pod1",
						Namespace: "default",
						Labels:    map[string]string{"app": "app1"},
					},
				},
			},
			expectConflict: false,
		},
		{
			name: "conflict detected - same pod",
			currentTool: &toev1alpha1.PowerTool{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "tool1",
					Namespace: "default",
				},
				Spec: toev1alpha1.PowerToolSpec{
					Targets: toev1alpha1.TargetSpec{
						LabelSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{"app": "myapp"},
						},
					},
				},
			},
			existingTools: []toev1alpha1.PowerTool{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "tool2",
						Namespace: "default",
					},
					Spec: toev1alpha1.PowerToolSpec{
						Targets: toev1alpha1.TargetSpec{
							LabelSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{"app": "myapp"},
							},
						},
					},
					Status: toev1alpha1.PowerToolStatus{
						ActivePods: map[string]string{"shared-pod": "container1"},
					},
				},
			},
			targetPods: []corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "shared-pod",
						Namespace: "default",
						Labels:    map[string]string{"app": "myapp"},
					},
				},
			},
			expectConflict: true,
		},
		{
			name: "no conflicts - no active pods in other tools",
			currentTool: &toev1alpha1.PowerTool{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "tool1",
					Namespace: "default",
				},
				Spec: toev1alpha1.PowerToolSpec{
					Targets: toev1alpha1.TargetSpec{
						LabelSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{"app": "myapp"},
						},
					},
				},
			},
			existingTools: []toev1alpha1.PowerTool{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "tool2",
						Namespace: "default",
					},
					Spec: toev1alpha1.PowerToolSpec{
						Targets: toev1alpha1.TargetSpec{
							LabelSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{"app": "myapp"},
							},
						},
					},
					Status: toev1alpha1.PowerToolStatus{
						ActivePods: map[string]string{},
					},
				},
			},
			targetPods: []corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pod1",
						Namespace: "default",
						Labels:    map[string]string{"app": "myapp"},
					},
				},
			},
			expectConflict: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			objects := []runtime.Object{tt.currentTool}
			for i := range tt.existingTools {
				objects = append(objects, &tt.existingTools[i])
			}
			for i := range tt.targetPods {
				objects = append(objects, &tt.targetPods[i])
			}

			client := fake.NewClientBuilder().
				WithScheme(scheme).
				WithRuntimeObjects(objects...).
				Build()

			r := &PowerToolReconciler{
				Client: client,
				Scheme: scheme,
			}

			hasConflict, _ := r.checkForConflicts(context.Background(), tt.currentTool, tt.targetPods)

			if hasConflict != tt.expectConflict {
				t.Errorf("checkForConflicts() = %v, want %v", hasConflict, tt.expectConflict)
			}
		})
	}
}
