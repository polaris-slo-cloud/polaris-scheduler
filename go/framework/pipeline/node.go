package pipeline

import (
	core "k8s.io/api/core/v1"
)

// NodeInfo stores a Node and additional pre-computed scheduling-relevant information about it.
type NodeInfo struct {

	// The Node described by this NodeInfo.
	Node *core.Node
}

// NodeScore describes the score of a particular node.
type NodeScore struct {
	Node  *NodeInfo
	Score int64
}
