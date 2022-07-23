package pipeline

import (
	"fmt"

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

	// The SchedulingContext of this queue pod.
	Ctx SchedulingContext
}

// Supplies new pods that need to be scheduled to the scheduling pipeline.
type PodSource interface {

	// Returns a channel that emits the incoming pods that need to be scheduled.
	IncomingPods() chan *core.Pod
}

// Returns a key that can be used to identify this pod in a map.
// The key is generated according to the following scheme: "<namespace>.<name>"
func (q *QueuedPodInfo) GetKey() string {
	return fmt.Sprintf("%s.%s", q.Pod.GetNamespace(), q.Pod.GetName())
}

// Creates a new QueuedPodInfo from a pod.
func NewQueuedPodInfo(pod *core.Pod, ctx SchedulingContext) *QueuedPodInfo {
	return &QueuedPodInfo{
		PodInfo: &PodInfo{
			Pod: pod,
		},
		Ctx: ctx,
	}
}
