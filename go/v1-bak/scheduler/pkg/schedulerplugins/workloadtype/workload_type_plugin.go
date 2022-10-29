package workloadtype

import (
	"context"

	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	framework "k8s.io/kubernetes/pkg/scheduler/framework"

	"k8s.rainbow-h2020.eu/rainbow/scheduler/internal/util"
)

const (
	// PluginName is the name of this scheduler plugin.
	PluginName = "WorkloadType"
)

var (
	_workloadType *WorkloadTypePlugin

	_ framework.Plugin          = _workloadType
	_ framework.PreScorePlugin  = _workloadType
	_ framework.ScorePlugin     = _workloadType
	_ framework.ScoreExtensions = _workloadType
)

// WorkloadTypePlugin is a Score plugin that assigns higher scores to nodes that are known to perform well for the pod's workload type.
type WorkloadTypePlugin struct {
	handle framework.Handle
}

// New creates a new WorkloadTypePlugin instance.
func New(obj runtime.Object, handle framework.Handle) (framework.Plugin, error) {
	return &WorkloadTypePlugin{
		handle: handle,
	}, nil
}

// Name returns the name of this scheduler plugin.
func (me *WorkloadTypePlugin) Name() string {
	return PluginName
}

// PreScore determines the workload type of the pod and stores that info in the state.
func (me *WorkloadTypePlugin) PreScore(ctx context.Context, cycleState *framework.CycleState, pod *core.Pod, nodes []*core.Node) *framework.Status {
	svcGraphState, noSvcGraphStatus := util.GetServiceGraphFromCycleStateOrStatus(cycleState)
	if noSvcGraphStatus != nil {
		return noSvcGraphStatus
	}

	workloadTypeState := workloadTypeStateData{
		// ToDo: Determine workload type, based on ML data.
		workloadType: svcGraphState.ServiceGraphCRD().Name,
	}
	cycleState.Write(workloadTypeStateKey, &workloadTypeState)

	return framework.NewStatus(framework.Success)
}

// ScoreExtensions returns a ScoreExtensions interface if the plugin implements one, or nil if does not.
func (me *WorkloadTypePlugin) ScoreExtensions() framework.ScoreExtensions {
	return me
}

// Score is called on each filtered node. It must return success and an integer
// indicating the rank of the node. All scoring plugins must return success or
// the pod will be rejected.
func (me *WorkloadTypePlugin) Score(ctx context.Context, cycleState *framework.CycleState, pod *core.Pod, nodeName string) (int64, *framework.Status) {
	workloadTypeState, noSvcGraphStatus := getWorkloadTypeStateDataOrStatus(cycleState)
	if noSvcGraphStatus != nil {
		return 100, noSvcGraphStatus
	}

	// ToDo: Implementation with ML data.
	var _ = workloadTypeState

	return 100, framework.NewStatus(framework.Success)
}

// NormalizeScore normalizes all scores to a range between 0 and 100.
func (me *WorkloadTypePlugin) NormalizeScore(ctx context.Context, cycleState *framework.CycleState, pod *core.Pod, scores framework.NodeScoreList) *framework.Status {
	_, noSvcGraphStatus := getWorkloadTypeStateDataOrStatus(cycleState)
	if noSvcGraphStatus != nil {
		return noSvcGraphStatus
	}

	util.NormalizeNodeScores(scores)
	return framework.NewStatus(framework.Success)
}
