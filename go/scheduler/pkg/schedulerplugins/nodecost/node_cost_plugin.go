package nodecost

import (
	"context"
	"fmt"
	"math"

	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	framework "k8s.io/kubernetes/pkg/scheduler/framework"

	"polaris-slo-cloud.github.io/polaris-scheduler/v1/scheduler/internal/util"
)

const (
	// PluginName is the name of this scheduler plugin.
	PluginName = "NodeCost"
)

var (
	_nodeCost *NodeCostPlugin

	_ framework.Plugin          = _nodeCost
	_ framework.ScorePlugin     = _nodeCost
	_ framework.ScoreExtensions = _nodeCost
)

// NodeCostPlugin is a Score plugin that provides a higher score for cheaper nodes.
type NodeCostPlugin struct {
	handle framework.Handle
}

// New creates a new NodeCostPlugin instance.
func New(obj runtime.Object, handle framework.Handle) (framework.Plugin, error) {
	return &NodeCostPlugin{
		handle: handle,
	}, nil
}

// Name returns the name of this scheduler plugin.
func (me *NodeCostPlugin) Name() string {
	return PluginName
}

// ScoreExtensions returns a ScoreExtensions interface if the plugin implements one, or nil if does not.
func (me *NodeCostPlugin) ScoreExtensions() framework.ScoreExtensions {
	return me
}

// Returns higher scores for cheaper nodes.
func (me *NodeCostPlugin) Score(ctx context.Context, state *framework.CycleState, pod *core.Pod, nodeName string) (int64, *framework.Status) {
	nodeInfo, err := util.GetNodeByName(me.handle, nodeName)
	if err != nil {
		return 0, framework.AsStatus(fmt.Errorf("%s", err))
	}

	nodeCost := util.GetNodeCost(nodeInfo)

	// When calculating the inverse of the cost we add 1 to the nodeCost to account for nodes with cost = 0
	inverseCost := 1.0/nodeCost + 1.0
	score := int64(math.Round(inverseCost * 100))

	return score, framework.NewStatus(framework.Success)
}

// NormalizeScore normalizes all scores to a range between 0 and 100.
func (me *NodeCostPlugin) NormalizeScore(ctx context.Context, state *framework.CycleState, pod *core.Pod, scores framework.NodeScoreList) *framework.Status {
	util.NormalizeNodeScores(scores)
	// for _, score := range scores {
	// 	klog.Infof("Pod %s, node: %s, finalScore: %d", pod.Name, score.Name, score.Score)
	// }
	return framework.NewStatus(framework.Success)
}
