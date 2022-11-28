package clusteragent

import (
	"context"

	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/pipeline"
)

// The cluster agent service is responsible for providing a remote polaris-scheduler access to the local cluster.
type PolarisClusterAgent interface {
	pipeline.ClusterAgentServices

	// Starts the cluster agent.
	//
	// The context can be used to stop the agent.
	// Returns nil if the agent has started successfully.
	Start(ctx context.Context) error
}
