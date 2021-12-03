package reserve

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	framework "k8s.io/kubernetes/pkg/scheduler/framework"
	"k8s.rainbow-h2020.eu/rainbow/orchestration/pkg/services/regionmanager"
	"k8s.rainbow-h2020.eu/rainbow/scheduler/internal/util"
)

const (
	// PluginName is the name of this scheduler plugin.
	PluginName = "Reserve"
)

// ReservePlugin is a Filter plugin that filters out nodes that violate the latency constraints of the application.
type ReservePlugin struct {
	regionManager regionmanager.RegionManager
}

var _ framework.ReservePlugin = &ReservePlugin{}

// New creates a new RainbowReserve plugin instance.
func New(obj runtime.Object, handle framework.Handle) (framework.Plugin, error) {
	return &ReservePlugin{
		regionManager: regionmanager.GetRegionManager(),
	}, nil
}

// Name returns the name of this scheduler plugin.
func (me *ReservePlugin) Name() string {
	return PluginName
}

// Reserve assigns the Kubernetes node to the node in the ServiceGraph
func (me *ReservePlugin) Reserve(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) *framework.Status {
	region := me.regionManager.RegionGraph()
	targetNode := region.NodeByLabel(nodeName)
	if targetNode == nil {
		klog.Errorf("RainbowReserve.Reserve() failed because target node was not found in region graph.")
		return framework.AsStatus(fmt.Errorf("Reserve() failed because target node was not found in region graph"))
	}

	svcGraph, err := util.GetServiceGraphFromCycleState(pod, state)
	if err != nil {
		// If the pod does not belong to a RAINBOW application, we just pass it on.
		klog.Infoln("RainbowReserve: Pod not associated with a ServiceGraph, skipping it.")
		return framework.NewStatus(framework.Success)
	}

	microserviceNode, err := util.GetServiceGraphNode(svcGraph, pod)
	if err != nil {
		klog.Errorf("RainbowReserve.Reserve() failed because of: %s", err.Error())
		return framework.AsStatus(err)
	}

	svcGraph.Mutex.Lock()
	microserviceNode.MicroserviceNodeInfo().ScheduledOnNode = targetNode
	svcGraph.Mutex.Unlock()

	return framework.NewStatus(framework.Success)
}

// Unreserve removes the Kubernetes node from the node in the ServiceGraph
// This method must not fail
func (me *ReservePlugin) Unreserve(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) {
	svcGraph, err := util.GetServiceGraphFromCycleState(pod, state)
	if err != nil {
		return
	}
	microserviceNode, err := util.GetServiceGraphNode(svcGraph, pod)
	if err != nil {
		return
	}

	svcGraph.Mutex.Lock()
	microserviceNode.MicroserviceNodeInfo().ScheduledOnNode = nil
	svcGraph.Mutex.Unlock()
}
