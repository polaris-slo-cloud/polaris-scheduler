package pipeline

import (
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/config"
)

// Provides access to the polaris-scheduler instance.
type PolarisSchedulerService interface {
	// Gets the scheduler configuration.
	Config() *config.SchedulerConfig
}
