package remotesampling

import "polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/pipeline"

// DTO for requesting a nodes sample from a remote nodes sampler.
type RemoteNodesSamplerRequest struct {

	// Information about the pod that should be scheduled.
	PodInfo *pipeline.PodInfo `json:"podInfo" yaml:"podInfo"`

	// The number of nodes to sample defined as basis points (bp) of the total number of nodes.
	// 1 bp = 0.01%
	NodesToSampleBp uint32 `json:"nodesToSampleBp" yaml:"nodesToSampleBp"`
}

// DTO for the response from a remote nodes sampler.
type RemoteNodesSamplerResponse struct {

	// The nodes that have been sampled.
	Nodes []*pipeline.NodeInfo `json:"nodes" yaml:"nodes"`
}

// DTO for error responses.
type RemoteNodesSamplerError struct {
	Error error `json:"error" yaml:"error"`
}
