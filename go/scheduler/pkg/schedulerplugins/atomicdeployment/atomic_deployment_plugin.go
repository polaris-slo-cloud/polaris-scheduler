package atomicdeployment

import (
	"context"
	"time"

	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	framework "k8s.io/kubernetes/pkg/scheduler/framework"

	"k8s.rainbow-h2020.eu/rainbow/scheduler/internal/util"
)

const (
	// PluginName is the name of this scheduler plugin.
	PluginName = "AtomicDeployment"

	waitMsec = "5000ms"
)

var (
	_atomicDeploymentPlugin *AtomicDeploymentPlugin

	_ framework.Plugin       = _atomicDeploymentPlugin
	_ framework.PermitPlugin = _atomicDeploymentPlugin
)

// AtomicDeploymentPlugin is a Permit plugin that ensures that all of an application's pods are permitted at the same time or not at all.
type AtomicDeploymentPlugin struct {
	waitDuration time.Duration
}

// New creates a new AtomicDeploymentPlugin instance.
func New(obj runtime.Object, handle framework.Handle) (framework.Plugin, error) {
	waitDuration, err := time.ParseDuration(waitMsec)
	if err != nil {
		return nil, err
	}
	return &AtomicDeploymentPlugin{
		waitDuration: waitDuration,
	}, nil
}

// Name returns the name of this scheduler plugin.
func (me *AtomicDeploymentPlugin) Name() string {
	return PluginName
}

// If this is the initial placement of the ServiceGraph associated with the pod,
// we delay the pod until all other pods have been reserved a node as well (atomic placement).
// In all other cases, the pod is permitted immediately.
//
// Permit is called before binding a pod (and before prebind plugins). Permit
// plugins are used to prevent or delay the binding of a Pod. A permit plugin
// must return success or wait with timeout duration, or the pod will be rejected.
// The pod will also be rejected if the wait timeout or the pod is rejected while
// waiting. Note that if the plugin returns "wait", the framework will wait only
// after running the remaining plugins given that no other plugin rejects the pod.
func (me *AtomicDeploymentPlugin) Permit(ctx context.Context, cycleState *framework.CycleState, pod *core.Pod, nodeName string) (*framework.Status, time.Duration) {
	svcGraphState, noSvcGraphStatus := util.GetServiceGraphFromCycleStateOrStatus(cycleState)
	if noSvcGraphStatus != nil {
		return noSvcGraphStatus, 0
	}

	placementMap, _ := svcGraphState.PlacementMap()
	if !placementMap.IsInitialPlacement() {
		return framework.NewStatus(framework.Success), 0
	}

	allNodes := svcGraphState.ServiceGraphCRD().Spec.Nodes
	for i := range allNodes {
		svcNodeName := allNodes[i].Name
		placement := placementMap.GetKubernetesNodes(svcNodeName)
		if len(placement) == 0 {
			return framework.NewStatus(framework.Wait), me.waitDuration
		}
	}

	return framework.NewStatus(framework.Success), 0
}
