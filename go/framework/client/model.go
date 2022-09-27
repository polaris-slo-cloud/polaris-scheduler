package client

import core "k8s.io/api/core/v1"

// Contains the scheduling decision for a pod within a cluster.
type ClusterSchedulingDecision struct {
	// The Pod to be scheduled.
	Pod *core.Pod `json:"pod" yaml:"pod"`

	// The name of the node, to which the pod has been assigned.
	NodeName string `json:"nodeName" yaml:"nodeName"`
}
