package pipeline

import (
	"context"

	"github.com/go-logr/logr"

	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/client"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/config"
)

// Provides access to the polaris-scheduler instance.
type PolarisScheduler interface {
	// Gets the scheduler configuration.
	Config() *config.SchedulerConfig

	// Gets the ClusterClientsManager for communicating with the node clusters.
	ClusterClientsManager() client.ClusterClientsManager

	// Starts the scheduling process and then returns nil
	// or an error, if any occurred.
	Start(ctx context.Context) error

	// Stops the scheduling process.
	Stop() error

	// Gets the logger used by this scheduler.
	Logger() *logr.Logger

	// Returns true if the scheduling process has been started.
	IsActive() bool

	// Returns the number of queued pods.
	PodsInQueueCount() int

	// Returns the number of pods, for which nodes are currently being sampled.
	PodsInNodeSamplingCount() int

	// Returns the number of pods, for which nodes have been sampled, and which are
	// now waiting to enter the decision pipeline.
	PodsWaitingForDecisionPipelineCount() int

	// Returns the number of pods currently in the decision pipeline.
	PodsInDecisionPipelineCount() int
}
