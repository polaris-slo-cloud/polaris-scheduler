package remotesampler

const (
	DefaultMaxConcurrentRequestsPerInstance = 50
)

// Configuration data for the RemoteNodesSamplerPlugin.
//
// The remoteClusters configured in the SchedulerConfig are used as sampling targets.
type RemoteNodesSamplerPluginConfig struct {

	// The sampling strategy that should be used.
	// This endpoint must be supported by all remove samplers.
	SamplingStrategy string `json:"samplingStrategy" yaml:"samplingStrategy"`

	// The maximum number of concurrent requests to remote samplers that a single instance of the plugin may make.
	//
	// Default: 50
	MaxConcurrentRequestsPerInstance int32 `json:"maxConcurrentRequestsPerInstance" yaml:"maxConcurrentRequestsPerInstance"`
}
