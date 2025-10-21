package controller

import (
	"testing"
	"time"

	toev1alpha1 "toe/api/v1alpha1"
)

func TestSetCondition(t *testing.T) {
	reconciler := &PowerToolReconciler{}
	powerTool := &toev1alpha1.PowerTool{
		Status: toev1alpha1.PowerToolStatus{},
	}

	// Test adding new condition
	reconciler.setCondition(powerTool, toev1alpha1.PowerToolConditionReady, "True", toev1alpha1.ReasonTargetsSelected, "Test message")

	if len(powerTool.Status.Conditions) != 1 {
		t.Errorf("Expected 1 condition, got %d", len(powerTool.Status.Conditions))
	}

	condition := powerTool.Status.Conditions[0]
	if condition.Type != toev1alpha1.PowerToolConditionReady {
		t.Errorf("Expected condition type %s, got %s", toev1alpha1.PowerToolConditionReady, condition.Type)
	}
	if condition.Status != "True" {
		t.Errorf("Expected condition status True, got %s", condition.Status)
	}
	if condition.Reason != toev1alpha1.ReasonTargetsSelected {
		t.Errorf("Expected reason %s, got %s", toev1alpha1.ReasonTargetsSelected, condition.Reason)
	}

	// Test updating existing condition
	time.Sleep(time.Millisecond) // Ensure different timestamp
	reconciler.setCondition(powerTool, toev1alpha1.PowerToolConditionReady, "False", toev1alpha1.ReasonFailed, "Updated message")

	if len(powerTool.Status.Conditions) != 1 {
		t.Errorf("Expected 1 condition after update, got %d", len(powerTool.Status.Conditions))
	}

	updatedCondition := powerTool.Status.Conditions[0]
	if updatedCondition.Status != "False" {
		t.Errorf("Expected updated condition status False, got %s", updatedCondition.Status)
	}
	if updatedCondition.Message != "Updated message" {
		t.Errorf("Expected updated message 'Updated message', got %s", updatedCondition.Message)
	}
}

func TestGetRequeueInterval(t *testing.T) {
	reconciler := &PowerToolReconciler{}

	tests := []struct {
		name     string
		phase    *string
		expected time.Duration
	}{
		{
			name:     "nil phase",
			phase:    nil,
			expected: SetupTeardownInterval,
		},
		{
			name:     "running phase",
			phase:    stringPtrTest("Running"),
			expected: ActiveRunningInterval,
		},
		{
			name:     "completed phase",
			phase:    stringPtrTest("Completed"),
			expected: CompletedJobInterval,
		},
		{
			name:     "failed phase",
			phase:    stringPtrTest("Failed"),
			expected: CompletedJobInterval,
		},
		{
			name:     "unknown phase",
			phase:    stringPtrTest("Unknown"),
			expected: SetupTeardownInterval,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			powerTool := &toev1alpha1.PowerTool{
				Status: toev1alpha1.PowerToolStatus{
					Phase: tt.phase,
				},
			}

			result := reconciler.getRequeueInterval(powerTool)
			if result != tt.expected {
				t.Errorf("Expected interval %v, got %v", tt.expected, result)
			}
		})
	}
}

func stringPtrTest(s string) *string {
	return &s
}
