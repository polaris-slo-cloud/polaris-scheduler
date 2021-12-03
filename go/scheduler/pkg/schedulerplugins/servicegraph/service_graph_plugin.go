package servicegraph

import (
	"context"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/scheduler/framework"
	"k8s.rainbow-h2020.eu/rainbow/orchestration/pkg/services/servicegraphmanager"
	"k8s.rainbow-h2020.eu/rainbow/scheduler/internal/util"
)

const (
	// PluginName is the name of this scheduler plugin.
	PluginName = "ServiceGraph"
)

// ServiceGraphPlugin is a PreFilter plugin that fetches the service graph for a pod's application
// and stores it in the preFilterState.
type ServiceGraphPlugin struct {
	svcGraphManager servicegraphmanager.ServiceGraphManager
}

var _ framework.PreFilterPlugin = &ServiceGraphPlugin{}

// New creates a new RainbowServiceGraph plugin instance.
func New(obj runtime.Object, handle framework.Handle) (framework.Plugin, error) {
	return &ServiceGraphPlugin{
		svcGraphManager: servicegraphmanager.GetServiceGraphManager(),
	}, nil
}

// Name returns the name of this scheduler plugin.
func (me *ServiceGraphPlugin) Name() string {
	return PluginName
}

// PreFilterExtensions returns a PreFilterExtensions interface if the plugin implements one,
// or nil if it does not. A Pre-filter plugin can provide extensions to incrementally
// modify its pre-processed info. The framework guarantees that the extensions
// AddPod/RemovePod will only be called after PreFilter, possibly on a cloned
// CycleState, and may call those functions more than once before calling
// Filter again on a specific node.
func (me *ServiceGraphPlugin) PreFilterExtensions() framework.PreFilterExtensions {
	return nil
}

// PreFilter loads the ServiceGraph for the pod's application and stores it in the CycleState.
func (me *ServiceGraphPlugin) PreFilter(ctx context.Context, state *framework.CycleState, p *v1.Pod) *framework.Status {
	stopwatch := util.NewStopwatch()
	stopwatch.Start()
	state.Write(util.StopwatchStateKey, stopwatch)

	_, err := util.GetPodServiceGraphNodeName(p)
	if err != nil {
		klog.Infof("RainbowServiceGraph: The pod %s is not associated with a RAINBOW application, skipping it.", p.Name)
		return framework.NewStatus(framework.Success)
	}

	serviceGraph, err := me.svcGraphManager.ServiceGraph(p)
	if err != nil {
		klog.Errorf("RainbowServiceGraph plugin error: %s", err.Error())
		return framework.AsStatus(err)
	}

	util.WriteServiceGraphToCycleState(p, state, serviceGraph)
	return framework.NewStatus(framework.Success)
}
