package prioritymqsort

import (
	"k8s.io/apimachinery/pkg/runtime"
	framework "k8s.io/kubernetes/pkg/scheduler/framework"
	kubequeuesort "k8s.io/kubernetes/pkg/scheduler/framework/plugins/queuesort"
	"rainbow-h2020.eu/gomod/rainbow-scheduler/pkg/util"
)

const (
	// PluginName is the name of this scheduler plugin.
	PluginName = "RainbowPriorityMqSort"
)

var (
	_priorityMqSort *RainbowPriorityMqSort

	_ framework.QueueSortPlugin = _priorityMqSort
)

// RainbowPriorityMqSort is a QueueSortPlugin that prioritizes message-queue pods and otherwise falls back to the original PrioritySort.
type RainbowPriorityMqSort struct {
	origQueueSort *kubequeuesort.PrioritySort
}

// New creates a new RainbowPriorityMqSort plugin instance.
func New(obj runtime.Object, handle framework.Handle) (framework.Plugin, error) {
	origQueueSort, err := kubequeuesort.New(obj, handle)
	if err != nil {
		return nil, err
	}

	return &RainbowPriorityMqSort{
		origQueueSort: origQueueSort.(*kubequeuesort.PrioritySort),
	}, nil
}

// Name returns the name of this scheduler plugin.
func (me *RainbowPriorityMqSort) Name() string {
	return PluginName
}

// Less prioritizes pods with a more stringent (lower) maxDelay requirement.
// If both maxDelay are the same, message-queue pods are prioritized, and otherwise falls back to the original PrioritySort.
// Returns true if podA should be scheduled before podB.
func (me *RainbowPriorityMqSort) Less(podA *framework.QueuedPodInfo, podB *framework.QueuedPodInfo) bool {
	aMaxDelay := util.GetPodMaxDelay(podA.Pod)
	bMaxDelay := util.GetPodMaxDelay(podB.Pod)
	if aMaxDelay < bMaxDelay {
		return true
	}
	if bMaxDelay < aMaxDelay {
		return false
	}

	aHostsMq := util.IsPodMessageQueue(podA.Pod)
	bHostsMq := util.IsPodMessageQueue(podB.Pod)
	if aHostsMq && !bHostsMq {
		return true
	}
	if bHostsMq && !aHostsMq {
		return false
	}

	return me.origQueueSort.Less(podA, podB)
}
