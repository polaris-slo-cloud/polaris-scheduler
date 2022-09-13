package remotesampling

import (
	"context"
)

// Combines a RemoteNodesSamplerResponse and a RemoteNodesSamplerError in one object.
// At all times, only one of its fields is set.
type RemoteNodesSamplerResult struct {
	// The response from the remote nodes sampler, if the request was successful.
	Response *RemoteNodesSamplerResponse

	// The error, if the request failed.
	Error *RemoteNodesSamplerError
}

// A client for obtaining node samples from a remote sampler.
type RemoteSamplerClient interface {

	// The name of the cluster, where the remote sampler is running.
	ClusterName() string

	// The base URI of the remote sampler.
	BaseURI() string

	// The name of the sampling strategy.
	SamplingStrategyName() string

	// Executes a request to the remote sampler to obtain a sample of nodes.
	//
	// Returns a sample of nodes or an error.
	SampleNodes(ctx context.Context, request *RemoteNodesSamplerRequest) (*RemoteNodesSamplerResponse, *RemoteNodesSamplerError)
}

// Facilitates the use of multiple RemoteSamplerClients.
type RemoteSamplerClientsManager interface {

	// Executes a request to all configured remote samplers to obtain node samples.
	//
	// Returns a map of responses indexed by cluster name or an error.
	// Note that an error is only returned in case of a fatal issue - if single clusters return an error, they will be contained in the results map.
	SampleNodesFromAllClusters(ctx context.Context, request *RemoteNodesSamplerRequest) (map[string]*RemoteNodesSamplerResult, error)
}
