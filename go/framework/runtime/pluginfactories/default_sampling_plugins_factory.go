package pluginfactories

import (
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/pipeline"
)

var (
	_ pipeline.SamplingPluginsFactory = (*DefaultSamplingPluginsFactory)(nil)
)

// Default implementation of the SamplingPluginsFactory.
type DefaultSamplingPluginsFactory struct {
	pluginsRegistry *pipeline.PluginsRegistry[pipeline.PolarisNodeSampler]
}

func NewDefaultSamplingPluginsFactory(registry *pipeline.PluginsRegistry[pipeline.PolarisNodeSampler]) *DefaultSamplingPluginsFactory {
	factory := &DefaultSamplingPluginsFactory{
		pluginsRegistry: registry,
	}
	return factory
}

func (f *DefaultSamplingPluginsFactory) NewSamplingStrategiesPlugins(nodeSampler pipeline.PolarisNodeSampler) ([]pipeline.SamplingStrategyPlugin, error) {
	clusterAgentConfig := nodeSampler.Config()
	pluginsConfig := clusterAgentConfig.PluginsConfig
	pluginsList := clusterAgentConfig.Plugins.SamplingStrategies
	pluginInstances := make(map[string]pipeline.Plugin, 0)

	if plugins, err := setUpPlugins[pipeline.SamplingStrategyPlugin](pluginsList, f.pluginsRegistry, nodeSampler, pluginInstances, pipeline.SamplingStrategyStage, pluginsConfig); err == nil {
		return plugins, nil
	} else {
		return nil, err
	}
}

func (f *DefaultSamplingPluginsFactory) NewSamplingPipelinePlugins(nodeSampler pipeline.PolarisNodeSampler) (*pipeline.SamplingPipelinePlugins, error) {
	clusterAgentConfig := nodeSampler.Config()
	pluginsConfig := clusterAgentConfig.PluginsConfig
	pluginsList := clusterAgentConfig.Plugins

	samplingPipelinePlugins := pipeline.SamplingPipelinePlugins{}
	pluginInstances := make(map[string]pipeline.Plugin, 0)

	if plugins, err := setUpPlugins[pipeline.PreFilterPlugin](pluginsList.PreFilter, f.pluginsRegistry, nodeSampler, pluginInstances, pipeline.PreFilterStage, pluginsConfig); err == nil {
		samplingPipelinePlugins.PreFilter = plugins
	} else {
		return nil, err
	}

	if plugins, err := setUpPlugins[pipeline.FilterPlugin](pluginsList.Filter, f.pluginsRegistry, nodeSampler, pluginInstances, pipeline.FilterStage, pluginsConfig); err == nil {
		samplingPipelinePlugins.Filter = plugins
	} else {
		return nil, err
	}

	if plugins, err := setUpPlugins[pipeline.PreScorePlugin](pluginsList.PreScore, f.pluginsRegistry, nodeSampler, pluginInstances, pipeline.PreScoreStage, pluginsConfig); err == nil {
		samplingPipelinePlugins.PreScore = plugins
	} else {
		return nil, err
	}

	if plugins, err := setUpPlugins[pipeline.ScorePlugin](pluginsList.Score, f.pluginsRegistry, nodeSampler, pluginInstances, pipeline.ScoreStage, pluginsConfig); err == nil {
		samplingPipelinePlugins.Score = createScorePluginsExtensions(plugins, pluginsList.Score)
	} else {
		return nil, err
	}

	return &samplingPipelinePlugins, nil
}
