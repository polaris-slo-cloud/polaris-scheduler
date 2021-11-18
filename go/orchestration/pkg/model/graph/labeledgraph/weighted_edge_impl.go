package labeledgraph

import (
	"gonum.org/v1/gonum/graph"
)

var (
	_weightedEdgeImpl *weightedEgeImpl

	_ WeightedEdge = _weightedEdgeImpl
)

// Default implementation of WeightedEdge.
type weightedEgeImpl struct {
	from   graph.Node
	to     graph.Node
	weight ComplexEdgeWeight
}

func newWeightedEdgeImpl(from, to graph.Node, weight ComplexEdgeWeight) *weightedEgeImpl {
	return &weightedEgeImpl{
		from:   from,
		to:     to,
		weight: weight,
	}
}

func (me *weightedEgeImpl) From() graph.Node {
	return me.from
}

func (me *weightedEgeImpl) To() graph.Node {
	return me.to
}

func (me *weightedEgeImpl) ReversedEdge() graph.Edge {
	return &weightedEgeImpl{from: me.to, to: me.from, weight: me.weight}
}

func (me *weightedEgeImpl) ComplexWeight() ComplexEdgeWeight {
	return me.weight
}

func (me *weightedEgeImpl) SetComplexWeight(weight ComplexEdgeWeight) {
	me.weight = weight
}

func (me *weightedEgeImpl) Weight() float64 {
	return me.weight.SimpleWeight()
}
