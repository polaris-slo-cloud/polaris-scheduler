package util

import (
	core "k8s.io/api/core/v1"
)

// Calculates the total resource limits across all of the specified pod's containers.
func CalculateTotalPodResources(pod *core.Pod) *Resources {
	podSpec := &pod.Spec
	reqResources := NewResources()

	for i := range podSpec.Containers {
		reqResources.AddResourceList(podSpec.Containers[i].Resources.Limits)
	}

	return reqResources
}
