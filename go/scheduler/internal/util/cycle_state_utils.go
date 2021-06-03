package util

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/kubernetes/pkg/scheduler/framework"
	"k8s.rainbow-h2020.eu/rainbow/orchestration/pkg/model/graph/servicegraph"
	"k8s.rainbow-h2020.eu/rainbow/scheduler/internal/schedulerstate"
)

// GetServiceGraphFromState gets the ServiceGraph of the specified pod's application from the CycleState.
func GetServiceGraphFromState(pod *v1.Pod, cycleState *framework.CycleState) (*servicegraph.ServiceGraph, error) {
	svcGraph, err := cycleState.Read(schedulerstate.GetServiceGraphStateKey(pod))
	if err == nil {
		return svcGraph.(*schedulerstate.ServiceGraphStateData).ServiceGraph, nil
	}
	return nil, err
}

// WriteServiceGraphToState places the ServiceGraph in the CycleState of the specified pod.
func WriteServiceGraphToState(pod *v1.Pod, cycleState *framework.CycleState, serviceGraph *servicegraph.ServiceGraph) {
	stateData := schedulerstate.NewServiceGraphStateData(serviceGraph)
	cycleState.Write(schedulerstate.GetServiceGraphStateKey(pod), stateData)
}
