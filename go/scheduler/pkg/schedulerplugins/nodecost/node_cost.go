package nodecost

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	framework "k8s.io/kubernetes/pkg/scheduler/framework"
	"k8s.rainbow-h2020.eu/rainbow/scheduler/internal/util"
)

const (
	// PluginName is the name of this scheduler plugin.
	PluginName = "RainbowNodeCost"
)

var (
	_cloudCost *RainbowNodeCost

	_ framework.ScorePlugin = _cloudCost
)

// RainbowNodeCost is a Score plugin that provides a higher score for cheaper nodes.
type RainbowNodeCost struct {
	handle framework.Handle
}

// New creates a new RainbowPodsPerNode plugin instance.
func New(obj runtime.Object, handle framework.Handle) (framework.Plugin, error) {
	return &RainbowNodeCost{
		handle: handle,
	}, nil
}

// Name returns the name of this scheduler plugin.
func (me *RainbowNodeCost) Name() string {
	return PluginName
}

// ScoreExtensions returns a ScoreExtensions interface if it implements one, or nil if does not.
func (me *RainbowNodeCost) ScoreExtensions() framework.ScoreExtensions {
	return nil
}

// Score returns thw following scores:
// - fog node: 100
// - cloud small: 75
// - cloud medium: 50
// - cloud large: 25
func (me *RainbowNodeCost) Score(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) (int64, *framework.Status) {
	nodeInfo, err := util.GetNodeByName(me.handle, nodeName)
	if err != nil {
		return 0, framework.AsStatus(fmt.Errorf("%s", err))
	}
	node := nodeInfo.Node()

	if !util.IsCloudNode(node) {
		return 100, framework.NewStatus(framework.Success)
	}

	cloudNodeType, err := util.GetCloudNodeType(node)
	if err != nil {
		return 0, framework.AsStatus(fmt.Errorf("%s", err))
	}

	var score int64
	switch cloudNodeType {
	case "small":
		score = 75
	case "medium":
		score = 50
	case "large":
		score = 25
	}
	return score, framework.NewStatus(framework.Success)
}
