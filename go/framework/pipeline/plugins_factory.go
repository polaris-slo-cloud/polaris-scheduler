package pipeline

import (
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/config"
)

// Defines a factory function for creating Polaris scheduling pipeline plugins.
type PluginFactoryFunc func(config config.PluginConfig, scheduler PolarisScheduler) (Plugin, error)

// Combines a ScorePlugin with its ScoreExtensions.
type ScorePluginWithExtensions struct {

	// The actual score plugin instance.
	ScorePlugin

	// The ScoreExtensions supplied by the ScorePlugin or nil, if the plugin does not have any.
	ScoreExtensions
}

// Contains plugin instances for a single DecisionPipeline instance.
//
// If a plugin ties into multiple stages, e.g., PreFilter, Filter, and Score,
// the same plugin instance is used for all of them.
type DecisionPipelinePlugins struct {
	PreFilter []PreFilterPlugin

	Filter []FilterPlugin

	PreScore []PreScorePlugin

	Score []ScorePluginWithExtensions

	Reserve []ReservePlugin
}

// Used to instantiate scheduler plugins
type PluginsFactory interface {

	// Creates a new instance of the configured SortPlugin.
	NewSortPlugin() SortPlugin

	// Creates a new instance of the configured SampleNodesPlugin.
	NewSampleNodesPlugin() SampleNodesPlugin

	// Creates a new set of instances of the plugins configured for the Decision Pipeline.
	NewDecisionPipelinePlugins() DecisionPipelinePlugins
}
