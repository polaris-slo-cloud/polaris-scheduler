package pipeline

import (
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/config"
)

// Provides access to the polaris-scheduler instance.
type PolarisSchedulerService interface {
	// Gets the scheduler configuration.
	Config() *config.SchedulerConfig

	// Starts the scheduling process.
	Start()
}

// Creates a new instance of the default implementation of the PolarisSchedulerService.
func NewDefaultPolarisSchedulerService(config *config.SchedulerConfig, podSource PodSource) PolarisSchedulerService {
	return newPolarisSchedulerServiceImpl(config, podSource)
}
