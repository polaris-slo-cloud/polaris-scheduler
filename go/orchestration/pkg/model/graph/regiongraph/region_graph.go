package regiongraph

import (
	cluster "k8s.rainbow-h2020.eu/rainbow/orchestration/apis/cluster/v1"
	lg "k8s.rainbow-h2020.eu/rainbow/orchestration/pkg/model/graph/labeledgraph"
)

// Node represents a node in a RegionGraph.
// The node's label refers to the name of the Kubernetes node.
type Node interface {
	lg.LabeledNode
}

// Edge represents an edge in a RegionGraph.
// Its weight is determined by a NetworkLinkQoS object.
type Edge interface {
	lg.WeightedEdge

	// Gets the cluster.NetworkLinkQoS that describes this edge.
	NetworkLinkQoS() *cluster.NetworkLinkQoS
}

// NetworkLinkQosWeight wraps a NetworkLinkQoS in a ComplexEdgeWeight object.
type NetworkLinkQosWeight interface {
	lg.ComplexEdgeWeight

	// Gets the NetworkLinkQoS stored by this weight.
	NetworkLinkQoS() *cluster.NetworkLinkQoS
}

// RegionGraph is a representation of a RAINBOW region as a weighted undirected graph.
// The weight of an edge is the number of milliseconds it takes to send a request between the
// two nodes that it connects.
type RegionGraph interface {
	// Graph returns the LabeledUndirectedGraph that models this region.
	Graph() lg.LabeledUndirectedGraph

	// Creates and returns a new Node with a unique ID and the specified label.
	//
	// The node must be added using AddNode()
	NewNode(label string) Node

	// AddNode adds a node to the graph. AddNode panics if
	// the added node ID matches an existing node ID or if the node's Label already exists.
	AddNode(node Node)

	// Creates a new edge from the `from` node to the `to` node and
	// assigns the specified qos as its weight.
	//
	// This edge can be added to the graph using the SetEdge() method.
	NewEdge(from, to Node, qos *cluster.NetworkLinkQoS) Edge

	// Adds the specified edge to this graph.
	SetEdge(edge Edge)

	// NodeByLabel gets the node with the specified label.
	NodeByLabel(label string) Node

	// Edge returns the regiongraph.Edge from the node with
	// the label fromNodeLabel to the node with the label toNodeLabel.
	// If such an edge does not exist, the method returns nil.
	Edge(fromLabel, toLabel string) Edge
}
