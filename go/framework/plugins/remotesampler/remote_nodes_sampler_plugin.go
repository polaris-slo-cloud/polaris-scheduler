package remotesampler

import (
	"fmt"

	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/config"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/pipeline"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/remotesampling"
)

var (
	_ pipeline.SampleNodesPlugin           = (*RemoteNodesSamplerPlugin)(nil)
	_ pipeline.SchedulingPluginFactoryFunc = NewRemoteNodesSamplerPlugin
)

const (
	PluginName = "RemoteNodesSampler"
)

// This SampleNodesPlugin contacts a remote sampling services to get the nodes sample.
type RemoteNodesSamplerPlugin struct {
	remoteSamplersMgr remotesampling.RemoteSamplerClientsManager
	nodesToSampleBp   uint32
}

func NewRemoteNodesSamplerPlugin(pluginConfig config.PluginConfig, scheduler pipeline.PolarisScheduler) (pipeline.Plugin, error) {
	schedulerConfig := scheduler.Config()
	parsedConfig, err := parseAndValidateConfig(pluginConfig, schedulerConfig)
	if err != nil {
		return nil, err
	}

	remoteSamplersMgr := remotesampling.NewDefaultRemoteSamplerClientsManager(
		schedulerConfig.RemoteClusters,
		parsedConfig.SamplingStrategy,
		int(parsedConfig.MaxConcurrentRequestsPerInstance),
		scheduler.Logger(),
	)

	plugin := &RemoteNodesSamplerPlugin{
		remoteSamplersMgr: remoteSamplersMgr,
		nodesToSampleBp:   schedulerConfig.NodesToSampleBp,
	}

	return plugin, nil
}

func (rs *RemoteNodesSamplerPlugin) Name() string {
	return PluginName
}

func (rs *RemoteNodesSamplerPlugin) SampleNodes(ctx pipeline.SchedulingContext, podInfo *pipeline.PodInfo) ([]*pipeline.NodeInfo, pipeline.Status) {
	req := &remotesampling.RemoteNodesSamplerRequest{
		PodInfo:         podInfo,
		NodesToSampleBp: rs.nodesToSampleBp,
	}

	results, err := rs.remoteSamplersMgr.SampleNodesFromAllClusters(ctx.Context(), req)
	if err != nil {
		return nil, pipeline.NewInternalErrorStatus(err)
	}

	nodesCount := getTotalNodesCount(results)
	nodes := make([]*pipeline.NodeInfo, nodesCount)

	currIndex := 0
	for _, result := range results {
		if result.Response != nil {
			for _, nodeInfo := range result.Response.Nodes {
				nodes[currIndex] = nodeInfo
				currIndex++
			}
		}
	}

	// ToDo: Add checking if a minimum percentage of clusters has responded with samples.

	return nodes, pipeline.NewSuccessStatus()
}

func parseAndValidateConfig(rawConfig config.PluginConfig, schedulerConfig *config.SchedulerConfig) (*RemoteNodesSamplerPluginConfig, error) {
	pluginConfig := &RemoteNodesSamplerPluginConfig{}

	if strategy, err := config.ReadStringFromPluginConfig(rawConfig, "samplingStrategy"); err == nil {
		pluginConfig.SamplingStrategy = strategy
	} else {
		return nil, err
	}

	if len(schedulerConfig.RemoteClusters) == 0 {
		return nil, fmt.Errorf("no remote clusters configured")
	}

	if _, ok := rawConfig["maxConcurrentRequestsPerInstance"]; ok {
		if maxConcReq, err := config.ReadInt32FromPluginConfig(rawConfig, "maxConcurrentRequestsPerInstance"); err == nil {
			pluginConfig.MaxConcurrentRequestsPerInstance = maxConcReq
		} else {
			return nil, err
		}
	} else {
		// This value is optional, so we provide a default, if it is not set.
		pluginConfig.MaxConcurrentRequestsPerInstance = DefaultMaxConcurrentRequestsPerInstance
	}

	return pluginConfig, nil
}

func getTotalNodesCount(results map[string]*remotesampling.RemoteNodesSamplerResult) int {
	count := 0
	for _, result := range results {
		if result.Response != nil {
			count += len(result.Response.Nodes)
		}
	}
	return count
}
