package labeledgraph

import (
	"fmt"

	"gonum.org/v1/gonum/graph"
)

var (
	_labeledGraphImpl *labeledGraphBase

	_ LabeledGraph = _labeledGraphImpl
)

// Combines all graph base interfaces needed for our weighted graphs.
type weightedGraph interface {
	graph.Weighted
	graph.NodeAdder
	graph.NodeRemover
	graph.WeightedEdgeAdder
	graph.EdgeRemover
}

// labeledGraphBase is the base implementation of LabeledGraph
type labeledGraphBase struct {
	weightedGraph
	nodeIdsByLabel map[string]int64
	createNewNode  LabeledNodeFactoryFn
	createNewEdge  WeightedEdgeFactoryFn
}

// Creates a new labeledGraphBase object with the specified graph and node factory.
func newLabeledGraphBase(graph weightedGraph, nodeFactory LabeledNodeFactoryFn, edgeFactory WeightedEdgeFactoryFn) *labeledGraphBase {
	return &labeledGraphBase{
		weightedGraph:  graph,
		nodeIdsByLabel: make(map[string]int64),
		createNewNode:  nodeFactory,
		createNewEdge:  edgeFactory,
	}
}

func (me *labeledGraphBase) NodeByLabel(label string) LabeledNode {
	nodeID, exists := me.nodeIdsByLabel[label]
	if exists {
		return me.Node(nodeID).(LabeledNode)
	}
	return nil
}

func (me *labeledGraphBase) NewNode(label string) LabeledNode {
	simpleNode := me.weightedGraph.NewNode()
	labeledNode := me.createNewNode(simpleNode.ID(), label)
	return labeledNode
}

func (me *labeledGraphBase) AddNode(node LabeledNode) {
	label := node.Label()
	if _, exists := me.nodeIdsByLabel[label]; exists {
		panic(fmt.Sprintf("LabeledGraph: The node Label already exists: %s", label))
	}
	me.weightedGraph.AddNode(node)
	me.nodeIdsByLabel[label] = node.ID()
}

func (me *labeledGraphBase) NewWeightedEdge(from, to LabeledNode, weight ComplexEdgeWeight) WeightedEdge {
	return me.createNewEdge(from, to, weight)
}

func (me *labeledGraphBase) SetWeightedEdge(edge WeightedEdge) {
	me.weightedGraph.SetWeightedEdge(edge)
}
