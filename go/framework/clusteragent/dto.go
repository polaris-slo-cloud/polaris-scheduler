package clusteragent

import (
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/client"
)

// Encapsulates an error response from the PolarisClusterAgent
type PolarisClusterAgentError struct {
	Error *client.PolarisErrorDto `json:"error" yaml:"error"`
}
