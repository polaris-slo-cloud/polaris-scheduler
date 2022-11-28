package batterylevel

import (
	"fmt"
	"strconv"

	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/config"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/pipeline"
)

const (
	PluginName = "BatteryLevel"

	// The label that describes the current battery level in percent (as an integer).
	NodeBatteryLevelLabel = "polaris-slo-cloud.github.io/battery.level"

	// The label that describes the battery capacity in mAh.
	NodeBatteryCapacityLabel = "polaris-slo-cloud.github.io/battery.capacity-mah"

	// The label that describes the requested minimum battery level in percent (as an integer).
	PodMinBatteryLevelLabel = "polaris-slo-cloud.github.io/battery.min-level"

	// The label that describes the minimum battery capacity in mAh.
	PodMinBatteryCapacityLabel = "polaris-slo-cloud.github.io/battery.min-capacity-mah"
)

var (
	_ pipeline.PreFilterPlugin               = (*BatteryLevelPlugin)(nil)
	_ pipeline.FilterPlugin                  = (*BatteryLevelPlugin)(nil)
	_ pipeline.ClusterAgentPluginFactoryFunc = NewBatteryLevelClusterAgentPlugin
)

// The BatteryLevelPlugin ensures that a node meets certain battery level requirements requested by a pod.
type BatteryLevelPlugin struct {
}

type nodeBatteryInfo struct {
	batteryLevelPercent             int
	batteryCapacityMilliAmpereHours int64
}

func NewBatteryLevelClusterAgentPlugin(configMap config.PluginConfig, clusterAgentServices pipeline.ClusterAgentServices) (pipeline.Plugin, error) {
	return &BatteryLevelPlugin{}, nil
}

func (blp *BatteryLevelPlugin) Name() string {
	return PluginName
}

func (blp *BatteryLevelPlugin) PreFilter(ctx pipeline.SchedulingContext, podInfo *pipeline.PodInfo) pipeline.Status {
	minBatteryLevelStr, ok := podInfo.Pod.Labels[PodMinBatteryLevelLabel]
	if !ok {
		// No battery requirements specified, we can ignore this pod.
		return pipeline.NewSuccessStatus()
	}
	minBatteryLevel, err := strconv.ParseInt(minBatteryLevelStr, 10, 32)
	if err != nil {
		return pipeline.NewStatus(pipeline.Unschedulable, fmt.Sprintf("invalid %s specified", PodMinBatteryLevelLabel), err.Error())
	}

	state := &batteryLevelState{
		minBatteryLevelPercent: int(minBatteryLevel),
	}

	if minBatteryCapacityStr, ok := podInfo.Pod.Labels[PodMinBatteryCapacityLabel]; ok {
		minCapacity, err := strconv.ParseInt(minBatteryCapacityStr, 10, 64)
		if err != nil {
			return pipeline.NewStatus(pipeline.Unschedulable, fmt.Sprintf("invalid %s specified", PodMinBatteryCapacityLabel), err.Error())
		}
		state.minBatteryCapacityMilliAmpereHours = &minCapacity
	}

	ctx.Write(stateKey, state)
	return pipeline.NewSuccessStatus()
}

func (blp *BatteryLevelPlugin) Filter(ctx pipeline.SchedulingContext, podInfo *pipeline.PodInfo, nodeInfo *pipeline.NodeInfo) pipeline.Status {
	state, ok := blp.readState(ctx)
	if !ok {
		// No state for this pod, so we can ignore it.
		return pipeline.NewSuccessStatus()
	}

	nodeBatteryInfo, err := fetchNodeBatteryInfo(nodeInfo)
	if err != nil {
		return pipeline.NewStatus(pipeline.Unschedulable, err.Error())
	}
	if nodeBatteryInfo == nil {
		// The node does not have a battery, e.g., a server.
		return pipeline.NewSuccessStatus()
	}

	if state.minBatteryCapacityMilliAmpereHours != nil && nodeBatteryInfo.batteryCapacityMilliAmpereHours < *state.minBatteryCapacityMilliAmpereHours {
		return pipeline.NewStatus(pipeline.Unschedulable, "the node's battery capacity does not match the pod's requirements")
	}
	if nodeBatteryInfo.batteryLevelPercent < state.minBatteryLevelPercent {
		return pipeline.NewStatus(pipeline.Unschedulable, "the node's battery level is below the pod's requirements")
	}

	return pipeline.NewSuccessStatus()
}

func (blp *BatteryLevelPlugin) readState(ctx pipeline.SchedulingContext) (*batteryLevelState, bool) {
	state, ok := ctx.Read(stateKey)
	if !ok {
		return nil, false
	}
	resState, ok := state.(*batteryLevelState)
	if !ok {
		panic(fmt.Sprintf("invalid object stored as %s", stateKey))
	}
	return resState, true
}

// Fetches the node's battery info form a monitoring system (mocked using labels for the proof of concept).
// Returns the battery information or nil, if the node has no battery (e.g., a server).
func fetchNodeBatteryInfo(nodeInfo *pipeline.NodeInfo) (*nodeBatteryInfo, error) {
	batteryLevelStr, ok := nodeInfo.Node.Labels[NodeBatteryLevelLabel]
	if !ok {
		return nil, nil
	}
	batteryLevel, err := strconv.ParseInt(batteryLevelStr, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid node battery level for %s, error: %v", nodeInfo.Node.Name, err)
	}

	batteryCapacityStr, ok := nodeInfo.Node.Labels[NodeBatteryCapacityLabel]
	if !ok {
		return nil, fmt.Errorf("node has a battery level, but no battery capacity %s", nodeInfo.Node.Name)
	}
	batteryCapacity, err := strconv.ParseInt(batteryCapacityStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid node battery capacity for %s, error: %v", nodeInfo.Node.Name, err)
	}

	ret := &nodeBatteryInfo{
		batteryLevelPercent:             int(batteryLevel),
		batteryCapacityMilliAmpereHours: batteryCapacity,
	}
	return ret, nil
}
