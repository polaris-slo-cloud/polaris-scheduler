package pipeline

import (
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/config"
)

// Defines a factory function for creating Polaris scheduling pipeline plugins.
type PluginFactoryFunc func(pluginConfig config.PluginConfig, scheduler PolarisScheduler) (Plugin, error)

// Combines a ScorePlugin with its ScoreExtensions.
type ScorePluginWithExtensions struct {

	// The actual score plugin instance.
	ScorePlugin

	// The ScoreExtensions supplied by the ScorePlugin or nil, if the plugin does not have any.
	ScoreExtensions

	// The weight assigned to this score plugin.
	Weight int32
}

// Contains plugin instances for a single DecisionPipeline instance.
//
// If a plugin ties into multiple stages, e.g., PreFilter, Filter, and Score,
// the same plugin instance is used for all of them.
type DecisionPipelinePlugins struct {
	PreFilter []PreFilterPlugin

	Filter []FilterPlugin

	PreScore []PreScorePlugin

	Score []*ScorePluginWithExtensions

	Reserve []ReservePlugin
}

// Used to instantiate scheduler plugins
type PluginsFactory interface {

	// Creates a new instance of the configured SortPlugin.
	NewSortPlugin(scheduler PolarisScheduler) (SortPlugin, error)

	// Creates a new instance of the configured SampleNodesPlugin.
	NewSampleNodesPlugin(scheduler PolarisScheduler) (SampleNodesPlugin, error)

	// Creates a new set of instances of the plugins configured for the Decision Pipeline.
	NewDecisionPipelinePlugins(scheduler PolarisScheduler) (*DecisionPipelinePlugins, error)
}

// Contains the factory functions for all available plugins.
type PluginsRegistry struct {
	registry map[string]PluginFactoryFunc
}

func NewPluginsRegistry(factories map[string]PluginFactoryFunc) *PluginsRegistry {
	reg := PluginsRegistry{
		registry: factories,
	}
	return &reg
}

// Returns the PluginFactoryFunc for the specified plugin name or nil, if no factory is registered for this name.
func (pr *PluginsRegistry) GetPluginFactory(pluginName string) PluginFactoryFunc {
	if factoryFn, ok := pr.registry[pluginName]; ok {
		return factoryFn
	}
	return nil
}
