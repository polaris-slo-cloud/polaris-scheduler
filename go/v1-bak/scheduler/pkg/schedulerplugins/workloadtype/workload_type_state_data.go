package workloadtype

import (
	framework "k8s.io/kubernetes/pkg/scheduler/framework"
)

const (
	workloadTypeStateKey = "WorkloadTypePlugin.workloadTypeStateData"
)

var (
	_ framework.StateData = (*workloadTypeStateData)(nil)
)

type workloadTypeStateData struct {
	workloadType string
}

func (me *workloadTypeStateData) Clone() framework.StateData {
	return &workloadTypeStateData{
		workloadType: me.workloadType,
	}
}

// Gets the workloadTypeStateData from the CycleState or returns a framework.Success state if the current pod is not associated with a ServiceGraph
// and, thus, does not have any workloadTypeStateData.
func getWorkloadTypeStateDataOrStatus(cycleState *framework.CycleState) (*workloadTypeStateData, *framework.Status) {
	stateData, err := cycleState.Read(workloadTypeStateKey)
	if err == nil {
		return stateData.(*workloadTypeStateData), nil
	}
	return nil, framework.NewStatus(framework.Success, "Skipping this pod, because it is not associated with a ServiceGraph.")
}
