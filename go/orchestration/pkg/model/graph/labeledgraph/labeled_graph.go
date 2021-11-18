package labeledgraph

import (
	"gonum.org/v1/gonum/graph"
)

// ToDo: Add unit tests that leverage those provided by gonum.
// See https://github.com/gonum/gonum/blob/master/graph/simple/directed_test.go

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

// ComplexEdgeWeight represents a complex object that constitutes the weight of a WeightedEdge.
// This interface needs to be implemented by objects that will be used as WeightedEdge weights.
//
// In case a simple float value should be used as the weight, use NewComplexEdgeWeightFromFloat()
// to wrap it in a ComplexEdgeWeight object.
type ComplexEdgeWeight interface {
	// SimpleWeight returns the weight of the edge, as a single floating-point number.
	// This may be the value of a single field of the ComplexEdgeWeight or an aggregation of multiple fields.
	SimpleWeight() float64
}

// WeightedEdge represents an edge in a LabeledGraph with a complex weight.
type WeightedEdge interface {
	graph.WeightedEdge

	// Gets the ComplexEdgeWeight object of this WeightedEdge.
	ComplexWeight() ComplexEdgeWeight

	// Sets the ComplexEdgeWeight object of this WeightedEdge.
	SetComplexWeight(weight ComplexEdgeWeight)
}

// LabeledGraph is a weighted graph that contains LabeledNodes.
// Like a node's ID, its Label must be unique within the graph.
type LabeledGraph interface {
	graph.Weighted
	graph.NodeRemover
	graph.EdgeRemover

	// Node returns the node with the given label if it exists
	// in the graph, and nil otherwise.
	NodeByLabel(label string) LabeledNode

	// NewNode returns a new Node with a unique ID and the specified label.
	NewNode(label string) LabeledNode

	// AddNode adds a node to the graph. AddNode panics if
	// the added node ID matches an existing node ID or if the node's Label already exists.
	AddNode(node LabeledNode)

	// NewWeightedEdge creates a new WeightedEdge from the `from` to the `to` node.
	// This edge can be added to the graph using the SetWeightedEdge() method.
	NewWeightedEdge(from, to LabeledNode, weight ComplexEdgeWeight) WeightedEdge

	// Adds the specified edge to this graph.
	SetWeightedEdge(edge WeightedEdge)
}

// LabeledUndirectedGraph is a weighted, undirected graph that contains LabeledNodes.
type LabeledUndirectedGraph interface {
	graph.WeightedUndirected
	LabeledGraph
}

// LabeledDirectedGraph is a weighted, directed graph that contains LabeledNodes.
type LabeledDirectedGraph interface {
	graph.WeightedDirected
	LabeledGraph
}

// NewLabeledUndirectedGraph creates an instance of the default LabeledUndirectedGraph type.
func NewLabeledUndirectedGraph(nodeFactory LabeledNodeFactoryFn) LabeledUndirectedGraph {
	return newLabeledUndirectedGraphImpl(nodeFactory)
}

// NewLabeledDirectedGraph creates an instance of the default LabeledDirectedGraph type.
func NewLabeledDirectedGraph(nodeFactory LabeledNodeFactoryFn) LabeledDirectedGraph {
	return newLabeledDirectedGraphImpl(nodeFactory)
}

// NewDefaultLabeledNode creates an instance of the default LabeledNode type.
func NewDefaultLabeledNode(id int64, label string) LabeledNode {
	return newLabeledNodeImpl(id, label)
}
