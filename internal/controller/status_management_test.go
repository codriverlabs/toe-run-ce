package controller

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	toev1alpha1 "toe/api/v1alpha1"
)

func TestSetCondition_Comprehensive(t *testing.T) {
	tests := []struct {
		name            string
		initialStatus   toev1alpha1.PowerToolStatus
		conditionType   string
		status          string
		reason          string
		message         string
		expectedCount   int
		expectedType    string
		expectedStatus  string
		expectedReason  string
		expectedMessage string
	}{
		{
			name:            "add new condition to empty status",
			initialStatus:   toev1alpha1.PowerToolStatus{},
			conditionType:   "Ready",
			status:          "True",
			reason:          "PodFound",
			message:         "Target pod found and ready",
			expectedCount:   1,
			expectedType:    "Ready",
			expectedStatus:  "True",
			expectedReason:  "PodFound",
			expectedMessage: "Target pod found and ready",
		},
		{
			name: "update existing condition",
			initialStatus: toev1alpha1.PowerToolStatus{
				Conditions: []toev1alpha1.PowerToolCondition{
					{
						Type:    "Ready",
						Status:  "False",
						Reason:  "PodNotFound",
						Message: "Target pod not found",
					},
				},
			},
			conditionType:   "Ready",
			status:          "True",
			reason:          "PodFound",
			message:         "Target pod found and ready",
			expectedCount:   1,
			expectedType:    "Ready",
			expectedStatus:  "True",
			expectedReason:  "PodFound",
			expectedMessage: "Target pod found and ready",
		},
		{
			name: "add second condition",
			initialStatus: toev1alpha1.PowerToolStatus{
				Conditions: []toev1alpha1.PowerToolCondition{
					{
						Type:   "Ready",
						Status: "True",
						Reason: "PodFound",
					},
				},
			},
			conditionType:   "Progressing",
			status:          "True",
			reason:          "ContainerStarted",
			message:         "Ephemeral container started",
			expectedCount:   2,
			expectedType:    "Progressing",
			expectedStatus:  "True",
			expectedReason:  "ContainerStarted",
			expectedMessage: "Ephemeral container started",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			powerTool := &toev1alpha1.PowerTool{
				Status: tt.initialStatus,
			}

			reconciler := &PowerToolReconciler{}
			reconciler.setCondition(powerTool, tt.conditionType, tt.status, tt.reason, tt.message)

			assert.Len(t, powerTool.Status.Conditions, tt.expectedCount)

			// Find the condition we just set/updated
			var foundCondition *toev1alpha1.PowerToolCondition
			for i := range powerTool.Status.Conditions {
				if powerTool.Status.Conditions[i].Type == tt.expectedType {
					foundCondition = &powerTool.Status.Conditions[i]
					break
				}
			}

			assert.NotNil(t, foundCondition, "Expected condition not found")
			assert.Equal(t, tt.expectedStatus, foundCondition.Status)
			assert.Equal(t, tt.expectedReason, foundCondition.Reason)
			assert.Equal(t, tt.expectedMessage, foundCondition.Message)
			assert.NotNil(t, foundCondition.LastTransitionTime)
		})
	}
}

func TestSetCondition_TimestampUpdate(t *testing.T) {
	powerTool := &toev1alpha1.PowerTool{
		Status: toev1alpha1.PowerToolStatus{
			Conditions: []toev1alpha1.PowerToolCondition{
				{
					Type:               "Ready",
					Status:             "False",
					Reason:             "PodNotFound",
					Message:            "Target pod not found",
					LastTransitionTime: metav1.Time{Time: time.Now().Add(-1 * time.Hour)},
				},
			},
		},
	}

	oldTime := powerTool.Status.Conditions[0].LastTransitionTime

	reconciler := &PowerToolReconciler{}
	reconciler.setCondition(powerTool, "Ready", "True", "PodFound", "Target pod found")

	newTime := powerTool.Status.Conditions[0].LastTransitionTime
	assert.True(t, newTime.After(oldTime.Time), "LastTransitionTime should be updated")
}

func TestSetCondition_NoTimestampUpdateForSameStatus(t *testing.T) {
	originalTime := metav1.Time{Time: time.Now().Add(-1 * time.Hour)}
	powerTool := &toev1alpha1.PowerTool{
		Status: toev1alpha1.PowerToolStatus{
			Conditions: []toev1alpha1.PowerToolCondition{
				{
					Type:               "Ready",
					Status:             "True",
					Reason:             "PodFound",
					Message:            "Target pod found",
					LastTransitionTime: originalTime,
				},
			},
		},
	}

	reconciler := &PowerToolReconciler{}
	reconciler.setCondition(powerTool, "Ready", "True", "PodFound", "Target pod found and ready")

	// Status didn't change, so timestamp should remain the same
	assert.Equal(t, originalTime, powerTool.Status.Conditions[0].LastTransitionTime)
	// But message should be updated
	assert.Equal(t, "Target pod found and ready", powerTool.Status.Conditions[0].Message)
}

func TestGetRequeueInterval_AllPhases(t *testing.T) {
	tests := []struct {
		name             string
		phase            *string
		expectedInterval time.Duration
	}{
		{
			name:             "nil phase",
			phase:            nil,
			expectedInterval: SetupTeardownInterval,
		},
		{
			name:             "running phase",
			phase:            &[]string{"Running"}[0],
			expectedInterval: ActiveRunningInterval,
		},
		{
			name:             "completed phase",
			phase:            &[]string{"Completed"}[0],
			expectedInterval: CompletedJobInterval,
		},
		{
			name:             "failed phase",
			phase:            &[]string{"Failed"}[0],
			expectedInterval: CompletedJobInterval, // Failed jobs use completed interval
		},
		{
			name:             "pending phase",
			phase:            &[]string{"Pending"}[0],
			expectedInterval: SetupTeardownInterval,
		},
		{
			name:             "unknown phase",
			phase:            &[]string{"Unknown"}[0],
			expectedInterval: SetupTeardownInterval,
		},
	}

	reconciler := &PowerToolReconciler{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			powerTool := &toev1alpha1.PowerTool{
				Status: toev1alpha1.PowerToolStatus{
					Phase: tt.phase,
				},
			}

			interval := reconciler.getRequeueInterval(powerTool)
			assert.Equal(t, tt.expectedInterval, interval)
		})
	}
}

func TestGetRequeueInterval_EdgeCases(t *testing.T) {
	reconciler := &PowerToolReconciler{}

	// Test with empty PowerTool
	emptyTool := &toev1alpha1.PowerTool{}
	interval := reconciler.getRequeueInterval(emptyTool)
	assert.Equal(t, SetupTeardownInterval, interval)

	// Test with empty string phase
	emptyPhase := ""
	toolWithEmptyPhase := &toev1alpha1.PowerTool{
		Status: toev1alpha1.PowerToolStatus{
			Phase: &emptyPhase,
		},
	}
	interval = reconciler.getRequeueInterval(toolWithEmptyPhase)
	assert.Equal(t, SetupTeardownInterval, interval)
}
