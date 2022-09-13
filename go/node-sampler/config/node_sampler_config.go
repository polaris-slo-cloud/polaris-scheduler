package config

const (
	DefaultListenAddress = "0.0.0.0:8080"
)

// Represents the configuration of a polaris-node-sampler instance.
type NodeSamplerConfig struct {

	// The list of addresses and ports to listen on in
	// the format "<IP>:<PORT>"
	//
	// Default: [ "0.0.0.0:8080" ]
	ListenOn []string `json:"listenOn" yaml:"listenOn"`
}

// Sets the default values in the NodeSamplerConfig for fields that are not set properly.
func SetDefaultsNodeSamplerConfig(config *NodeSamplerConfig) {
	if config.ListenOn == nil || len(config.ListenOn) == 0 {
		config.ListenOn = []string{DefaultListenAddress}
	}
}
