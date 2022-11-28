package util

import (
	core "k8s.io/api/core/v1"
)

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
