package labeledgraph

import (
	"fmt"
	"math"

	"gonum.org/v1/gonum/graph/simple"
)

var (
	_labeledGraphImpl *labeledGraphImpl

	_ LabeledGraph = _labeledGraphImpl
)

// labeledGraphImpl is the main implementation of RegionGraph
type labeledGraphImpl struct {
	*simple.WeightedUndirectedGraph
	nodeIdsByLabel map[string]int64
	createNewNode  LabeledNodeFactoryFn
}

func newLabeledGraphImpl(nodeFactory LabeledNodeFactoryFn) *labeledGraphImpl {
	return &labeledGraphImpl{
		WeightedUndirectedGraph: simple.NewWeightedUndirectedGraph(0, math.Inf(1)),
		nodeIdsByLabel:          make(map[string]int64),
		createNewNode:           nodeFactory,
	}
}

func (me *labeledGraphImpl) NodeByLabel(label string) LabeledNode {
	nodeID, exists := me.nodeIdsByLabel[label]
	if exists {
		return me.Node(nodeID).(LabeledNode)
	}
	return nil
}

func (me *labeledGraphImpl) NewNode(label string) LabeledNode {
	simpleNode := me.WeightedUndirectedGraph.NewNode()
	labeledNode := me.createNewNode(simpleNode.ID(), label)
	return labeledNode
}

func (me *labeledGraphImpl) AddNode(node LabeledNode) {
	label := node.Label()
	if _, exists := me.nodeIdsByLabel[label]; exists {
		panic(fmt.Sprintf("LabeledGraph: The node Label already exists: %s", label))
	}
	me.WeightedUndirectedGraph.AddNode(node)
	me.nodeIdsByLabel[label] = node.ID()
}
