package queue

import (
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/pipeline"
)

// Used to queue incoming pods for the scheduling pipeline.
//
// All method implementations must be thread-safe.
type SchedulingQueue interface {

	// Adds a new pod to the scheduling queue or replaces an existing pod with the same key,
	// if it already exists in the queue.
	//
	// If the queue is closed, this method does nothing.
	Enqueue(podInfo *pipeline.QueuedPodInfo)

	// Dequeues the next pod for scheduling.
	// If no pod is currently waiting, this method will block until there is such a pod or the queue is closed.
	//
	// Returns the next queued pod or nil if the queue has been closed.
	Dequeue() *pipeline.QueuedPodInfo

	// Closes this queue and causes all pending Dequeue() operations to return nil.
	Close()

	// Returns true if the queue is closed.
	IsClosed() bool
}
