/*
Copyright 2025.

*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PowerToolConfigSpec defines the desired state of PowerToolConfig
type PowerToolConfigSpec struct {
	// Name is the unique identifier for this power tool
	// +required
	Name string `json:"name"`

	// Image is the container image for this power tool
	// +required
	Image string `json:"image"`

	// SecurityContext defines the security context for this power tool
	// +required
	SecurityContext SecuritySpec `json:"securityContext"`

	// AllowedNamespaces restricts which namespaces can use this tool
	// If empty, tool can be used in any namespace
	// +optional
	AllowedNamespaces []string `json:"allowedNamespaces,omitempty"`

	// Description provides information about what this tool does
	// +optional
	Description *string `json:"description,omitempty"`

	// Version specifies the tool version
	// +optional
	Version *string `json:"version,omitempty"`

	// DefaultArgs provides default arguments for the tool
	// +optional
	DefaultArgs []string `json:"defaultArgs,omitempty"`

	// Resources defines the resource requirements for the ephemeral container
	// +optional
	Resources *ResourceSpec `json:"resources,omitempty"`
}

// PowerToolConfigStatus defines the observed state of PowerToolConfig
type PowerToolConfigStatus struct {
	// Phase represents the current phase of the PowerToolConfig
	// +optional
	Phase *string `json:"phase,omitempty"`

	// LastValidated indicates when this configuration was last validated
	// +optional
	LastValidated *metav1.Time `json:"lastValidated,omitempty"`

	// Conditions represent the latest available observations of the PowerToolConfig's state
	// +optional
	Conditions []PowerToolConfigCondition `json:"conditions,omitempty"`
}

// PowerToolConfigCondition represents a condition of a PowerToolConfig
type PowerToolConfigCondition struct {
	Type               string      `json:"type"`
	Status             string      `json:"status"`
	LastTransitionTime metav1.Time `json:"lastTransitionTime"`
	Reason             string      `json:"reason,omitempty"`
	Message            string      `json:"message,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// PowerToolConfig is the Schema for the powertoolconfigs API
type PowerToolConfig struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty,omitzero"`

	// spec defines the desired state of PowerToolConfig
	// +required
	Spec PowerToolConfigSpec `json:"spec"`

	// status defines the observed state of PowerToolConfig
	// +optional
	Status PowerToolConfigStatus `json:"status,omitempty,omitzero"`
}

// +kubebuilder:object:root=true

// PowerToolConfigList contains a list of PowerToolConfig
type PowerToolConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PowerToolConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PowerToolConfig{}, &PowerToolConfigList{})
}
