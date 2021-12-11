package util

import (
	"k8s.io/kubernetes/pkg/scheduler/framework"
	"k8s.rainbow-h2020.eu/rainbow/orchestration/pkg/services/servicegraphmanager"
	"k8s.rainbow-h2020.eu/rainbow/scheduler/internal/schedulerstate"
)

// Gets the ServiceGraphState of the specified pod's application from the CycleState.
//
// If no ServiceGraphState is present in the CycleState of the pod, an error is returned - in such a case, plugins
// that depend on the ServiceGraph should ignore this pod.
//
// This function is thread-safe.
func GetServiceGraphFromCycleState(cycleState *framework.CycleState) (servicegraphmanager.ServiceGraphState, error) {
	cycleState.RLock()
	svcGraphState, err := cycleState.Read(schedulerstate.ServiceGraphStateKey)
	cycleState.RUnlock()
	if err == nil {
		return svcGraphState.(*schedulerstate.ServiceGraphStateData).ServiceGraphState, nil
	}
	return nil, err
}

// Gets the ServiceGraphState of the specified pod's application from the CycleState or,
// if no ServiceGraphState is present in the CycleState of the pod, a framework.Status is returned - in such a case, plugins
// that depend on the ServiceGraph should ignore this pod and return the status.
//
// This function is thread-safe.
func GetServiceGraphFromCycleStateOrStatus(cycleState *framework.CycleState) (servicegraphmanager.ServiceGraphState, *framework.Status) {
	if svcGraphState, err := GetServiceGraphFromCycleState(cycleState); err == nil {
		return svcGraphState, nil
	}
	return nil, framework.NewStatus(framework.Success, "Skipping this pod, because it is not associated with a ServiceGraph.")
}

// Places the ServiceGraphState in the CycleState of the specified pod.
// This function is thread-safe.
func WriteServiceGraphToCycleState(cycleState *framework.CycleState, svcGraphState servicegraphmanager.ServiceGraphState) {
	stateData := schedulerstate.NewServiceGraphStateData(svcGraphState)
	cycleState.Lock()
	defer cycleState.Unlock()
	cycleState.Write(schedulerstate.ServiceGraphStateKey, stateData)
}

// Removes the ServiceGraphState form the CycleState of the specified pod.
func DeleteServiceGraphFromCycleState(cycleState *framework.CycleState) {
	cycleState.Lock()
	defer cycleState.Unlock()
	cycleState.Delete(schedulerstate.ServiceGraphStateKey)
}
