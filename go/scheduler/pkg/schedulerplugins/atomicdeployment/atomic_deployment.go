package atomicdeployment

import (
	"context"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	framework "k8s.io/kubernetes/pkg/scheduler/framework"
	"rainbow-h2020.eu/gomod/rainbow-scheduler/pkg/util"
)

const (
	// PluginName is the name of this scheduler plugin.
	PluginName = "RainbowAtomicDeployment"
)

// RainbowAtomicDeployment is a Permit plugin that ensures that all of an application's pods are permitted at the same time or not at all.
type RainbowAtomicDeployment struct {
}

var _ framework.PermitPlugin = &RainbowAtomicDeployment{}

// New creates a new RainbowAtomicDeployment plugin instance.
func New(obj runtime.Object, handle framework.Handle) (framework.Plugin, error) {
	return &RainbowAtomicDeployment{}, nil
}

// Name returns the name of this scheduler plugin.
func (me *RainbowAtomicDeployment) Name() string {
	return PluginName
}

// Permit is called before binding a pod (and before prebind plugins). Permit
// plugins are used to prevent or delay the binding of a Pod. A permit plugin
// must return success or wait with timeout duration, or the pod will be rejected.
// The pod will also be rejected if the wait timeout or the pod is rejected while
// waiting. Note that if the plugin returns "wait", the framework will wait only
// after running the remaining plugins given that no other plugin rejects the pod.
func (me *RainbowAtomicDeployment) Permit(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) (*framework.Status, time.Duration) {
	if stateData, err := state.Read(util.StopwatchStateKey); err == nil {
		stopwatch := stateData.(*util.Stopwatch)
		stopwatch.Stop()
		durationMs := float64(stopwatch.Duration().Microseconds()) / 1000
		klog.Infof("Scheduling pod %s.%s took %f ms", pod.Namespace, pod.Name, durationMs)
	}

	return framework.NewStatus(framework.Success, ""), 0
}
