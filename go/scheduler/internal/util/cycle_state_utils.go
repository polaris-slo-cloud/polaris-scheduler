package util

import (
	core "k8s.io/api/core/v1"
	"k8s.io/kubernetes/pkg/scheduler/framework"
	"k8s.rainbow-h2020.eu/rainbow/orchestration/pkg/services/servicegraphmanager"
	"k8s.rainbow-h2020.eu/rainbow/scheduler/internal/schedulerstate"
)

// GetServiceGraphFromCycleState gets the ServiceGraphState of the specified pod's application from the CycleState.
//
// If no ServiceGraphState is present in the CycleState of the pod, an error is returned - in such a case, plugins
// that depend on the ServiceGraph should ignore this pod.
//
// This function is thread-safe.
func GetServiceGraphFromCycleState(pod *core.Pod, cycleState *framework.CycleState) (servicegraphmanager.ServiceGraphState, error) {
	key := schedulerstate.GetServiceGraphStateKey(pod)
	cycleState.RLock()
	svcGraphState, err := cycleState.Read(key)
	cycleState.RUnlock()
	if err == nil {
		return svcGraphState.(*schedulerstate.ServiceGraphStateData).ServiceGraphState, nil
	}
	return nil, err
}

// WriteServiceGraphToCycleState places the ServiceGraphState in the CycleState of the specified pod.
// This function is thread-safe.
func WriteServiceGraphToCycleState(pod *core.Pod, cycleState *framework.CycleState, svcGraphState servicegraphmanager.ServiceGraphState) {
	stateData := schedulerstate.NewServiceGraphStateData(svcGraphState)
	key := schedulerstate.GetServiceGraphStateKey(pod)

	cycleState.Lock()
	defer cycleState.Unlock()
	cycleState.Write(key, stateData)
}
