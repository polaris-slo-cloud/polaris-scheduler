package sampling

import "polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/remotesampling"

// Encapsulates a node sampling strategy.
type SamplingStrategy interface {

	// Gets the name of this sampling strategy.
	//
	// This should be in the format of a URI component, such that it can be used to register the sampling strategy in the REST API.
	// This must be unique among all sampling strategies.
	Name() string

	// Executes the sampling strategy and returns a sample of nodes or an error.
	//
	// Important: This method may be called concurrently on multiple goroutines, so its implementation must be thread-safe.
	SampleNodes(request *remotesampling.RemoteNodesSamplerRequest) (*remotesampling.RemoteNodesSamplerResponse, error)
}
