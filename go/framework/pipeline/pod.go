package pipeline

import (
	core "k8s.io/api/core/v1"
)

// PodInfo stores a Pod and additional pre-computed scheduling-relevant information about it.
type PodInfo struct {

	// The Pod to be scheduled.
	Pod *core.Pod
}

// Represents information about a queued pod.
type QueuedPodInfo struct {
	*PodInfo
}
