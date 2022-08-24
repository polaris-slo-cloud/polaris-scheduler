package config

import (
	"runtime"
)

const (
	// Default number of nodes to sample = 2%.
	DefaultNodesToSampleBp uint32 = 200

	// Default size of the incoming pods buffer.
	DefaultIncomingPodsBufferSize uint32 = 1000
)

var (
	// Default number of parallel node samplers = number of CPU cores.
	DefaultParallelNodeSamplers uint32 = uint32(runtime.NumCPU())

	// Default number of parallel Scheduling Decision Pipelines = number of CPU cores.
	DefaultParallelDecisionPipelines uint32 = uint32(runtime.NumCPU())
)

// Represents the configuration of a polaris-scheduler instance.
type SchedulerConfig struct {

	// The name of this scheduler.
	SchedulerName string `yaml:"schedulerName"`

	// The number of nodes to sample defined as basis points (bp) of the total number of nodes.
	// 1 bp = 0.01%
	//
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=10000
	NodesToSampleBp uint32 `yaml:"nodesToSampleBp"`

	// The number of node samplers to run in parallel.
	//
	// Default: number of CPU cores.
	ParallelNodeSamplers uint32 `yaml:"parallelNodeSamplers"`

	// The number of Scheduling Decision Pipelines to run in parallel.
	//
	// Default: number of CPU cores.
	ParallelDecisionPipelines uint32 `yaml:"parallelDecisionPipelines"`

	// The size of the buffer used for incoming pods.
	//
	// Default: 1000
	IncomingPodsBufferSize uint32 `yaml:"incomingPodsBufferSize"`

	// The list of plugins for the scheduling pipeline.
	Plugins PluginsList `yaml:"plugins"`

	// (optional) Allows specifying configuration parameters for each plugin.
	PluginsConfig []*PluginsConfigListEntry `yaml:"pluginsConfig"`
}

// Sets the default values in the SchedulerConfig for fields that are not set properly.
func SetDefaultsSchedulerConfig(config *SchedulerConfig) {
	if config.NodesToSampleBp == 0 {
		config.NodesToSampleBp = DefaultNodesToSampleBp
	}
	if config.NodesToSampleBp > 10000 {
		config.NodesToSampleBp = 10000
	}

	if config.ParallelNodeSamplers == 0 {
		config.ParallelNodeSamplers = DefaultParallelNodeSamplers
	}

	if config.ParallelDecisionPipelines == 0 {
		config.ParallelDecisionPipelines = DefaultParallelDecisionPipelines
	}

	if config.IncomingPodsBufferSize == 0 {
		config.IncomingPodsBufferSize = DefaultIncomingPodsBufferSize
	}
}
