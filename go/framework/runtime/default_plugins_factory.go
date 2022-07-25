package runtime

import (
	"fmt"

	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/config"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/pipeline"
)

var (
	_ pipeline.PluginsFactory = (*DefaultPluginsFactory)(nil)
)

// Default implementation of the PluginsFactory.
type DefaultPluginsFactory struct {
	pluginsRegistry *pipeline.PluginsRegistry
}

func NewDefaultPluginsFactory(registry *pipeline.PluginsRegistry) *DefaultPluginsFactory {
	factory := DefaultPluginsFactory{
		pluginsRegistry: registry,
	}
	return &factory
}

func (f *DefaultPluginsFactory) NewSortPlugin(scheduler pipeline.PolarisScheduler) (pipeline.SortPlugin, error) {
	schedulerConfig := scheduler.Config()
	pluginEntry := schedulerConfig.Plugins.Sort
	if pluginEntry == nil {
		return nil, fmt.Errorf("no SortPlugin configured")
	}

	return createPluginInstance[pipeline.SortPlugin](pluginEntry.Name, f.pluginsRegistry, scheduler, pipeline.SortStage)
}

func (f *DefaultPluginsFactory) NewSampleNodesPlugin(scheduler pipeline.PolarisScheduler) (pipeline.SampleNodesPlugin, error) {
	schedulerConfig := scheduler.Config()
	pluginEntry := schedulerConfig.Plugins.SampleNodes
	if pluginEntry == nil {
		return nil, fmt.Errorf("no SortPlugin configured")
	}

	return createPluginInstance[pipeline.SampleNodesPlugin](pluginEntry.Name, f.pluginsRegistry, scheduler, pipeline.SampleNodesStage)
}

func (f *DefaultPluginsFactory) NewDecisionPipelinePlugins(scheduler pipeline.PolarisScheduler) (*pipeline.DecisionPipelinePlugins, error) {
	schedulerConfig := scheduler.Config()
	pluginsList := schedulerConfig.Plugins

	decisionPipelinePlugins := pipeline.DecisionPipelinePlugins{}
	pluginInstances := make(map[string]pipeline.Plugin, 0)

	if plugins, err := setUpPlugins[pipeline.PreFilterPlugin](pluginsList.PreFilter, f.pluginsRegistry, scheduler, pluginInstances, pipeline.PreFilterStage); err == nil {
		decisionPipelinePlugins.PreFilter = plugins
	} else {
		return nil, err
	}

	if plugins, err := setUpPlugins[pipeline.FilterPlugin](pluginsList.Filter, f.pluginsRegistry, scheduler, pluginInstances, pipeline.FilterStage); err == nil {
		decisionPipelinePlugins.Filter = plugins
	} else {
		return nil, err
	}

	if plugins, err := setUpPlugins[pipeline.PreScorePlugin](pluginsList.PreScore, f.pluginsRegistry, scheduler, pluginInstances, pipeline.PreScoreStage); err == nil {
		decisionPipelinePlugins.PreScore = plugins
	} else {
		return nil, err
	}

	if plugins, err := setUpPlugins[pipeline.ScorePlugin](pluginsList.Score, f.pluginsRegistry, scheduler, pluginInstances, pipeline.ScoreStage); err == nil {
		scorePluginsWithExt := make([]*pipeline.ScorePluginWithExtensions, len(plugins))
		for i, scorePlugin := range plugins {
			weight := pluginsList.Score[i].Weight
			if weight == 0 {
				weight = 1
			}
			scorePluginsWithExt[i] = &pipeline.ScorePluginWithExtensions{
				ScorePlugin:     scorePlugin,
				ScoreExtensions: scorePlugin.ScoreExtensions(),
				Weight:          weight,
			}
		}
		decisionPipelinePlugins.Score = scorePluginsWithExt
	} else {
		return nil, err
	}

	if plugins, err := setUpPlugins[pipeline.ReservePlugin](pluginsList.Reserve, f.pluginsRegistry, scheduler, pluginInstances, pipeline.ReserveStage); err == nil {
		decisionPipelinePlugins.Reserve = plugins
	} else {
		return nil, err
	}

	return &decisionPipelinePlugins, nil
}

// Returns the configuration for the specified plugin or nil, if none can be found.
func findPluginConfig(pluginName string, schedulerConfig *config.SchedulerConfig) config.PluginConfig {
	for _, conf := range schedulerConfig.PluginsConfig {
		if conf.Name == pluginName {
			return conf
		}
	}
	return nil
}

func createPluginInstance[T pipeline.Plugin](
	pluginName string,
	pluginsRegistry *pipeline.PluginsRegistry,
	scheduler pipeline.PolarisScheduler,
	pluginInterfaceName string,
) (T, error) {
	var nilValue T

	factoryFn := pluginsRegistry.GetPluginFactory(pluginName)
	if factoryFn == nil {
		return nilValue, fmt.Errorf("no PluginFactoryFunc found for plugin name %s", pluginName)
	}

	schedulerConfig := scheduler.Config()
	pluginConfig := findPluginConfig(pluginName, schedulerConfig)

	pluginInstance, err := factoryFn(pluginConfig, scheduler)
	if err != nil {
		return nilValue, err
	}

	if sortPluginInstance, ok := pluginInstance.(T); ok {
		return sortPluginInstance, nil
	} else {
		return nilValue, fmt.Errorf("plugin %s does not implement the %s interface", pluginName, pluginInterfaceName)
	}
}

func getExistingOrCreateNewPlugin[T pipeline.Plugin](
	pluginName string,
	pluginsRegistry *pipeline.PluginsRegistry,
	scheduler pipeline.PolarisScheduler,
	existingInstances map[string]pipeline.Plugin,
	pluginInterfaceName string,
) (T, error) {
	if pluginInstance, ok := existingInstances[pluginName]; ok {
		if typedPlugin, ok := pluginInstance.(T); ok {
			return typedPlugin, nil
		} else {
			return typedPlugin, fmt.Errorf("plugin %s does not implement the %s interface", pluginName, pluginInterfaceName)
		}
	}

	pluginInstance, err := createPluginInstance[T](pluginName, pluginsRegistry, scheduler, pluginInterfaceName)
	if err == nil {
		existingInstances[pluginName] = pluginInstance
	}
	return pluginInstance, err
}

func setUpPlugins[T pipeline.Plugin](
	list []*config.PluginListEntry,
	pluginsRegistry *pipeline.PluginsRegistry,
	scheduler pipeline.PolarisScheduler,
	existingInstances map[string]pipeline.Plugin,
	pluginInterfaceName string,
) ([]T, error) {
	pluginInstances := make([]T, len(list))

	for i, pluginEntry := range list {
		instance, err := getExistingOrCreateNewPlugin[T](pluginEntry.Name, pluginsRegistry, scheduler, existingInstances, pluginInterfaceName)
		if err != nil {
			return nil, err
		}
		pluginInstances[i] = instance
	}

	return pluginInstances, nil
}
