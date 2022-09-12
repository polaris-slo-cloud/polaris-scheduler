package runtime

import (
	"context"

	"github.com/go-logr/logr"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/client"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/node-sampler/config"
)

// Main service that is responsible for sampling nodes.
//
// This service is responsible for managing the REST interface and the nodes watch.
type PolarisNodeSampler interface {

	// Starts the node sampler.
	//
	// The context can be used to stop the sampler.
	// Returns nil if the sampler has started successfully.
	Start(ctx context.Context) error

	// Gets the config used by this sampler.
	Config() *config.NodeSamplerConfig

	// Gets the ClusterClient used by this sampler.
	ClusterClient() client.ClusterClient

	// Gets the logger used by this sampler.
	Logger() *logr.Logger
}