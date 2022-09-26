package config

import (
	"runtime"
)

// Defines the mode ("singleCluster" or "multiCluster") in which to operate the scheduler.
//
//   - In "singleCluster" mode, the scheduler connects directly interacts with cluster's orchestrator to
//     get incoming pods and to assign them to nodes.
//   - In "multiCluster" mode, the scheduler can handle multiple clusters (possibly operated by multiple orchestrators)
//     In this mode polaris-scheduler accepts pods through an external API and submits scheduling decisions to the polaris-scheduler-broker
//     running in each cluster.
type SchedulerOperatingMode string

const (
	SingleCluster SchedulerOperatingMode = "singleCluster"
	MultiCluster  SchedulerOperatingMode = "multiCluster"
)

const (
	// Default number of nodes to sample = 2%.
	DefaultNodesToSampleBp uint32 = 200

	// Default size of the incoming pods buffer.
	DefaultIncomingPodsBufferSize uint32 = 1000

	// Default operating mode of the scheduler.
	DefaultSchedulerOperatingMode SchedulerOperatingMode = MultiCluster

	// Default listen address for the submit pod API.
	DefaultSubmitPodListenAddress = "0.0.0.0:8080"
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
	SchedulerName string `json:"schedulerName" yaml:"schedulerName"`

	// The number of nodes to sample defined as basis points (bp) of the total number of nodes.
	// 1 bp = 0.01%
	//
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=10000
	NodesToSampleBp uint32 `json:"nodesToSampleBp" yaml:"nodesToSampleBp"`

	// The number of node samplers to run in parallel.
	//
	// Default: number of CPU cores.
	ParallelNodeSamplers uint32 `json:"parallelNodeSamplers" yaml:"parallelNodeSamplers"`

	// The number of Scheduling Decision Pipelines to run in parallel.
	//
	// Default: number of CPU cores.
	ParallelDecisionPipelines uint32 `json:"parallelDecisionPipelines" yaml:"parallelDecisionPipelines"`

	// The size of the buffer used for incoming pods.
	//
	// Default: 1000
	IncomingPodsBufferSize uint32 `json:"incomingPodsBufferSize" yaml:"incomingPodsBufferSize"`

	// Defines the mode ("singleCluster" or "multiCluster") in which to operate the scheduler.
	//
	//   - In "singleCluster" mode, the scheduler connects directly interacts with cluster's orchestrator to
	//     get incoming pods and to assign them to nodes.
	//   - In "multiCluster" mode, the scheduler can handle multiple clusters (possibly operated by multiple orchestrators)
	//     In this mode polaris-scheduler accepts pods through an external API and submits scheduling decisions to the polaris-scheduler-broker
	//     running in each cluster.
	//
	// Default: "multiCluster"
	OperatingMode SchedulerOperatingMode `json:"operatingMode" yaml:"operatingMode"`

	// The list of addresses and ports that the submit pod API should listen on in
	// the format "<IP>:<PORT>"
	// This setting applies only if "operatingMode" is set to "multiCluster".
	//
	// Default: [ "0.0.0.0:8080" ]
	SubmitPodListenOn []string `json:"submitPodListenOn" yaml:"submitPodListenOn"`

	// The list of plugins for the scheduling pipeline.
	Plugins PluginsList `json:"plugins" yaml:"plugins"`

	// (optional) Allows specifying configuration parameters for each plugin.
	PluginsConfig []*PluginsConfigListEntry `json:"pluginsConfig" yaml:"pluginsConfig"`
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

	if config.OperatingMode == "" {
		config.OperatingMode = DefaultSchedulerOperatingMode
	}

	if config.OperatingMode == MultiCluster && (config.SubmitPodListenOn == nil || len(config.SubmitPodListenOn) == 0) {
		config.SubmitPodListenOn = []string{DefaultSubmitPodListenAddress}
	}
}
