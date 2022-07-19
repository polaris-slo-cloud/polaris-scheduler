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

// Supplies new pods that need to be scheduled to the scheduling pipeline.
type PodSource interface {

	// Returns a channel that emits the incoming pods that need to be scheduled.
	IncomingPods() chan *core.Pod
}
