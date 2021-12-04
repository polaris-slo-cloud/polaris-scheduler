package servicegraphmanager

import (
	"sync"

	core "k8s.io/api/core/v1"
	fogappsCRDs "k8s.rainbow-h2020.eu/rainbow/orchestration/apis/fogapps/v1"
	"k8s.rainbow-h2020.eu/rainbow/orchestration/pkg/model/graph/servicegraph"
	"k8s.rainbow-h2020.eu/rainbow/orchestration/pkg/serviceplacement"
	"k8s.rainbow-h2020.eu/rainbow/orchestration/pkg/util"
)

var (
	_ ServiceGraphState = (*serviceGraphStateImpl)(nil)
)

// Contains the ServiceGraph of an application as
// a CRD object and as a traversable graph and allows tracking the
// placement of pods on Kubernetes nodes.
//
// When obtaining a ServiceGraphState from the ServiceGraphManager its internal reference counter
// that tracks the pods it is used by is incremented. When the pod's scheduling cycle ends, it
// must Release() the ServiceGraphState to decrement the reference count and allow the state to
// be removed from the shared cache.
type ServiceGraphState interface {
	// Gets the ServiceGraph graph object.
	// This must be treated as immutable.
	ServiceGraph() servicegraph.ServiceGraph

	// Gets the ServiceGraph CRD instance.
	// This must be treated as immutable.
	ServiceGraphCRD() *fogappsCRDs.ServiceGraph

	// Gets the ServicePlacementMap for the service graph.
	//
	// If the map has not yet been loaded, this method will block until the loading has completed.
	PlacementMap() (serviceplacement.ServiceGraphPlacementMap, error)

	// Gets the map that contains the scheduling priorities of the nodes.
	NodePriorityMap() NodePriorityMap

	// Removes the specified pod from the reference count list.
	Release(pod *core.Pod)
}

// Used to notify the ServiceGraphManager when all references to a ServiceGraphState have been cleared
// and the state is ready for being deleted.
type onServiceGraphStateUnreferencedFn func(svcGraphState *serviceGraphStateImpl)

// Default implementation of ServiceGraphState.
type serviceGraphStateImpl struct {

	// The ServiceGraph as a traversable graph.
	graph servicegraph.ServiceGraph

	// The ServiceGraph CRD instance.
	crd *fogappsCRDs.ServiceGraph

	// The ServiceGraphPlacementMap that indicates on which nodes a service has been placed.
	placementMap util.Future

	// The scheduling priorities of the nodes.
	// This field is initialized lazily.
	nodePriorities NodePriorityMap

	// Tracks the pod names that this object is referenced by in a set like fashion.
	// When the reference count drop to 0 after the initial acquire(),
	// this map will be set to nil to make the ServiceGraphState object unavailable for
	// future acquisition (i.e., the ServiceGraph must be fetched again and a new state must be created).
	referencedBy map[string]bool

	// Synchronizes access to the referencedBy map.
	referencedByMutex sync.Mutex

	// Called when the reference count drops to zero.
	readyForDeletionCallback onServiceGraphStateUnreferencedFn
}

func newServiceGraphStateImpl(
	graph servicegraph.ServiceGraph,
	crd *fogappsCRDs.ServiceGraph,
	placementMap util.Future,
	readyForDeletionCallback onServiceGraphStateUnreferencedFn,
) *serviceGraphStateImpl {
	return &serviceGraphStateImpl{
		graph:                    graph,
		crd:                      crd,
		placementMap:             placementMap,
		referencedBy:             make(map[string]bool),
		referencedByMutex:        sync.Mutex{},
		readyForDeletionCallback: readyForDeletionCallback,
	}
}

func (me *serviceGraphStateImpl) ServiceGraph() servicegraph.ServiceGraph {
	return me.graph
}

func (me *serviceGraphStateImpl) ServiceGraphCRD() *fogappsCRDs.ServiceGraph {
	return me.crd
}

func (me *serviceGraphStateImpl) PlacementMap() (serviceplacement.ServiceGraphPlacementMap, error) {
	placementMap, err := me.placementMap.Get()
	return placementMap.(serviceplacement.ServiceGraphPlacementMap), err
}

func (me *serviceGraphStateImpl) NodePriorityMap() NodePriorityMap {
	if me.nodePriorities == nil {
		me.nodePriorities = NewNodePriorityMapFromServiceGraph(me.graph)
	}
	return me.nodePriorities
}

func (me *serviceGraphStateImpl) Release(pod *core.Pod) {
	me.referencedByMutex.Lock()
	defer me.referencedByMutex.Unlock()

	delete(me.referencedBy, pod.Name)
	if len(me.referencedBy) == 0 {
		me.referencedBy = nil
		me.readyForDeletionCallback(me)
	}
}

// Adds the specified pod to the reference count of this state.
func (me *serviceGraphStateImpl) acquire(pod *core.Pod) bool {
	me.referencedByMutex.Lock()
	defer me.referencedByMutex.Unlock()

	if me.referencedBy != nil {
		me.referencedBy[pod.Name] = true
		return true
	}

	// This state object is ready for deletion, it cannot be acquired anymore.
	return false
}
