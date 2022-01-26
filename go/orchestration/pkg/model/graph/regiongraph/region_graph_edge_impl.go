package regiongraph

import (
	"gonum.org/v1/gonum/graph"
	cluster "k8s.rainbow-h2020.eu/rainbow/orchestration/apis/cluster/v1"
	lg "k8s.rainbow-h2020.eu/rainbow/orchestration/pkg/model/graph/labeledgraph"
)

var (
	_regionGraphEdgeImpl *regionGraphEdgeImpl

	_ Edge = _regionGraphEdgeImpl
)

// The default implementation of a regiongraph.RegionGraphEdge.
type regionGraphEdgeImpl struct {
	lg.WeightedEdge
}

// NewEdge is the factory function for creating a new regiongraph.RegionGraphEdge
var NewEdge lg.WeightedEdgeFactoryFn = func(from, to lg.LabeledNode, weight lg.ComplexEdgeWeight) lg.WeightedEdge {
	return &regionGraphEdgeImpl{
		WeightedEdge: lg.NewDefaultWeightedEdge(from, to, weight),
	}
}

func (me *regionGraphEdgeImpl) NetworkLinkQoS() *cluster.NetworkLinkQoS {
	return me.ComplexWeight().(NetworkLinkQosWeight).NetworkLinkQoS()
}

func (me *regionGraphEdgeImpl) ReversedEdge() graph.Edge {
	return NewEdge(me.To().(Node), me.From().(Node), me.ComplexWeight())
}
