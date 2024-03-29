package podspernode

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
	PluginName = "PodsPerNode"

	preScoreStateKey = "PodsPerNode.preFilterState"
)

var (
	_podsPerNode *PodsPerNodePlugin

	_ framework.Plugin          = _podsPerNode
	_ framework.PreScorePlugin  = _podsPerNode
	_ framework.ScorePlugin     = _podsPerNode
	_ framework.ScoreExtensions = _podsPerNode

	_ framework.StateData = (*preScoreState)(nil)
)

// PodsPerNodePlugin is a Score plugin that increases colocation of an application's components on a node.
type PodsPerNodePlugin struct {
	handle framework.Handle
}

type preScoreState struct {
	requiredResources     *framework.Resource
	eligibleFogNodesCount int
}

// New creates a new PodsPerNodePlugin instance.
func New(obj runtime.Object, handle framework.Handle) (framework.Plugin, error) {
	return &PodsPerNodePlugin{
		handle: handle,
	}, nil
}

// Name returns the name of this scheduler plugin.
func (me *PodsPerNodePlugin) Name() string {
	return PluginName
}

// ScoreExtensions returns a ScoreExtensions interface if the plugin implements one, or nil if does not.
func (me *PodsPerNodePlugin) ScoreExtensions() framework.ScoreExtensions {
	return me
}

// PreScore computes the total resources required by the pod and stores that info in the state.
func (me *PodsPerNodePlugin) PreScore(ctx context.Context, cycleState *framework.CycleState, pod *core.Pod, nodes []*core.Node) *framework.Status {
	_, noSvcGraphStatus := util.GetServiceGraphFromCycleStateOrStatus(cycleState)
	if noSvcGraphStatus != nil {
		return noSvcGraphStatus
	}

	requiredResources, err := util.CalcTotalRequiredResources(pod)
	if err != nil {
		return framework.AsStatus(err)
	}

	fogNodes := 0
	for _, node := range nodes {
		if util.IsFogNode(node) {
			fogNodes++
		}
	}

	cycleState.Write(preScoreStateKey, &preScoreState{requiredResources: requiredResources, eligibleFogNodesCount: fogNodes})
	return framework.NewStatus(framework.Success)
}

// Score is called on each filtered node. It must return success and an integer
// indicating the rank of the node. All scoring plugins must return success or
// the pod will be rejected.
func (me *PodsPerNodePlugin) Score(ctx context.Context, cycleState *framework.CycleState, pod *core.Pod, nodeName string) (int64, *framework.Status) {
	_, noSvcGraphStatus := util.GetServiceGraphFromCycleStateOrStatus(cycleState)
	if noSvcGraphStatus != nil {
		return 100, noSvcGraphStatus
	}

	nodeInfo, err := util.GetNodeByName(me.handle, nodeName)
	if err != nil {
		return 0, framework.AsStatus(fmt.Errorf("%s", err))
	}

	requiredResourcesInfo, err := getPreScoreState(cycleState)
	if err != nil {
		return 0, framework.AsStatus(fmt.Errorf("%s", err))
	}

	if requiredResourcesInfo.eligibleFogNodesCount > 0 && util.IsCloudNode(nodeInfo.Node()) {
		return 0, framework.NewStatus(framework.Success, "Fog nodes are preferred if they are eligible")
	}

	maxReplicasPerNode, err := me.calcMaxReplicasPerNode(requiredResourcesInfo, pod, nodeInfo)
	if err != nil {
		return 0, framework.AsStatus(err)
	}

	var score int64
	// ToDo:  Implement configuration of this plugin to allow both modes described in the presentation form 2021-04-20
	// if me.config.moreReplicasPreferred {
	score = maxReplicasPerNode
	// } else {
	// 	var inverse float64 = 1.0 / float64(maxReplicasPerNode)
	// 	score = int64(math.Round(inverse * 100))
	// }

	return score, framework.NewStatus(framework.Success, fmt.Sprintf("Pod %s, node: %s, maxReplicasPerNode: %d, score: %d", pod.Name, nodeName, maxReplicasPerNode, score))
}

// NormalizeScore normalizes all scores to a range between 0 and 100.
func (me *PodsPerNodePlugin) NormalizeScore(ctx context.Context, cycleState *framework.CycleState, pod *core.Pod, scores framework.NodeScoreList) *framework.Status {
	_, noSvcGraphStatus := util.GetServiceGraphFromCycleStateOrStatus(cycleState)
	if noSvcGraphStatus != nil {
		return noSvcGraphStatus
	}

	util.NormalizeNodeScores(scores)
	// for _, score := range scores {
	// 	klog.Infof("Pod %s, node: %s, finalScore: %d", pod.Name, score.Name, score.Score)
	// }
	return framework.NewStatus(framework.Success)
}

func (me *PodsPerNodePlugin) calcMaxReplicasPerNode(state *preScoreState, pod *core.Pod, nodeInfo *framework.NodeInfo) (int64, error) {
	var maxReplicasByResource map[string]int64 = make(map[string]int64)

	if state.requiredResources.Memory > 0 {
		maxReplicasByResource["memory"] = (nodeInfo.Allocatable.Memory - nodeInfo.Requested.Memory) / state.requiredResources.Memory
	}
	if state.requiredResources.MilliCPU > 0 {
		maxReplicasByResource["cpu"] = (nodeInfo.Allocatable.MilliCPU - nodeInfo.Requested.MilliCPU) / state.requiredResources.MilliCPU
	}
	if state.requiredResources.EphemeralStorage > 0 {
		maxReplicasByResource["ephemeralStorage"] = (nodeInfo.Allocatable.EphemeralStorage - nodeInfo.Requested.EphemeralStorage) / state.requiredResources.EphemeralStorage
	}

	for resName, resQuant := range state.requiredResources.ScalarResources {
		if resQuant > 0 {
			maxReplicasByResource[resName.String()] = (nodeInfo.Allocatable.ScalarResources[resName] - nodeInfo.Requested.ScalarResources[resName]) / resQuant
		}
	}

	return minValue(maxReplicasByResource), nil
}

func (me *preScoreState) Clone() framework.StateData {
	return &preScoreState{
		requiredResources: me.requiredResources,
	}
}

func minValue(values map[string]int64) int64 {
	if len(values) == 0 {
		return 0
	}

	var minValue int64 = math.MaxInt64
	for _, currVal := range values {
		if currVal < minValue {
			minValue = currVal
		}
	}
	return minValue
}

func getPreScoreState(cycleState *framework.CycleState) (*preScoreState, error) {
	requiredResourcesInfo, err := cycleState.Read(preScoreStateKey)
	if err != nil {
		return nil, err
	}
	return requiredResourcesInfo.(*preScoreState), nil
}
