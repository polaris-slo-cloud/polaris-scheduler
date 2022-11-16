package v1

import (
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ServiceGraphConditionType identifies the possible types of conditions.
type ServiceGraphConditionType string

const (
	// The ServiceGraph is in a ready state, i.e., for each ServiceGraphNode,
	// the minimum number of pod replicas are in a ready state.
	ServiceGraphReady ServiceGraphConditionType = "Ready"

	// The ServiceGraph deployment is currently in progress.
	// Some pods may already be ready, but not yet the minimum number of replicas
	// required for each ServiceGraphNode.
	ServiceGraphProgressing ServiceGraphConditionType = "Progressing"

	// A failure has occurred that requires intervention by the user.
	ServiceGraphFailure ServiceGraphConditionType = "Failure"
)

// ServiceGraphCondition describes a high-level observed state of a ServiceGraph that is
// derived from lower level info contained in the ServiceGraphStatus.
type ServiceGraphCondition struct {

	// Defines the type of this condition.
	Type ServiceGraphConditionType `json:"type"`

	// The status of this condition, one of True, False, Unknown.
	Status core.ConditionStatus `json:"status"`

	// A one-word reason why the last transition from one condition to another occurred.
	Reason string `json:"reason"`

	// A human-readable message that indicates why the last transition from one condition to another occurred.
	//
	// +optional
	Message *string `json:"message,omitempty"`

	// The last time that the condition transitioned from one state to another.
	//
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
}
