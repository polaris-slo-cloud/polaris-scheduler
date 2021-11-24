package servicegraph

import (
	fogappsCRDs "k8s.rainbow-h2020.eu/rainbow/orchestration/apis/fogapps/v1"
	lg "k8s.rainbow-h2020.eu/rainbow/orchestration/pkg/model/graph/labeledgraph"
)

var (
	_serviceGraphEdgeImpl *serviceGraphEdgeImpl

	_ Edge = _serviceGraphEdgeImpl
)

// The default implementation of a serviceGraph.Edge.
type serviceGraphEdgeImpl struct {
	lg.WeightedEdge
}

// NewEdge is the factory function for creating a new serviceGraph.Edge
var NewEdge lg.WeightedEdgeFactoryFn = func(from, to lg.LabeledNode, weight lg.ComplexEdgeWeight) lg.WeightedEdge {
	return &serviceGraphEdgeImpl{
		WeightedEdge: lg.NewDefaultWeightedEdge(from, to, weight),
	}
}

func (me *serviceGraphEdgeImpl) ServiceLink() *fogappsCRDs.ServiceLink {
	return me.ComplexWeight().(ServiceLinkWeight).ServiceLink()
}
