package pluginfactories

import (
	"fmt"

	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/pipeline"
)

var (
	_ pipeline.SchedulingPluginsFactory = (*DefaultSchedulingPluginsFactory)(nil)
)

// Default implementation of the SchedulingPluginsFactory.
type DefaultSchedulingPluginsFactory struct {
	pluginsRegistry *pipeline.PluginsRegistry[pipeline.PolarisScheduler]
}

func NewDefaultSchedulingPluginsFactory(registry *pipeline.PluginsRegistry[pipeline.PolarisScheduler]) *DefaultSchedulingPluginsFactory {
	factory := DefaultSchedulingPluginsFactory{
		pluginsRegistry: registry,
	}
	return &factory
}

func (f *DefaultSchedulingPluginsFactory) NewSortPlugin(scheduler pipeline.PolarisScheduler) (pipeline.SortPlugin, error) {
	schedulerConfig := scheduler.Config()
	pluginEntry := schedulerConfig.Plugins.Sort
	if pluginEntry == nil {
		return nil, fmt.Errorf("no SortPlugin configured")
	}

	return createPluginInstance[pipeline.SortPlugin](pluginEntry.Name, f.pluginsRegistry, scheduler, pipeline.SortStage, schedulerConfig.PluginsConfig)
}

func (f *DefaultSchedulingPluginsFactory) NewSampleNodesPlugin(scheduler pipeline.PolarisScheduler) (pipeline.SampleNodesPlugin, error) {
	schedulerConfig := scheduler.Config()
	pluginEntry := schedulerConfig.Plugins.SampleNodes
	if pluginEntry == nil {
		return nil, fmt.Errorf("no SortPlugin configured")
	}

	return createPluginInstance[pipeline.SampleNodesPlugin](pluginEntry.Name, f.pluginsRegistry, scheduler, pipeline.SampleNodesStage, schedulerConfig.PluginsConfig)
}

func (f *DefaultSchedulingPluginsFactory) NewDecisionPipelinePlugins(scheduler pipeline.PolarisScheduler) (*pipeline.DecisionPipelinePlugins, error) {
	schedulerConfig := scheduler.Config()
	pluginsList := schedulerConfig.Plugins

	decisionPipelinePlugins := pipeline.DecisionPipelinePlugins{}
	pluginInstances := make(map[string]pipeline.Plugin, 0)

	if plugins, err := setUpPlugins[pipeline.PreFilterPlugin](pluginsList.PreFilter, f.pluginsRegistry, scheduler, pluginInstances, pipeline.PreFilterStage, schedulerConfig.PluginsConfig); err == nil {
		decisionPipelinePlugins.PreFilter = plugins
	} else {
		return nil, err
	}

	if plugins, err := setUpPlugins[pipeline.FilterPlugin](pluginsList.Filter, f.pluginsRegistry, scheduler, pluginInstances, pipeline.FilterStage, schedulerConfig.PluginsConfig); err == nil {
		decisionPipelinePlugins.Filter = plugins
	} else {
		return nil, err
	}

	if plugins, err := setUpPlugins[pipeline.PreScorePlugin](pluginsList.PreScore, f.pluginsRegistry, scheduler, pluginInstances, pipeline.PreScoreStage, schedulerConfig.PluginsConfig); err == nil {
		decisionPipelinePlugins.PreScore = plugins
	} else {
		return nil, err
	}

	if plugins, err := setUpPlugins[pipeline.ScorePlugin](pluginsList.Score, f.pluginsRegistry, scheduler, pluginInstances, pipeline.ScoreStage, schedulerConfig.PluginsConfig); err == nil {
		decisionPipelinePlugins.Score = createScorePluginsExtensions(plugins, pluginsList.Score)
	} else {
		return nil, err
	}

	if plugins, err := setUpPlugins[pipeline.ReservePlugin](pluginsList.Reserve, f.pluginsRegistry, scheduler, pluginInstances, pipeline.ReserveStage, schedulerConfig.PluginsConfig); err == nil {
		decisionPipelinePlugins.Reserve = plugins
	} else {
		return nil, err
	}

	return &decisionPipelinePlugins, nil
}
