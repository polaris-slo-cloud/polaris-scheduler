package broker

// Encapsulates an error response from the PolarisClusterBroker
type PolarisClusterBrokerError struct {
	Error error `json:"error" yaml:"error"`
}
