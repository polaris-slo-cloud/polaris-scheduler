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
//
// This is intentionally not part of the ClusterClient interface, because ClusterClient
// architecturally resides at a lower level.
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

	// Executes a request to the specified percentage of all configured remote samplers to obtain node samples.
	//
	// The percentageOfClustersToSample needs to be specified as a percentage in the range (0.0, 1.0].
	//
	// Returns a map of responses indexed by cluster name or an error.
	// Note that an error is only returned in case of a fatal issue - if single clusters return an error, they will be contained in the results map.
	SampleNodesFromClusters(ctx context.Context, request *RemoteNodesSamplerRequest, percentageOfClustersToSample float64) (map[string]*RemoteNodesSamplerResult, error)
}
