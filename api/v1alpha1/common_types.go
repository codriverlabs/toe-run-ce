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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Common types shared across PowerTool, PowerToolConfig, and related resources

// TargetSpec defines the target for tool execution
type TargetSpec struct {
	NamespaceSelector *NamespaceSelector    `json:"namespaceSelector,omitempty"`
	LabelSelector     *metav1.LabelSelector `json:"labelSelector"`
	Container         *string               `json:"container,omitempty"`
}

// NamespaceSelector defines the namespace selection criteria
type NamespaceSelector struct {
	MatchNames []string `json:"matchNames,omitempty"`
	MatchRegex *string  `json:"matchRegex,omitempty"`
}

// OutputSpec defines the output configuration
type OutputSpec struct {
	Mode            string         `json:"mode"`
	PVC             *PVCSpec       `json:"pvc,omitempty"`
	Collector       *CollectorSpec `json:"collector,omitempty"`
	RollingInterval *string        `json:"rollingInterval,omitempty"`
	Compress        *string        `json:"compress,omitempty"`
	RetentionDays   *int32         `json:"retentionDays,omitempty"`
}

// PVCSpec defines the PVC output configuration
type PVCSpec struct {
	ClaimName string  `json:"claimName"`
	Path      *string `json:"path,omitempty"`
}

// CollectorSpec defines the collector output configuration
type CollectorSpec struct {
	Endpoint string `json:"endpoint"`
}

// BudgetSpec defines the budget configuration
type BudgetSpec struct {
	MaxConcurrentPods    *int32 `json:"maxConcurrentPods,omitempty"`
	PerPodOverheadTarget *int32 `json:"perPodOverheadTarget,omitempty"`
}

// SecuritySpec defines the security configuration
type SecuritySpec struct {
	AllowPrivileged *bool         `json:"allowPrivileged,omitempty"`
	AllowHostPID    *bool         `json:"allowHostPID,omitempty"`
	Capabilities    *Capabilities `json:"capabilities,omitempty"`
}

// Capabilities defines the container capabilities
type Capabilities struct {
	Add  []string `json:"add,omitempty"`
	Drop []string `json:"drop,omitempty"`
}

// FailurePolicySpec defines the failure policy
type FailurePolicySpec struct {
	OnError *string      `json:"onError,omitempty"`
	Backoff *BackoffSpec `json:"backoff,omitempty"`
}

// BackoffSpec defines the backoff configuration
type BackoffSpec struct {
	Initial    *string `json:"initial,omitempty"`
	Max        *string `json:"max,omitempty"`
	Multiplier *string `json:"multiplier,omitempty"`
}
