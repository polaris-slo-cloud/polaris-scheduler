package pipeline

import (
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/config"
)

// Interface for the owner of a Polaris plugin.
type PolarisPluginOwner interface{}

// Defines a factory function for creating Polaris plugins with a generic owner type.
type PluginFactoryFunc[O PolarisPluginOwner] func(pluginConfig config.PluginConfig, owner O) (Plugin, error)

// Defines a factory function for creating Polaris scheduling pipeline plugins.
type SchedulingPluginFactoryFunc PluginFactoryFunc[PolarisScheduler]

// Defines a factory function for creating Polaris sampling pipeline plugins.
type SamplingPluginFactoryFunc PluginFactoryFunc[PolarisNodeSampler]

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
type SchedulingPluginsFactory interface {

	// Creates a new instance of the configured SortPlugin.
	NewSortPlugin(scheduler PolarisScheduler) (SortPlugin, error)

	// Creates a new instance of the configured SampleNodesPlugin.
	NewSampleNodesPlugin(scheduler PolarisScheduler) (SampleNodesPlugin, error)

	// Creates a new set of instances of the plugins configured for the Decision Pipeline.
	NewDecisionPipelinePlugins(scheduler PolarisScheduler) (*DecisionPipelinePlugins, error)
}

// Contains plugin instances for a single SamplingPipeline instance.
//
// If a plugin ties into multiple stages, e.g., PreFilter, Filter, and Score,
// the same plugin instance is used for all of them.
type SamplingPipelinePlugins struct {
	PreFilter []PreFilterPlugin

	Filter []FilterPlugin

	PreScore []PreScorePlugin

	Score []*ScorePluginWithExtensions
}

// Used to instantiate sampling plugins
type SamplingPluginsFactory interface {

	// Creates instances of all configured SamplingStrategyPlugins.
	NewSamplingStrategiesPlugins(nodeSampler PolarisNodeSampler) ([]SamplingStrategyPlugin, error)

	// Creates a new set of instances of the plugins configured for the Sampling Pipeline.
	NewSamplingPipelinePlugins(nodeSampler PolarisNodeSampler) (*SamplingPipelinePlugins, error)
}

// Contains the factory functions for all available plugins.
// The generic type parameter O defines the owner type of the created plugins (i.e., PolarisScheduler or PolarisNodeSampler).
type PluginsRegistry[O PolarisPluginOwner] struct {
	registry map[string]PluginFactoryFunc[O]
}

func NewPluginsRegistry[O PolarisPluginOwner](factories map[string]PluginFactoryFunc[O]) *PluginsRegistry[O] {
	reg := PluginsRegistry[O]{
		registry: factories,
	}
	return &reg
}

// Returns the SchedulingPluginFactoryFunc for the specified plugin name or nil, if no factory is registered for this name.
func (pr *PluginsRegistry[O]) GetPluginFactory(pluginName string) PluginFactoryFunc[O] {
	if factoryFn, ok := pr.registry[pluginName]; ok {
		return factoryFn
	}
	return nil
}
