package client

import core "k8s.io/api/core/v1"

var (
	_ error = (*PolarisErrorDto)(nil)
)

// Contains the scheduling decision for a pod within a cluster.
type ClusterSchedulingDecision struct {
	// The Pod to be scheduled.
	Pod *core.Pod `json:"pod" yaml:"pod"`

	// The name of the node, to which the pod has been assigned.
	NodeName string `json:"nodeName" yaml:"nodeName"`
}

// A generic DTO for transmitting error information.
type PolarisErrorDto struct {
	Message string `json:"message" yaml:"message"`
}

func NewPolarisErrorDto(err error) *PolarisErrorDto {
	polarisErr := &PolarisErrorDto{
		Message: err.Error(),
	}
	return polarisErr
}

// Error implements error
func (e *PolarisErrorDto) Error() string {
	return e.Message
}
