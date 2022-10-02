package config

const (
	DefaultClusterBrokerListenAddress = "0.0.0.0:8081"
	DefaultNodesCacheUpdateInterval   = 200
	DefaultNodesCacheUpdateQueueSize  = 1000
)

// Represents the configuration of a polaris-cluster-broker instance.
type ClusterBrokerConfig struct {

	// The list of addresses and ports to listen on in
	// the format "<IP>:<PORT>"
	//
	// Default: [ "0.0.0.0:8081" ]
	ListenOn []string `json:"listenOn" yaml:"listenOn"`

	// The update interval for the nodes cache in milliseconds.
	//
	// Default: 200
	NodesCacheUpdateIntervalMs uint32 `json:"nodesCacheUpdateIntervalMs" yaml:"nodesCacheUpdateIntervalMs"`

	// The size of the update queue in the nodes cache.
	// The update queue caches watch events that arrive between the update intervals.
	//
	// Default: 1000
	NodesCacheUpdateQueueSize uint32 `json:"nodesCacheUpdateQueueSize" yaml:"nodesCacheUpdateQueueSize"`
}

// Sets the default values in the ClusterBrokerConfig for fields that are not set properly.
func SetDefaultsClusterBrokerConfig(config *ClusterBrokerConfig) {
	if config.ListenOn == nil || len(config.ListenOn) == 0 {
		config.ListenOn = []string{DefaultClusterBrokerListenAddress}
	}
	if config.NodesCacheUpdateIntervalMs == 0 {
		config.NodesCacheUpdateIntervalMs = DefaultNodesCacheUpdateInterval
	}
	if config.NodesCacheUpdateQueueSize == 0 {
		config.NodesCacheUpdateQueueSize = DefaultNodesCacheUpdateQueueSize
	}
}
