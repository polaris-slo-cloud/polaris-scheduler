package pipeline

import (
	core "k8s.io/api/core/v1"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/util"
)

// NodeInfo stores a Node and additional pre-computed scheduling-relevant information about it.
type NodeInfo struct {

	// The Node described by this NodeInfo.
	Node *core.Node

	// The name of the cluster that the node is part of.
	ClusterName string

	// The resources that are currently available for allocation on the node.
	AllocatableResources *util.Resources

	// The total amount of resources that are available on the node.
	TotalResources *util.Resources
}

// NodeScore describes the score of a particular node.
type NodeScore struct {
	Node  *NodeInfo
	Score int64
}
