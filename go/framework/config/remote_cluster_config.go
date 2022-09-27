package config

// Contains all the configuration needed to connect to a remote cluster's PolarisClusterBroker.
type RemoteClusterConfig struct {
	// The base URI of the remote cluster's PolarisClusterBroker.
	//
	// Example: "https://cluster-a:8081"
	BaseURI string `json:"baseUri" yaml:"baseUri"`

	// ToDo: Add authentication information.
}
