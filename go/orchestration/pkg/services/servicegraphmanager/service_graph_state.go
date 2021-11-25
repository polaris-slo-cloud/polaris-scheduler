package servicegraphmanager

import (
	fogappsCRDs "k8s.rainbow-h2020.eu/rainbow/orchestration/apis/fogapps/v1"
	"k8s.rainbow-h2020.eu/rainbow/orchestration/pkg/model/graph/servicegraph"
	"k8s.rainbow-h2020.eu/rainbow/orchestration/pkg/serviceplacement"
	"k8s.rainbow-h2020.eu/rainbow/orchestration/pkg/util"
)

var (
	_serviceGraphStateImpl *serviceGraphStateImpl

	_ ServiceGraphState = _serviceGraphStateImpl
)

// Contains the ServiceGraph of an application as
// a CRD object and as a traversable graph and allows tracking the
// placement of pods on Kubernetes nodes
type ServiceGraphState interface {
	// Gets the ServiceGraph graph object.
	// This must be treated as immutable.
	ServiceGraph() *servicegraph.ServiceGraph

	// Gets the ServiceGraph CRD instance.
	// This must be treated as immutable.
	ServiceGraphCRD() *fogappsCRDs.ServiceGraph

	// Gets the ServicePlacementMap for the service graph.
	//
	// If the map has not yet been loaded, this method will block until the loading has completed.
	PlacementMap() (serviceplacement.ServiceGraphPlacementMap, error)
}

type serviceGraphStateImpl struct {
	graph        *servicegraph.ServiceGraph
	crd          *fogappsCRDs.ServiceGraph
	placementMap util.Future
}

func newServiceGraphStateImpl(graph *servicegraph.ServiceGraph, crd *fogappsCRDs.ServiceGraph, placementMap util.Future) *serviceGraphStateImpl {
	return &serviceGraphStateImpl{
		graph:        graph,
		crd:          crd,
		placementMap: placementMap,
	}
}

func (me *serviceGraphStateImpl) ServiceGraph() *servicegraph.ServiceGraph {
	return me.graph
}

func (me *serviceGraphStateImpl) ServiceGraphCRD() *fogappsCRDs.ServiceGraph {
	return me.crd
}

func (me *serviceGraphStateImpl) PlacementMap() (serviceplacement.ServiceGraphPlacementMap, error) {
	placementMap, err := me.placementMap.Get()
	return placementMap.(serviceplacement.ServiceGraphPlacementMap), err
}
