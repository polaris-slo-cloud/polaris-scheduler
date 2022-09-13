package remotesampler

const (
	DefaultMaxConcurrentRequestsPerInstance = 50
)

// Configuration data for the RemoteNodesSamplerPlugin.
type RemoteNodesSamplerPluginConfig struct {

	// The sampling strategy that should be used.
	// This endpoint must be supported by all remove samplers.
	SamplingStrategy string `json:"samplingStrategy" yaml:"samplingStrategy"`

	// The remote sampler URIs, indexed by cluster names.
	// There must be exactly one URI for each cluster.
	//
	// Example:
	// { "clusterA": "http://sampler.cluster-a:8080/v1", "clusterB": "https://sampler.cluster-b:8888/v1" }
	RemoteSamplers map[string]string `json:"remoteSamplers" yaml:"remoteSamplers"`

	// The maximum number of concurrent requests to remote samplers that a single instance of the plugin may make.
	//
	// Default: 50
	MaxConcurrentRequestsPerInstance int32 `json:"maxConcurrentRequestsPerInstance" yaml:"maxConcurrentRequestsPerInstance"`
}
