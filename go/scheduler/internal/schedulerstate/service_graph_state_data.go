package schedulerstate

import (
	"k8s.io/kubernetes/pkg/scheduler/framework"
	"k8s.rainbow-h2020.eu/rainbow/orchestration/pkg/services/servicegraphmanager"
)

const (
	// The key that under which the ServiceGraphState is stored in a pod's CycleState.
	ServiceGraphStateKey = "rainbow-h2020/ServiceGraphState"
)

var (
	_ framework.StateData = (*ServiceGraphStateData)(nil)
)

// ServiceGraphStateData wraps ServiceGraphState for placement in the scheduler's CycleState
type ServiceGraphStateData struct {
	servicegraphmanager.ServiceGraphState
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
