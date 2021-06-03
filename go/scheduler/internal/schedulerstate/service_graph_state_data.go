package schedulerstate

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/kubernetes/pkg/scheduler/framework"
	"k8s.rainbow-h2020.eu/rainbow/orchestration/pkg/model/graph/servicegraph"
)

const (
	stateKeyBase = "rainbow-h2020/ServiceGraph/"
)

var _ framework.StateData = &ServiceGraphStateData{}

// ServiceGraphStateData wraps ServiceGraph for placement in the scheduler's CycleState
type ServiceGraphStateData struct {
	*servicegraph.ServiceGraph
}

// GetServiceGraphStateKey returns the key, under which the pod application's ServiceGraph can be stored in the framework.CycleState.
func GetServiceGraphStateKey(pod *v1.Pod) framework.StateKey {
	// The CycleState is unique for every pod, so using just the base key should be fine.
	return stateKeyBase
}

// NewServiceGraphStateData creates a new instance of ServiceGraph.
func NewServiceGraphStateData(serviceGraph *servicegraph.ServiceGraph) *ServiceGraphStateData {
	return &ServiceGraphStateData{
		ServiceGraph: serviceGraph,
	}
}

// Clone creates a shallow copy of this ServiceGraph.
func (me *ServiceGraphStateData) Clone() framework.StateData {
	return &ServiceGraphStateData{
		ServiceGraph: me.ServiceGraph,
	}
}
