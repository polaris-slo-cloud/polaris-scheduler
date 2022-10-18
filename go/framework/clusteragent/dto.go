package clusteragent

// Encapsulates an error response from the PolarisClusterAgent
type PolarisClusterAgentError struct {
	Error error `json:"error" yaml:"error"`
}
