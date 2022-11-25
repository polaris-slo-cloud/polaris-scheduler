package pipeline

import (
	"context"
)

// ToDo: Move this out from the pipeline package, because
// plugins get a ClusterAgentServices reference.

// Main service that is responsible for sampling nodes.
//
// This service is responsible for managing the REST interface and the nodes watch.
type PolarisNodeSampler interface {
	ClusterAgentServices

	// Starts the node sampler service.
	//
	// Note that, depending on the actual implementation, the REST interface may need to be started by
	// the caller after Start() returns.
	//
	// The context can be used to stop the sampler.
	// Returns nil if the sampler has started successfully.
	Start(ctx context.Context) error

	// Gets the sampling strategies available in this sampler.
	SamplingStrategies() []SamplingStrategyPlugin
}
