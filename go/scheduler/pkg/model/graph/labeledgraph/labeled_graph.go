package labeledgraph

import (
	"gonum.org/v1/gonum/graph"
)

// NodePayload represents an arbitrary payload of a LabeledNode
type NodePayload interface{}

// LabeledNodeFactoryFn is used to allow the consumer of a LabeledGraph to configure the type of LabeledNode to be used.
type LabeledNodeFactoryFn func(id int64, label string) LabeledNode

// LabeledNode is a graph node with a label.
// Like a node's ID, its Label must be unique within a graph.
type LabeledNode interface {
	graph.Node

	// Label gets the unique label of this node.
	Label() string

	// Gets the payload of this node.
	Payload() NodePayload

	// Sets the payload of this node.
	SetPayload(payload NodePayload)
}

// LabeledGraph is a weighted, undirected graph that contains LabeledNodes.
// Like a node's ID, its Label must be unique within the graph.
type LabeledGraph interface {
	graph.WeightedUndirected
	graph.NodeRemover
	graph.WeightedEdgeAdder
	graph.EdgeRemover

	// Node returns the node with the given label if it exists
	// in the graph, and nil otherwise.
	NodeByLabel(label string) LabeledNode

	// NewNode returns a new Node with a unique ID and the specified label.
	NewNode(label string) LabeledNode

	// AddNode adds a node to the graph. AddNode panics if
	// the added node ID matches an existing node ID or if the node's Label already exists.
	AddNode(node LabeledNode)
}

// NewLabeledGraph creates an instance of the default LabeledGraph type.
func NewLabeledGraph(nodeFactory LabeledNodeFactoryFn) LabeledGraph {
	return newLabeledGraphImpl(nodeFactory)
}

// NewDefaultLabeledNode creates an instance of the default LabeledNode type.
func NewDefaultLabeledNode(id int64, label string) LabeledNode {
	return newLabeledNodeImpl(id, label)
}
