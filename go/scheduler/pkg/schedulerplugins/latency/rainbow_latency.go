package latency

import (
	"context"
	"fmt"

	graphpath "gonum.org/v1/gonum/graph/path"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	framework "k8s.io/kubernetes/pkg/scheduler/framework"
	"rainbow-h2020.eu/gomod/rainbow-scheduler/pkg/model/graph/regiongraph"
	"rainbow-h2020.eu/gomod/rainbow-scheduler/pkg/model/graph/servicegraph"
	"rainbow-h2020.eu/gomod/rainbow-scheduler/pkg/services/regionmanager"
	"rainbow-h2020.eu/gomod/rainbow-scheduler/pkg/util"
)

const (
	// PluginName is the name of this scheduler plugin.
	PluginName = "RainbowLatency"
)

// RainbowLatency is a Filter plugin that filters out nodes that violate the latency constraints of the application.
type RainbowLatency struct {
	regionManager regionmanager.RegionManager
}

var _ framework.FilterPlugin = &RainbowLatency{}

// New creates a new RainbowLatency plugin instance.
func New(obj runtime.Object, handle framework.Handle) (framework.Plugin, error) {
	return &RainbowLatency{
		regionManager: regionmanager.GetRegionManager(),
	}, nil
}

// Name returns the name of this scheduler plugin.
func (me *RainbowLatency) Name() string {
	return PluginName
}

// Filter checks if the latency between the node represented by nodeInfo and the message queue node is within the
// bounds required by the pod. When scheduling the message queue itself, the latency to the fog-region-head is evaluated.
// Cloud nodes are always considered to be suitable, regardless of whether they meet the latency requirements.
func (me *RainbowLatency) Filter(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeInfo *framework.NodeInfo) *framework.Status {
	region := me.regionManager.RegionGraph()
	currNode := region.NodeByLabel(nodeInfo.Node().Name)
	if currNode == nil {
		err := fmt.Errorf("Filter() failed because the node %s was not found in the region graph", nodeInfo.Node().Name)
		klog.Errorf("RainbowLatency: %s", err.Error())
		return framework.AsStatus(err)
	}

	if util.IsCloudNode(nodeInfo.Node()) {
		return framework.NewStatus(framework.Success)
	}

	svcGraph, err := util.GetServiceGraphFromState(pod, state)
	if err != nil {
		// If the pod is not part of a RAINBOW application, we skip it and pass it to the next plugin.
		klog.Infof("RainbowLatency: Pod %s is not part of a RAINBOW application, skipping it.", pod.Name)
		return framework.NewStatus(framework.Success)
	}

	targetNode, err := me.getCommunicationTargetNode(region, svcGraph, pod)
	if err != nil {
		return framework.NewStatus(framework.Unschedulable, err.Error())
	}

	maxLatency := util.GetPodMaxDelay(pod)

	fastestPath := graphpath.DijkstraFrom(currNode, region.LabeledGraph)
	currNodeLatency := fastestPath.WeightTo(targetNode.ID())

	if int64(currNodeLatency) <= maxLatency {
		return framework.NewStatus(framework.Success)
	}
	return framework.NewStatus(framework.Unschedulable, fmt.Sprintf("Node %s with latency %d ms exceeds max allowed latency of %d ms.", currNode.Label(), int64(currNodeLatency), maxLatency))
}

func (me *RainbowLatency) getCommunicationTargetNode(regGraph *regiongraph.RegionGraph, svcGraph *servicegraph.ServiceGraph, pod *v1.Pod) (*regiongraph.Node, error) {
	if util.IsPodMessageQueue(pod) {
		return regGraph.RegionHead(), nil
	}

	mqNode := svcGraph.MessageQueueNode().MicroserviceNodeInfo().ScheduledOnNode
	if mqNode != nil {
		return mqNode, nil
	}
	return nil, fmt.Errorf("Cannot schedule pod %s, because the message queue has not been scheduled yet", pod.Name)
}
