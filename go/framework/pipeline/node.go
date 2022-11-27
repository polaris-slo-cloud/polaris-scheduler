package pipeline

import (
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/client"
)

// NodeInfo stores a ClusterNode and additional scheduling-relevant information about it.
type NodeInfo struct {

	// The Node described by this NodeInfo.
	Node *client.ClusterNode `json:"node" yaml:"node"`

	// The accumulated score computed by the Score plugins of the sampling pipeline.
	// This is nil if no sampling score plugins are configured.
	SamplingScore *SamplingScore

	// The name of the cluster that the node is part of.
	ClusterName string `json:"clusterName" yaml:"clusterName"`
}

// NodeScore describes the score of a particular node.
type NodeScore struct {
	Node  *NodeInfo
	Score int64
}

// SamplingScore describes the accumulated score from all sampling score plugins.
type SamplingScore struct {
	// The accumulated score of all sampling score plugins.
	AccumulatedScore int64 `json:"accumulatedScore" yaml:"accumulatedScore"`

	// The number score plugins that contributed to the accumulated score.
	ScorePluginsCount int `json:"scorePluginsCount" yaml:"scorePluginsCount"`
}

// Creates a new NodeInfo object and computes its resources.
func NewNodeInfo(clusterName string, node *client.ClusterNode) *NodeInfo {
	return &NodeInfo{
		Node:        node,
		ClusterName: clusterName,
	}
}
