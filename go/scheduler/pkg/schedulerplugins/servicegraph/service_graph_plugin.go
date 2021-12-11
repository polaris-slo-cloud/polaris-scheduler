package servicegraph

import (
	"context"
	"fmt"
	"time"

	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	klog "k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/scheduler/framework"
	kubequeuesort "k8s.io/kubernetes/pkg/scheduler/framework/plugins/queuesort"

	"k8s.rainbow-h2020.eu/rainbow/orchestration/pkg/kubeutil"
	"k8s.rainbow-h2020.eu/rainbow/orchestration/pkg/services/servicegraphmanager"
	"k8s.rainbow-h2020.eu/rainbow/scheduler/internal/util"
	"k8s.rainbow-h2020.eu/rainbow/scheduler/pkg/schedulerplugins/atomicdeployment"
)

const (
	// PluginName is the name of this scheduler plugin.
	PluginName = "ServiceGraph"
)

var (
	_svcGraphPlugin *ServiceGraphPlugin

	_ framework.Plugin           = _svcGraphPlugin
	_ framework.QueueSortPlugin  = _svcGraphPlugin
	_ framework.PreFilterPlugin  = _svcGraphPlugin
	_ framework.PostFilterPlugin = _svcGraphPlugin
	_ framework.ReservePlugin    = _svcGraphPlugin
	_ framework.PermitPlugin     = _svcGraphPlugin
)

// ServiceGraphPlugin handles all managerial operations related to the ServiceGraph.
//
// Specifically, its tasks are:
// - QueueSort: Loads the ServiceGraph CRD and sort pods according to their position in the service graph.
// - PreFilter: Store the ServiceGraphState in the CycleState.
// - PostFilter: Release the ServiceGraphState if no suitable K8s node was found.
// - Reserve: Record the selected K8s node in the ServiceGraphState.
// - Permit: Wrap AtomicDeploymentPlugin and release the ServiceGraphState when the AtomicDeploymentPlugin gives its "thumbs-up".
//   Note that the ServiceGraphPlugin should be configured as the last Permit plugin in the scheduler configuration.
type ServiceGraphPlugin struct {
	// The original kube-scheduler sorting plugin, which we use after our ServiceGraph node sorting.
	origQueueSort *kubequeuesort.PrioritySort

	// The ServiceGraphManager used for creating the ServiceGraphState.
	svcGraphManager servicegraphmanager.ServiceGraphManager

	// The AtomicDeploymentPlugin, to which we delegate the decision whether the pod may enter the bind phase.
	atomicDeployment *atomicdeployment.AtomicDeploymentPlugin
}

// New creates a new ServiceGraphPlugin instance.
func New(obj runtime.Object, handle framework.Handle) (framework.Plugin, error) {
	origQueueSort, err := kubequeuesort.New(obj, handle)
	if err != nil {
		return nil, err
	}

	atomicDeployment, err := atomicdeployment.New(obj, handle)
	if err != nil {
		return nil, err
	}

	return &ServiceGraphPlugin{
		origQueueSort:    origQueueSort.(*kubequeuesort.PrioritySort),
		svcGraphManager:  servicegraphmanager.GetServiceGraphManager(),
		atomicDeployment: atomicDeployment.(*atomicdeployment.AtomicDeploymentPlugin),
	}, nil
}

// Name returns the name of this scheduler plugin.
func (me *ServiceGraphPlugin) Name() string {
	return PluginName
}

// Less prioritizes pods, based on the position in their ServiceGraph.
// If the pods are not associated with a ServiceGraph or are associated with different ServiceGraphs
// we fall back to the original PrioritySort.
//
// Returns true if podA should be scheduled before podB.
func (me *ServiceGraphPlugin) Less(podA *framework.QueuedPodInfo, podB *framework.QueuedPodInfo) bool {
	// ToDo: Add an AcquireServiceGraphStateAsync() method to the ServiceGraphManager.
	// This will allow the fetching to take place in the background if podA and podB are not part of the same serviceGraph.
	svcGraphState := me.getCommonServiceGraphState(podA, podB)
	if svcGraphState == nil {
		return me.origQueueSort.Less(podA, podB)
	}

	svcNodeAName, ok := util.GetPodServiceGraphNodeName(podA.Pod)
	if !ok {
		return me.origQueueSort.Less(podA, podB)
	}
	svcNodeBName, ok := util.GetPodServiceGraphNodeName(podB.Pod)
	if !ok {
		return me.origQueueSort.Less(podA, podB)
	}

	nodePriorities := svcGraphState.NodePriorityMap()
	priorityA := nodePriorities.NodePriority(svcNodeAName)
	priorityB := nodePriorities.NodePriority(svcNodeBName)

	if priorityA == -1 || priorityB == -1 || priorityA == priorityB {
		return me.origQueueSort.Less(podA, podB)
	}
	return priorityA < priorityB
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

// PreFilter stores the ServiceGraphState for the pod's ServiceGraph in the CycleState
// if the pod is associated with a ServiceGraph and with a node within that ServiceGraph.
//
// This ensures that all future plugins can assume that the pod has a valid ServiceGraph and node reference if
// it CycleState contains the ServiceGraphState.
func (me *ServiceGraphPlugin) PreFilter(ctx context.Context, cycleState *framework.CycleState, pod *core.Pod) *framework.Status {
	// Time how long this pod takes to be scheduled.
	stopwatch := util.NewStopwatch()
	stopwatch.Start()

	svcGraphState, err := me.svcGraphManager.AcquireServiceGraphState(pod)
	if err != nil {
		return framework.AsStatus(err)
	}
	if svcGraphState == nil {
		// If both err and svcGraphState are nil, this pod is not associated with a ServiceGraph.
		return framework.NewStatus(framework.Success, fmt.Sprintf("The pod %s is not associated with a ServiceGraph", pod.Name))
	}

	svcGraphNodeName, ok := util.GetPodServiceGraphNodeName(pod)
	if !ok || svcGraphState.ServiceGraph().NodeByLabel(svcGraphNodeName) == nil {
		// If there is a ServiceGraph, but no valid reference to a node within that ServiceGraph, we signal an error about this pod.
		svcGraphState.Release(pod)
		return framework.AsStatus(fmt.Errorf("The pod %s is not associated with a node within its ServiceGraph", pod.Name))
	}

	util.WriteServiceGraphToCycleState(cycleState, svcGraphState)
	cycleState.Lock()
	cycleState.Write(util.StopwatchStateKey, stopwatch)
	cycleState.Unlock()
	return framework.NewStatus(framework.Success)
}

// PostFilter is called when no suitable K8s node was found during the Filter phase, so we
// release the ServiceGraphState here.
//
// A PostFilter plugin should return one of the following statuses:
// - Unschedulable: the plugin gets executed successfully but the pod cannot be made schedulable.
// - Success: the plugin gets executed successfully and the pod can be made schedulable.
// - Error: the plugin aborts due to some internal error.
//
// Informational plugins should be configured ahead of other ones, and always return Unschedulable status.
// Optionally, a non-nil PostFilterResult may be returned along with a Success status. For example,
// a preemption plugin may choose to return nominatedNodeName, so that framework can reuse that to update the
// preemptor pod's .spec.status.nominatedNodeName field.
func (me *ServiceGraphPlugin) PostFilter(ctx context.Context, cycleState *framework.CycleState, pod *core.Pod, filteredNodeStatusMap framework.NodeToStatusMap) (*framework.PostFilterResult, *framework.Status) {
	if svcGraphState, err := util.GetServiceGraphFromCycleState(cycleState); err == nil {
		me.releaseServiceGraphState(ctx, cycleState, pod, svcGraphState)
	}
	return nil, framework.NewStatus(framework.Unschedulable)
}

// Reserve assigns the Kubernetes node to the node in the ServiceGraph
func (me *ServiceGraphPlugin) Reserve(ctx context.Context, cycleState *framework.CycleState, pod *core.Pod, nodeName string) *framework.Status {
	svcGraphState, noSvcGraphStatus := util.GetServiceGraphFromCycleStateOrStatus(cycleState)
	if noSvcGraphStatus != nil {
		return noSvcGraphStatus
	}

	placementMap, err := svcGraphState.PlacementMap()
	if err != nil {
		return framework.NewStatus(framework.Error, err.Error())
	}

	svcNodeLabel, _ := util.GetPodServiceGraphNodeName(pod)
	placementMap.SetKubernetesNodes(svcNodeLabel, func(curr []string) []string {
		// We append the K8s node to a new list, if it is not already present (i.e., another instance of the service has already been scheduled there).
		newList := make([]string, len(curr), len(curr)+1)
		k8sNodeAlreadyPresent := false
		for i, existingNode := range curr {
			newList[i] = existingNode
			k8sNodeAlreadyPresent = k8sNodeAlreadyPresent || existingNode == nodeName
		}
		if !k8sNodeAlreadyPresent {
			newList = append(newList, nodeName)
		}
		return newList
	})

	return framework.NewStatus(framework.Success)
}

// Unreserve removes the Kubernetes node from the node in the ServiceGraph and releases its ServiceGraphState.
// This method must not fail
func (me *ServiceGraphPlugin) Unreserve(ctx context.Context, cycleState *framework.CycleState, pod *core.Pod, nodeName string) {
	svcGraphState, err := util.GetServiceGraphFromCycleState(cycleState)
	if err != nil {
		return
	}

	defer me.releaseServiceGraphState(ctx, cycleState, pod, svcGraphState)

	// ToDo: Change PlacementMap to track number of pods placed on a K8s node.
	// Otherwise, we might remove the node from the list, even though an existing pod is running on it.
	//
	// placementMap, err := svcGraphState.PlacementMap()
	// if err != nil {
	// 	return
	// }
	// svcNodeLabel, _ := util.GetPodServiceGraphNodeName(pod)
	// placementMap.SetKubernetesNodes(svcNodeLabel, func(curr []string) []string {
	// 	newList := make([]string, len(curr) - 1)
	// 	indexNew := 0
	// 	for _, k8sNode := range curr {
	// 		if k8sNode != nodeName {
	// 			newList[indexNew] = k8sNode
	// 			indexNew++
	// 		}
	// 	}
	// 	return newList
	// })
}

// Permit calls atomicDeployment.Permit() and releases the ServiceGraphState if
// atomicDeployment permits the pod.
//
// Permit is called before binding a pod (and before prebind plugins). Permit
// plugins are used to prevent or delay the binding of a Pod. A permit plugin
// must return success or wait with timeout duration, or the pod will be rejected.
// The pod will also be rejected if the wait timeout or the pod is rejected while
// waiting. Note that if the plugin returns "wait", the framework will wait only
// after running the remaining plugins given that no other plugin rejects the pod.
func (me *ServiceGraphPlugin) Permit(ctx context.Context, cycleState *framework.CycleState, pod *core.Pod, nodeName string) (*framework.Status, time.Duration) {
	status, duration := me.atomicDeployment.Permit(ctx, cycleState, pod, nodeName)
	if status == nil || status.IsSuccess() {
		if svcGraphState, err := util.GetServiceGraphFromCycleState(cycleState); err == nil {
			me.releaseServiceGraphState(ctx, cycleState, pod, svcGraphState)
		}
		me.readStopwatch(cycleState, pod)
	}
	return status, duration
}

func (me *ServiceGraphPlugin) releaseServiceGraphState(ctx context.Context, cycleState *framework.CycleState, pod *core.Pod, svcGraphState servicegraphmanager.ServiceGraphState) {
	util.DeleteServiceGraphFromCycleState(cycleState)
	svcGraphState.Release(pod)
}

// Gets the ServiceGraphState for the two pods, if both are associated with the same ServiceGraph and both have a ServiceGraphNode reference, otherwise it returns nil.
func (me *ServiceGraphPlugin) getCommonServiceGraphState(podA *framework.QueuedPodInfo, podB *framework.QueuedPodInfo) servicegraphmanager.ServiceGraphState {
	if podA.Pod.Namespace != podB.Pod.Namespace {
		return nil
	}
	svcGraphA, ok := kubeutil.GetLabel(podA.Pod, kubeutil.LabelRefServiceGraph)
	if !ok {
		return nil
	}
	if _, ok := kubeutil.GetLabel(podA.Pod, kubeutil.LabelRefServiceGraphNode); !ok {
		return nil
	}
	svcGraphB, ok := kubeutil.GetLabel(podB.Pod, kubeutil.LabelRefServiceGraph)
	if !ok || svcGraphA != svcGraphB {
		return nil
	}
	if _, ok := kubeutil.GetLabel(podB.Pod, kubeutil.LabelRefServiceGraphNode); !ok {
		return nil
	}

	// It does not matter which pod we use to acquire the ServiceGraphState, because they will share the same one
	// and the second pod will be added to the reference count in the PreFilter stage.
	svcGraphState, err := me.svcGraphManager.AcquireServiceGraphState(podA.Pod)
	if err != nil {
		klog.Errorf("Could not acquire ServiceGraph for pod %s.%s Error: %s", podA.Pod.Namespace, podA.Pod.Name, err)
		return nil
	}
	return svcGraphState
}

func (me *ServiceGraphPlugin) readStopwatch(cycleState *framework.CycleState, pod *core.Pod) {
	if stateData, err := cycleState.Read(util.StopwatchStateKey); err == nil {
		stopwatch := stateData.(*util.Stopwatch)
		stopwatch.Stop()
		durationMs := float64(stopwatch.Duration().Microseconds()) / 1000
		klog.Infof("Scheduling pod %s.%s took %f ms", pod.Namespace, pod.Name, durationMs)
	}
}
