package util

import (
	core "k8s.io/api/core/v1"
)

const (
	// The state key for storing a NodeEligibilityStats object.
	NodeEligibilityStatsInfoStateKey = "polaris-internal.node-eligibility-stats"
)

// Information about sampled and eligible node counts.
type NodeEligibilityStats struct {
	// The total number of nodes that were sampled.
	SampledNodesCount int

	// The number of nodes that are eligible to host a pod after the Filter stage in the decision pipeline.
	EligibleNodesCount int
}

// Calculates a node's available resources, i.e., its total resource - resources used by already assigned pods.
func CalculateNodeAvailableResources(node *core.Node, assignedPods []core.Pod) *Resources {
	availableResources := NewResourcesFromList(node.Status.Allocatable)

	for podIndex := range assignedPods {
		podSpec := &assignedPods[podIndex].Spec

		for containerIndex := range podSpec.Containers {
			availableResources.SubtractResourceList(podSpec.Containers[containerIndex].Resources.Limits)
		}
	}

	return availableResources
}
