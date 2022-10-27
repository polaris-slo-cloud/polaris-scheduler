package pluginfactories

import (
	"fmt"

	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/config"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/pipeline"
)

// Returns the configuration for the specified plugin or nil, if none can be found.
func findPluginConfig(pluginName string, pluginsConfig []*config.PluginsConfigListEntry) config.PluginConfig {
	for _, conf := range pluginsConfig {
		if conf.Name == pluginName {
			return conf.Config
		}
	}
	return nil
}

// Creates an instance of the specified plugin with the specified owner.
func createPluginInstance[T pipeline.Plugin, O pipeline.PolarisPluginOwner](
	pluginName string,
	pluginsRegistry *pipeline.PluginsRegistry[O],
	owner O,
	pluginInterfaceName string,
	pluginsConfig []*config.PluginsConfigListEntry,
) (T, error) {
	var nilValue T

	factoryFn := pluginsRegistry.GetPluginFactory(pluginName)
	if factoryFn == nil {
		return nilValue, fmt.Errorf("no PluginFactoryFunc found for plugin name %s", pluginName)
	}

	pluginConfig := findPluginConfig(pluginName, pluginsConfig)

	pluginInstance, err := factoryFn(pluginConfig, owner)
	if err != nil {
		return nilValue, err
	}

	if sortPluginInstance, ok := pluginInstance.(T); ok {
		return sortPluginInstance, nil
	} else {
		return nilValue, fmt.Errorf("plugin %s does not implement the %s interface", pluginName, pluginInterfaceName)
	}
}

func getExistingOrCreateNewPlugin[T pipeline.Plugin, O pipeline.PolarisPluginOwner](
	pluginName string,
	pluginsRegistry *pipeline.PluginsRegistry[O],
	owner O,
	existingInstances map[string]pipeline.Plugin,
	pluginInterfaceName string,
	pluginsConfig []*config.PluginsConfigListEntry,
) (T, error) {
	if pluginInstance, ok := existingInstances[pluginName]; ok {
		if typedPlugin, ok := pluginInstance.(T); ok {
			return typedPlugin, nil
		} else {
			return typedPlugin, fmt.Errorf("plugin %s does not implement the %s interface", pluginName, pluginInterfaceName)
		}
	}

	pluginInstance, err := createPluginInstance[T](pluginName, pluginsRegistry, owner, pluginInterfaceName, pluginsConfig)
	if err == nil {
		existingInstances[pluginName] = pluginInstance
	}
	return pluginInstance, err
}

func setUpPlugins[T pipeline.Plugin, O pipeline.PolarisPluginOwner](
	list []*config.PluginListEntry,
	pluginsRegistry *pipeline.PluginsRegistry[O],
	owner O,
	existingInstances map[string]pipeline.Plugin,
	pluginInterfaceName string,
	pluginsConfig []*config.PluginsConfigListEntry,
) ([]T, error) {
	pluginInstances := make([]T, len(list))

	for i, pluginEntry := range list {
		instance, err := getExistingOrCreateNewPlugin[T](pluginEntry.Name, pluginsRegistry, owner, existingInstances, pluginInterfaceName, pluginsConfig)
		if err != nil {
			return nil, err
		}
		pluginInstances[i] = instance
	}

	return pluginInstances, nil
}

func createScorePluginsExtensions(pluginInstances []pipeline.ScorePlugin, scorePluginsConfigList []*config.PluginListEntry) []*pipeline.ScorePluginWithExtensions {
	scorePluginsWithExt := make([]*pipeline.ScorePluginWithExtensions, len(pluginInstances))
	for i, scorePlugin := range pluginInstances {
		weight := scorePluginsConfigList[i].Weight
		if weight == 0 {
			weight = 1
		}
		scorePluginsWithExt[i] = &pipeline.ScorePluginWithExtensions{
			ScorePlugin:     scorePlugin,
			ScoreExtensions: scorePlugin.ScoreExtensions(),
			Weight:          weight,
		}
	}
	return scorePluginsWithExt
}
