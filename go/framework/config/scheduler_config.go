package config

import (
	"runtime"
)

// Represents the configuration of a polaris-scheduler instance.
type SchedulerConfig struct {

	// The number of nodes to sample defined as basis points (bp) of the total number of nodes.
	// 1 bp = 0.01%
	//
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=10000
	NodesToSampleBp uint32

	// The number of scheduling pipelines to run in parallel.
	ParallelSchedulingPipelines uint32
}

// Sets the default values in the SchedulerConfig for fields that are not set properly.
func SetDefaultsSchedulerConfig(config *SchedulerConfig) {
	if config.NodesToSampleBp == 0 {
		config.NodesToSampleBp = 200 // = 2%
	}
	if config.NodesToSampleBp > 10000 {
		config.NodesToSampleBp = 10000
	}

	if config.ParallelSchedulingPipelines == 0 {
		config.ParallelSchedulingPipelines = uint32(runtime.NumCPU())
	}
}
