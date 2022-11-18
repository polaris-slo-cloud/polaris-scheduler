package config

import "runtime"

const (
	DefaultClusterAgentListenAddress = "0.0.0.0:8081"
	DefaultNodesCacheUpdateInterval  = 200
	DefaultNodesCacheUpdateQueueSize = 1000
)

var (
	DefaultParallelSamplingPipelines uint32 = uint32(runtime.NumCPU()) * 10
)

// Represents the configuration of a polaris-cluster-agent instance.
type ClusterAgentConfig struct {

	// The list of addresses and ports to listen on in
	// the format "<IP>:<PORT>"
	//
	// Default: [ "0.0.0.0:8081" ]
	ListenOn []string `json:"listenOn" yaml:"listenOn"`

	// The update interval for the nodes cache in milliseconds.
	//
	// Default: 200
	NodesCacheUpdateIntervalMs uint32 `json:"nodesCacheUpdateIntervalMs" yaml:"nodesCacheUpdateIntervalMs"`

	// The size of the update queue in the nodes cache.
	// The update queue caches watch events that arrive between the update intervals.
	//
	// Default: 1000
	NodesCacheUpdateQueueSize uint32 `json:"nodesCacheUpdateQueueSize" yaml:"nodesCacheUpdateQueueSize"`

	// The number of Sampling Pipelines to run in parallel.
	//
	// Default: 100.
	ParallelSamplingPipelines uint32 `json:"parallelSamplingPipelines" yaml:"parallelSamplingPipelines"`

	// The list of plugins for the sampling pipeline.
	SamplingPlugins SamplingPluginsList `json:"samplingPlugins" yaml:"samplingPlugins"`

	// The list of plugins for the binding pipeline.
	BindingPlugins BindingPluginsList `json:"bindingPlugins" yaml:"bindingPlugins"`

	// (optional) Allows specifying configuration parameters for each plugin.
	PluginsConfig []*PluginsConfigListEntry `json:"pluginsConfig" yaml:"pluginsConfig"`
}

// Sets the default values in the ClusterAgentConfig for fields that are not set properly.
func SetDefaultsClusterAgentConfig(config *ClusterAgentConfig) {
	if config.ListenOn == nil || len(config.ListenOn) == 0 {
		config.ListenOn = []string{DefaultClusterAgentListenAddress}
	}
	if config.NodesCacheUpdateIntervalMs == 0 {
		config.NodesCacheUpdateIntervalMs = DefaultNodesCacheUpdateInterval
	}
	if config.NodesCacheUpdateQueueSize == 0 {
		config.NodesCacheUpdateQueueSize = DefaultNodesCacheUpdateQueueSize
	}
	if config.ParallelSamplingPipelines == 0 {
		config.ParallelSamplingPipelines = DefaultParallelSamplingPipelines
	}
}
