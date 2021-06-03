package util

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/kubernetes/pkg/scheduler/framework"
	"rainbow-h2020.eu/gomod/rainbow-scheduler/pkg/model/graph/servicegraph"
)

// GetServiceGraphFromState gets the ServiceGraph of the specified pod's application from the CycleState.
func GetServiceGraphFromState(pod *v1.Pod, state *framework.CycleState) (*servicegraph.ServiceGraph, error) {
	svcGraph, err := state.Read(servicegraph.GetServiceGraphStateKey(pod))
	if err == nil {
		return svcGraph.(*servicegraph.ServiceGraph), nil
	}
	return nil, err
}
