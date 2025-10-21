/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PowerTool condition types
const (
	PowerToolConditionReady      = "Ready"
	PowerToolConditionRunning    = "Running"
	PowerToolConditionCompleted  = "Completed"
	PowerToolConditionFailed     = "Failed"
	PowerToolConditionConflicted = "Conflicted"
)

// PowerTool condition reasons
const (
	ReasonConflictDetected = "ConflictDetected"
	ReasonRunning          = "Running"
	ReasonCompleted        = "Completed"
	ReasonFailed           = "Failed"
	ReasonTargetsSelected  = "TargetsSelected"
)

// PowerToolSpec defines the desired state of PowerTool
type PowerToolSpec struct {
	Targets                 TargetSpec         `json:"targets"`
	Tool                    ToolSpec           `json:"tool"`
	Output                  OutputSpec         `json:"output"`
	Budgets                 *BudgetSpec        `json:"budgets,omitempty"`
	FailurePolicy           *FailurePolicySpec `json:"failurePolicy,omitempty"`
	Schedule                *string            `json:"schedule,omitempty"`
	TTLSecondsAfterFinished *int32             `json:"ttlSecondsAfterFinished,omitempty"`
}

// ToolSpec defines the tool configuration (renamed from ProfilerSpec)
type ToolSpec struct {
	Name             string                `json:"name"`
	Args             *apiextensionsv1.JSON `json:"args,omitempty"`
	Duration         string                `json:"duration"`
	Warmup           *string               `json:"warmup,omitempty"`
	ResolutionPreset *string               `json:"resolutionPreset,omitempty"`
	MaxCPUPercent    *int32                `json:"maxCPUPercent,omitempty"`
}

// PowerToolStatus defines the observed state of PowerTool
type PowerToolStatus struct {
	Phase         *string              `json:"phase,omitempty"`
	SelectedPods  *int32               `json:"selectedPods,omitempty"`
	CompletedPods *int32               `json:"completedPods,omitempty"`
	BytesWritten  *string              `json:"bytesWritten,omitempty"`
	Artifacts     []string             `json:"artifacts,omitempty"`
	LastError     *string              `json:"lastError,omitempty"`
	StartedAt     *metav1.Time         `json:"startedAt,omitempty"`
	FinishedAt    *metav1.Time         `json:"finishedAt,omitempty"`
	Conditions    []PowerToolCondition `json:"conditions,omitempty"`
	ActivePods    map[string]string    `json:"activePods,omitempty"` // podName -> containerName
}

// PowerToolCondition represents a condition of a PowerTool
type PowerToolCondition struct {
	Type               string      `json:"type"`
	Status             string      `json:"status"`
	LastTransitionTime metav1.Time `json:"lastTransitionTime"`
	Reason             string      `json:"reason,omitempty"`
	Message            string      `json:"message,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// PowerTool is the Schema for the powertools API
type PowerTool struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty,omitzero"`

	// spec defines the desired state of PowerTool
	// +required
	Spec PowerToolSpec `json:"spec"`

	// status defines the observed state of PowerTool
	// +optional
	Status PowerToolStatus `json:"status,omitempty,omitzero"`
}

// +kubebuilder:object:root=true

// PowerToolList contains a list of PowerTool
type PowerToolList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PowerTool `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PowerTool{}, &PowerToolList{})
}
