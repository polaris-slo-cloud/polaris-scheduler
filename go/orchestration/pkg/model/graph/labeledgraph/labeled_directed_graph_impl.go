package labeledgraph

import (
	"math"

	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
)

var (
	_labeledDirectedGraphImpl *labeledDirectedGraphImpl

	_ LabeledDirectedGraph = _labeledDirectedGraphImpl
)

// Default implementation for LabeledDirectedGraph
type labeledDirectedGraphImpl struct {
	*labeledGraphBase
}

func newLabeledDirectedGraphImpl(nodeFactory LabeledNodeFactoryFn) *labeledDirectedGraphImpl {
	graph := simple.NewWeightedDirectedGraph(0, math.Inf(1))
	return &labeledDirectedGraphImpl{
		labeledGraphBase: newLabeledGraphBase(graph, nodeFactory),
	}
}

func (me *labeledDirectedGraphImpl) HasEdgeFromTo(uid, vid int64) bool {
	return me.weightedGraph.(graph.WeightedDirected).HasEdgeFromTo(uid, vid)
}

func (me *labeledDirectedGraphImpl) To(id int64) graph.Nodes {
	return me.weightedGraph.(graph.WeightedDirected).To(id)
}
