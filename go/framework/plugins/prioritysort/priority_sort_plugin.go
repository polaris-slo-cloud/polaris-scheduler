package prioritysort

import (
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/config"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/pipeline"
)

var (
	_ pipeline.SortPlugin                  = (*PrioritySortPlugin)(nil)
	_ pipeline.SchedulingPluginFactoryFunc = NewPrioritySortPlugin
)

const (
	PluginName = "PrioritySort"
)

// Implements sorting of incoming pods based on their priorities and creation timestamps.
type PrioritySortPlugin struct {
}

func (*PrioritySortPlugin) Name() string {
	return PluginName
}

// Creates a new PrioritySortPlugin instance.
func NewPrioritySortPlugin(config config.PluginConfig, scheduler pipeline.PolarisScheduler) (pipeline.Plugin, error) {
	return &PrioritySortPlugin{}, nil
}

func (ps *PrioritySortPlugin) Less(podA *pipeline.QueuedPodInfo, podB *pipeline.QueuedPodInfo) bool {
	priorityA := getPodPriority(podA)
	priorityB := getPodPriority(podB)
	return (priorityA > priorityB) || (priorityA == priorityB && podA.Pod.CreationTimestamp.Before(&podB.Pod.CreationTimestamp))
}

func getPodPriority(podInfo *pipeline.QueuedPodInfo) int32 {
	podSpec := podInfo.Pod.Spec
	if podSpec.Priority != nil {
		return *podSpec.Priority
	}
	return 0
}
