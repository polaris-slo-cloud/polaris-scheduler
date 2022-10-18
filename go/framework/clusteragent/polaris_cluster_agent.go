package clusteragent

import (
	"context"

	"github.com/go-logr/logr"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/client"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/config"
)

// The cluster agent service is responsible for providing a remote polaris-scheduler access to the local cluster.
type PolarisClusterAgent interface {

	// Starts the cluster agent.
	//
	// The context can be used to stop the agent.
	// Returns nil if the agent has started successfully.
	Start(ctx context.Context) error

	// Gets the config used by this agent.
	Config() *config.ClusterAgentConfig

	// Gets the ClusterClient used by this agent.
	ClusterClient() client.ClusterClient

	// Gets the logger used by this agent.
	Logger() *logr.Logger
}
