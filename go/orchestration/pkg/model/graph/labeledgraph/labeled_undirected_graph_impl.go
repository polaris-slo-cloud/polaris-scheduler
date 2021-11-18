package labeledgraph

import (
	"math"

	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
)

var (
	_labeledUndirectedGraphImpl *labeledUndirectedGraphImpl

	_ LabeledUndirectedGraph = _labeledUndirectedGraphImpl
)

// Default implementation for LabeledUndirectedGraph
type labeledUndirectedGraphImpl struct {
	*labeledGraphBase
}

func newLabeledUndirectedGraphImpl(nodeFactory LabeledNodeFactoryFn) *labeledUndirectedGraphImpl {
	graph := simple.NewWeightedUndirectedGraph(0, math.Inf(1))
	return &labeledUndirectedGraphImpl{
		labeledGraphBase: newLabeledGraphBase(graph, nodeFactory),
	}
}

func (me *labeledUndirectedGraphImpl) WeightedEdgeBetween(xid, yid int64) graph.WeightedEdge {
	return me.weightedGraph.(graph.WeightedUndirected).WeightedEdgeBetween(xid, yid)
}
