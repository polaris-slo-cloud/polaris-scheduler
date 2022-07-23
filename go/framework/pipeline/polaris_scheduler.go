package pipeline

import (
	"context"

	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/config"
)

// Provides access to the polaris-scheduler instance.
type PolarisScheduler interface {
	// Gets the scheduler configuration.
	Config() *config.SchedulerConfig

	// Starts the scheduling process and then returns nil
	// or an error, if any occurred.
	Start(ctx context.Context) error

	// Stops the scheduling process.
	Stop() error

	// Returns true if the scheduling process has been started.
	IsActive() bool

	// Returns the number of queued pods.
	PodsInQueueCount() int

	// Returns the number of pods currently in the scheduling pipeline.
	PodsInPipelineCount() int
}
