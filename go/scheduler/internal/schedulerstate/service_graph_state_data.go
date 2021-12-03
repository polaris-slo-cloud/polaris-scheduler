package schedulerstate

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/kubernetes/pkg/scheduler/framework"
	"k8s.rainbow-h2020.eu/rainbow/orchestration/pkg/services/servicegraphmanager"
)

const (
	stateKeyBase = "rainbow-h2020/ServiceGraphState/"
)

var (
	_ framework.StateData = (*ServiceGraphStateData)(nil)
)

// ServiceGraphStateData wraps ServiceGraphState for placement in the scheduler's CycleState
type ServiceGraphStateData struct {
	servicegraphmanager.ServiceGraphState
}

// GetServiceGraphStateKey returns the key, under which the pod application's ServiceGraph can be stored in the framework.CycleState.
func GetServiceGraphStateKey(pod *v1.Pod) framework.StateKey {
	// The CycleState is unique for every pod, so using just the base key should be fine.
	return stateKeyBase
}

// NewServiceGraphStateData creates a new instance of ServiceGraphStateData.
func NewServiceGraphStateData(svcGraphState servicegraphmanager.ServiceGraphState) *ServiceGraphStateData {
	return &ServiceGraphStateData{
		ServiceGraphState: svcGraphState,
	}
}

// Clone creates a shallow copy of this ServiceGraphStateData.
func (me *ServiceGraphStateData) Clone() framework.StateData {
	return &ServiceGraphStateData{
		ServiceGraphState: me.ServiceGraphState,
	}
}
