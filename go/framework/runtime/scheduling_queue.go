package runtime

import (
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/pipeline"
)

// Used to queue incoming pods for the scheduling pipeline.
//
// All method implementations must be thread-safe.
type SchedulingQueue interface {

	// Adds a new pod to the scheduling queue.
	Enqueue(podInfo *pipeline.QueuedPodInfo)

	// Dequeues the next pod for scheduling.
	// If no pod is currently waiting, this method will block until there is such a pod.
	Dequeue() *pipeline.QueuedPodInfo
}
