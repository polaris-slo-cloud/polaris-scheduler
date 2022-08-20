package resourcesfit

import (
	"fmt"

	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/config"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/pipeline"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/util"
)

var (
	_ pipeline.PreFilterPlugin   = (*ResourcesFitPlugin)(nil)
	_ pipeline.FilterPlugin      = (*ResourcesFitPlugin)(nil)
	_ pipeline.PluginFactoryFunc = NewResourcesFitPlugin
)

const (
	PluginName = "ResourcesFit"
)

type ResourcesFitPlugin struct {
	scheduler pipeline.PolarisScheduler
}

func NewResourcesFitPlugin(config config.PluginConfig, scheduler pipeline.PolarisScheduler) (pipeline.Plugin, error) {
	plugin := ResourcesFitPlugin{
		scheduler: scheduler,
	}

	return &plugin, nil
}

func (rf *ResourcesFitPlugin) Name() string {
	return PluginName
}

func (*ResourcesFitPlugin) PreFilter(ctx pipeline.SchedulingContext, podInfo *pipeline.PodInfo) pipeline.Status {
	reqResources := calculateTotalResources(podInfo)
	state := resourcesFitState{
		reqResources: reqResources,
	}
	ctx.Write(stateKey, state)

	return pipeline.NewSuccessStatus()
}

func (rf *ResourcesFitPlugin) Filter(ctx pipeline.SchedulingContext, podInfo *pipeline.PodInfo, nodeInfo *pipeline.NodeInfo) pipeline.Status {
	state, ok := ctx.Read(stateKey)
	if !ok {
		return pipeline.NewInternalErrorStatus(fmt.Errorf("%s not found", stateKey))
	}
	resState, ok := state.(*resourcesFitState)
	if !ok {
		return pipeline.NewInternalErrorStatus(fmt.Errorf("invalid object stored as %s", stateKey))
	}

	if resState.reqResources.LessThanOrEqual(nodeInfo.AllocatableResources) {
		return pipeline.NewSuccessStatus()
	}
	return pipeline.NewStatus(pipeline.Unschedulable, fmt.Sprintf("node %s does not have enough resources", nodeInfo.Node.Name))
}

func calculateTotalResources(podInfo *pipeline.PodInfo) *util.Resources {
	podSpec := podInfo.Pod.Spec
	reqResources := util.NewResources()

	for _, container := range podSpec.Containers {
		reqResources.Add(container.Resources.Limits)
	}

	return reqResources
}
