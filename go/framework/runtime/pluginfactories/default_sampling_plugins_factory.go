package pluginfactories

import (
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/pipeline"
)

var (
	_ pipeline.SamplingPluginsFactory = (*DefaultSamplingPluginsFactory)(nil)
)

// Default implementation of the SamplingPluginsFactory.
type DefaultSamplingPluginsFactory struct {
	pluginsRegistry *pipeline.PluginsRegistry[pipeline.ClusterAgentServices]
}

func NewDefaultSamplingPluginsFactory(registry *pipeline.PluginsRegistry[pipeline.ClusterAgentServices]) *DefaultSamplingPluginsFactory {
	factory := &DefaultSamplingPluginsFactory{
		pluginsRegistry: registry,
	}
	return factory
}

func (f *DefaultSamplingPluginsFactory) NewSamplingStrategiesPlugins(clusterAgentServices pipeline.ClusterAgentServices) ([]pipeline.SamplingStrategyPlugin, error) {
	clusterAgentConfig := clusterAgentServices.Config()
	pluginsConfig := clusterAgentConfig.PluginsConfig
	pluginsList := clusterAgentConfig.SamplingPlugins.SamplingStrategies
	pluginInstances := make(map[string]pipeline.Plugin, 0)

	if plugins, err := setUpPlugins[pipeline.SamplingStrategyPlugin](pluginsList, f.pluginsRegistry, clusterAgentServices, pluginInstances, pipeline.SamplingStrategyStage, pluginsConfig); err == nil {
		return plugins, nil
	} else {
		return nil, err
	}
}

func (f *DefaultSamplingPluginsFactory) NewSamplingPipelinePlugins(clusterAgentServices pipeline.ClusterAgentServices) (*pipeline.SamplingPipelinePlugins, error) {
	clusterAgentConfig := clusterAgentServices.Config()
	pluginsConfig := clusterAgentConfig.PluginsConfig
	pluginsList := clusterAgentConfig.SamplingPlugins

	samplingPipelinePlugins := pipeline.SamplingPipelinePlugins{}
	pluginInstances := make(map[string]pipeline.Plugin, 0)

	if plugins, err := setUpPlugins[pipeline.PreFilterPlugin](pluginsList.PreFilter, f.pluginsRegistry, clusterAgentServices, pluginInstances, pipeline.PreFilterStage, pluginsConfig); err == nil {
		samplingPipelinePlugins.PreFilter = plugins
	} else {
		return nil, err
	}

	if plugins, err := setUpPlugins[pipeline.FilterPlugin](pluginsList.Filter, f.pluginsRegistry, clusterAgentServices, pluginInstances, pipeline.FilterStage, pluginsConfig); err == nil {
		samplingPipelinePlugins.Filter = plugins
	} else {
		return nil, err
	}

	if plugins, err := setUpPlugins[pipeline.PreScorePlugin](pluginsList.PreScore, f.pluginsRegistry, clusterAgentServices, pluginInstances, pipeline.PreScoreStage, pluginsConfig); err == nil {
		samplingPipelinePlugins.PreScore = plugins
	} else {
		return nil, err
	}

	if plugins, err := setUpPlugins[pipeline.ScorePlugin](pluginsList.Score, f.pluginsRegistry, clusterAgentServices, pluginInstances, pipeline.ScoreStage, pluginsConfig); err == nil {
		samplingPipelinePlugins.Score = createScorePluginsExtensions(plugins, pluginsList.Score)
	} else {
		return nil, err
	}

	return &samplingPipelinePlugins, nil
}
