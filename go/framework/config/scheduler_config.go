package config

// Represents the configuration of a polaris-scheduler instance.
type SchedulerConfig struct {

	// The number of nodes to sample defined as basis points (bp) of the total number of noder.
	// 1 bp = 0.01%
	//
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=10000
	NodesToSampleBp uint32
}
