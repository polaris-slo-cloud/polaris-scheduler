package resourcesfit

import (
	"fmt"
	"math"

	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/config"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/pipeline"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/util"
)

var (
	_ pipeline.PreFilterPlugin               = (*ResourcesFitPlugin)(nil)
	_ pipeline.FilterPlugin                  = (*ResourcesFitPlugin)(nil)
	_ pipeline.ScorePlugin                   = (*ResourcesFitPlugin)(nil)
	_ pipeline.CheckConflictsPlugin          = (*ResourcesFitPlugin)(nil)
	_ pipeline.SchedulingPluginFactoryFunc   = NewResourcesFitSchedulingPlugin
	_ pipeline.ClusterAgentPluginFactoryFunc = NewResourcesFitClusterAgentPlugin
)

const (
	PluginName = "ResourcesFit"

	// Used to configure the scoring mode of the ResourcesFitPlugin.
	// The possible options are: LeastAllocated (default) and MostAllocated
	ScoringModeKey = "scoringMode"
)

// Used to configure the scoring mode of the ResourcesFitPlugin.
// The possible options are: LeastAllocated (default) and MostAllocated
type ResourcesFitScoringMode string

const (
	LeastAllocated ResourcesFitScoringMode = "LeastAllocated"
	MostAllocated  ResourcesFitScoringMode = "MostAllocated"
)

// Defines the type of scoring function used by the ResourcesFitPlugin
type scoringFn func(state *resourcesFitState, podInfo *pipeline.PodInfo, nodeInfo *pipeline.NodeInfo) int64

// The ResourcesFitPlugin ensures that a node fulfills the resources requirements of a pod.
// It ties into the following pipeline stages:
// - PreFilter
// - Filter
// - Score
// - CheckConflicts
type ResourcesFitPlugin struct {
	scoringFn scoringFn
}

func NewResourcesFitSchedulingPlugin(configMap config.PluginConfig, scheduler pipeline.PolarisScheduler) (pipeline.Plugin, error) {
	return newResourcesFitPlugin(configMap)
}

func NewResourcesFitClusterAgentPlugin(configMap config.PluginConfig, clusterAgentServices pipeline.ClusterAgentServices) (pipeline.Plugin, error) {
	return newResourcesFitPlugin(configMap)
}

func newResourcesFitPlugin(configMap config.PluginConfig) (pipeline.Plugin, error) {
	rf := &ResourcesFitPlugin{
		scoringFn: calculateLeastAllocatedScore,
	}

	if scoringMode, ok := configMap[ScoringModeKey]; ok {
		if typedScoringMode, ok := scoringMode.(ResourcesFitScoringMode); ok {
			switch typedScoringMode {
			case LeastAllocated:
				rf.scoringFn = calculateLeastAllocatedScore
			case MostAllocated:
				rf.scoringFn = calculateMostAllocatedScore
			default:
				return nil, fmt.Errorf("invalid value for ResourcesFit.%s: %s", ScoringModeKey, typedScoringMode)
			}
		}
	}

	return rf, nil
}

func (rf *ResourcesFitPlugin) Name() string {
	return PluginName
}

func (*ResourcesFitPlugin) PreFilter(ctx pipeline.SchedulingContext, podInfo *pipeline.PodInfo) pipeline.Status {
	reqResources := calculateTotalResources(podInfo)
	state := resourcesFitState{
		reqResources: reqResources,
	}
	ctx.Write(stateKey, &state)

	return pipeline.NewSuccessStatus()
}

func (rf *ResourcesFitPlugin) Filter(ctx pipeline.SchedulingContext, podInfo *pipeline.PodInfo, nodeInfo *pipeline.NodeInfo) pipeline.Status {
	state, err := rf.readState(ctx)
	if err != nil {
		return pipeline.NewInternalErrorStatus(err)
	}

	if state.reqResources.LessThanOrEqual(nodeInfo.AllocatableResources) {
		return pipeline.NewSuccessStatus()
	}
	return pipeline.NewStatus(pipeline.Unschedulable, fmt.Sprintf("node %s does not have enough resources", nodeInfo.Node.Name))
}

func (rf *ResourcesFitPlugin) ScoreExtensions() pipeline.ScoreExtensions {
	return nil
}

func (rf *ResourcesFitPlugin) Score(ctx pipeline.SchedulingContext, podInfo *pipeline.PodInfo, nodeInfo *pipeline.NodeInfo) (int64, pipeline.Status) {
	state, err := rf.readState(ctx)
	if err != nil {
		return 0, pipeline.NewInternalErrorStatus(err)
	}
	score := rf.scoringFn(state, podInfo, nodeInfo)
	return score, pipeline.NewSuccessStatus()
}

func (rf *ResourcesFitPlugin) CheckForConflicts(ctx pipeline.SchedulingContext, decision *pipeline.SchedulingDecision) pipeline.Status {
	// To check for conflicts we simply rerun the PreFilter and Filter stages.
	// If the selected node no longer has enough resources, we have found a scheduling conflict.
	status := rf.PreFilter(ctx, decision.Pod)
	if !pipeline.IsSuccessStatus(status) {
		return status
	}
	return rf.Filter(ctx, decision.Pod, decision.TargetNode)
}

func (rf *ResourcesFitPlugin) readState(ctx pipeline.SchedulingContext) (*resourcesFitState, error) {
	state, ok := ctx.Read(stateKey)
	if !ok {
		return nil, fmt.Errorf("%s not found", stateKey)
	}
	resState, ok := state.(*resourcesFitState)
	if !ok {
		return nil, fmt.Errorf("invalid object stored as %s", stateKey)
	}
	return resState, nil
}

func calculateTotalResources(podInfo *pipeline.PodInfo) *util.Resources {
	podSpec := podInfo.Pod.Spec
	reqResources := util.NewResources()

	for _, container := range podSpec.Containers {
		reqResources.Add(container.Resources.Limits)
	}

	return reqResources
}

// Calculates the percentage of resources used on the node as an integer between 0 and 100.
// Only resources requested by the pod are considered.
func calculateResourcesUsedPercentage(podResources *util.Resources, nodeInfo *pipeline.NodeInfo) int64 {
	var usedPercentagesSum float64 = 0.0
	nodeAllocatableRes := nodeInfo.AllocatableResources
	nodeTotalRes := nodeInfo.TotalResources

	// We assume that CPU and memory are always used (even though some pods might not specify them).
	resourceTypesCount := 2
	usedPercentagesSum += calculateResourceUsedPercentage(nodeTotalRes.MilliCPU, nodeAllocatableRes.MilliCPU)
	usedPercentagesSum += calculateResourceUsedPercentage(nodeTotalRes.MemoryBytes, nodeAllocatableRes.MemoryBytes)

	if podResources.EphemeralStorage > 0 {
		resourceTypesCount++
		usedPercentagesSum += calculateResourceUsedPercentage(nodeTotalRes.EphemeralStorage, nodeAllocatableRes.EphemeralStorage)
	}

	for resType := range podResources.ExtendedResources {
		resourceTypesCount++
		usedPercentagesSum += calculateResourceUsedPercentage(nodeTotalRes.ExtendedResources[resType], nodeAllocatableRes.ExtendedResources[resType])
	}

	usedPercentage := usedPercentagesSum / float64(resourceTypesCount)
	return int64(math.Floor(usedPercentage * 100.0))
}

func calculateResourceUsedPercentage(totalCapacity int64, availableCapacity int64) float64 {
	used := totalCapacity - availableCapacity
	return float64(used) / float64(totalCapacity)
}

// Calculates a score that favors nodes with low resource utilization.
func calculateLeastAllocatedScore(state *resourcesFitState, podInfo *pipeline.PodInfo, nodeInfo *pipeline.NodeInfo) int64 {
	usedPercentage := calculateResourcesUsedPercentage(state.reqResources, nodeInfo)
	return 100 - usedPercentage
}

// Calculates a score that favors nodes with high resource utilization.
func calculateMostAllocatedScore(state *resourcesFitState, podInfo *pipeline.PodInfo, nodeInfo *pipeline.NodeInfo) int64 {
	usedPercentage := calculateResourcesUsedPercentage(state.reqResources, nodeInfo)
	return usedPercentage
}
