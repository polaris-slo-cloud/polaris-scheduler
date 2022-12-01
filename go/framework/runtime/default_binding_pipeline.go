package runtime

import (
	"fmt"
	"sync"

	"github.com/go-logr/logr"
	core "k8s.io/api/core/v1"

	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/client"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/collections"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/pipeline"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/util"
)

var (
	_ pipeline.BindingPipeline = (*DefaultBindingPipeline)(nil)
)

const (
	BindingPipelineStopwatchesStateKey = "polaris-internal.stopwatches.binding"
)

// Default implementation of the BindingPipeline.
type DefaultBindingPipeline struct {
	id                   int
	plugins              *pipeline.BindingPipelinePlugins
	clusterAgentServices pipeline.ClusterAgentServices
	nodesLocker          collections.EntityLocker
	logger               *logr.Logger
}

// A collection of stopwatches for timing various parts of the binding pipeline.
//
// ToDo: find a better way to handle the responsibilities for timing other than this shared struct.
type BindingPipelineStopwatches struct {
	QueueTime       *util.Stopwatch
	NodeLockTime    *util.Stopwatch
	FetchNodeInfo   *util.Stopwatch
	BindingPipeline *util.Stopwatch
	CommitDecision  *util.Stopwatch
}

// Creates a new instance of the default implementation of the BindingPipeline.
func NewDefaultBindingPipeline(
	id int,
	plugins *pipeline.BindingPipelinePlugins,
	clusterAgentServices pipeline.ClusterAgentServices,
	nodesLocker collections.EntityLocker,
) *DefaultBindingPipeline {
	bp := &DefaultBindingPipeline{
		id:                   id,
		plugins:              plugins,
		clusterAgentServices: clusterAgentServices,
		nodesLocker:          nodesLocker,
		logger:               clusterAgentServices.Logger(),
	}
	return bp
}

func NewBindingPipelineStopwatches() *BindingPipelineStopwatches {
	stopwatches := &BindingPipelineStopwatches{
		QueueTime:       util.NewStopwatch(),
		FetchNodeInfo:   util.NewStopwatch(),
		NodeLockTime:    util.NewStopwatch(),
		BindingPipeline: util.NewStopwatch(),
		CommitDecision:  util.NewStopwatch(),
	}
	return stopwatches
}

func (bp *DefaultBindingPipeline) CommitSchedulingDecision(schedCtx pipeline.SchedulingContext, schedDecision *client.ClusterSchedulingDecision) (*client.CommitSchedulingDecisionSuccess, pipeline.Status) {
	stopwatches := bp.getStopwatches(schedCtx)

	stopwatches.NodeLockTime.Start()
	nodeLock := bp.nodesLocker.Lock(schedDecision.NodeName)
	defer nodeLock.Unlock()
	stopwatches.NodeLockTime.Stop()

	// Fetch the NodeInfo.
	stopwatches.FetchNodeInfo.Start()
	updatedNodeInfo, err := bp.fetchNodeInfo(schedCtx, schedDecision)
	stopwatches.FetchNodeInfo.Stop()
	if err != nil {
		return nil, pipeline.NewStatus(pipeline.Unschedulable, "error fetching node information", err.Error())
	}
	decision := &pipeline.SchedulingDecision{
		Pod:        &pipeline.PodInfo{Pod: schedDecision.Pod},
		TargetNode: updatedNodeInfo,
	}

	// Run the pipeline.
	stopwatches.BindingPipeline.Start()
	status := bp.runCheckConflictsPlugins(schedCtx, decision)
	stopwatches.BindingPipeline.Stop()
	if !pipeline.IsSuccessStatus(status) {
		return nil, status
	}

	stopwatches.CommitDecision.Start()
	result, status := bp.commitSchedulingDecision(schedCtx, decision)
	stopwatches.CommitDecision.Stop()
	if result != nil {
		bp.setTimings(result, stopwatches)
	}
	return result, status
}

func (bp *DefaultBindingPipeline) getStopwatches(schedCtx pipeline.SchedulingContext) *BindingPipelineStopwatches {
	stopwatches, ok, err := pipeline.ReadTypedStateData[*BindingPipelineStopwatches](schedCtx, BindingPipelineStopwatchesStateKey)
	if !ok || err != nil {
		panic("could not read BindingPipelineStopwatches from SchedulingContext")
	}
	return stopwatches
}

func (bp *DefaultBindingPipeline) fetchNodeInfo(schedCtx pipeline.SchedulingContext, schedDecision *client.ClusterSchedulingDecision) (*pipeline.NodeInfo, error) {
	// If we cut off the binding pipeline before committing to the orchestrator
	// (to evaluate the scheduler performance only, without the orchestrator commit latency),
	// we need to get the node information from the cache, because the read requests from the
	// orchestrator might also have a high latency, due to the latency induced by the write requests.
	if bp.clusterAgentServices.Config().CutoffBeforeCommit {
		return bp.fetchNodeInfoFromCache(schedDecision)
	}

	clusterClient := bp.clusterAgentServices.ClusterClient()
	nodeName := schedDecision.NodeName
	var node *core.Node
	var podsOnNode []core.Pod
	var nodeFetchErr, podsFetchErr error

	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		node, nodeFetchErr = clusterClient.FetchNode(schedCtx.Context(), nodeName)
		wg.Done()
	}()

	go func() {
		podsOnNode, podsFetchErr = clusterClient.FetchPodsScheduledOnNode(schedCtx.Context(), nodeName)
		wg.Done()
	}()

	wg.Wait()
	if nodeFetchErr != nil {
		return nil, nodeFetchErr
	}
	if podsFetchErr != nil {
		return nil, podsFetchErr
	}

	clusterNode := &client.ClusterNode{
		Node:               node,
		AvailableResources: util.CalculateNodeAvailableResources(node, podsOnNode),
		TotalResources:     util.NewResourcesFromList(node.Status.Capacity),
	}

	return pipeline.NewNodeInfo(clusterClient.ClusterName(), clusterNode), nil
}

func (bp *DefaultBindingPipeline) fetchNodeInfoFromCache(schedDecision *client.ClusterSchedulingDecision) (*pipeline.NodeInfo, error) {
	storeReader := bp.clusterAgentServices.NodesCache().Nodes().ReadLock()
	defer storeReader.Unlock()
	if cachedNode, ok := storeReader.GetByKey(schedDecision.NodeName); ok {
		// Since the cachedNode already considers the pod's resources as reserved,
		// we need to unreserve them locally for the binding pipeline.
		// To this end, we need to create a copy of the cached node and deep copy the AvailableResources,
		// because cached nodes must be treated as immutable.
		podResources := util.CalculateTotalPodResources(schedDecision.Pod)
		cachedNodeCopy := cachedNode.ShallowCopy()
		cachedNodeCopy.Pods = nil
		cachedNodeCopy.QueuedPods = nil
		cachedNodeCopy.AvailableResources = cachedNodeCopy.AvailableResources.DeepCopy()
		cachedNodeCopy.AvailableResources.Add(podResources)
		return pipeline.NewNodeInfo("", cachedNode), nil
	}
	return nil, fmt.Errorf("cannot find node %s in the cache", schedDecision.NodeName)
}

func (bp *DefaultBindingPipeline) runCheckConflictsPlugins(schedCtx pipeline.SchedulingContext, decision *pipeline.SchedulingDecision) pipeline.Status {
	var status pipeline.Status

	for _, plugin := range bp.plugins.CheckConflicts {
		status = plugin.CheckForConflicts(schedCtx, decision)
		if !pipeline.IsSuccessStatus(status) {
			status.SetFailedPlugin(plugin, pipeline.CheckConflictsStage)
			return status
		}
	}

	return status
}

func (bp *DefaultBindingPipeline) commitSchedulingDecision(schedCtx pipeline.SchedulingContext, decision *pipeline.SchedulingDecision) (*client.CommitSchedulingDecisionSuccess, pipeline.Status) {
	clusterSchedDecision := &client.ClusterSchedulingDecision{
		Pod:      decision.Pod.Pod,
		NodeName: decision.TargetNode.Node.Name,
	}

	if bp.clusterAgentServices.Config().CutoffBeforeCommit {
		go bp.clusterAgentServices.ClusterClient().CommitSchedulingDecision(schedCtx.Context(), clusterSchedDecision)
		result := &client.CommitSchedulingDecisionSuccess{
			Namespace: decision.Pod.Pod.Namespace,
			PodName:   decision.Pod.Pod.Name,
			NodeName:  decision.TargetNode.Node.Name,
			Timings:   &client.CommitSchedulingDecisionTimings{},
		}
		return result, pipeline.NewSuccessStatus()
	}

	result, err := bp.clusterAgentServices.ClusterClient().CommitSchedulingDecision(schedCtx.Context(), clusterSchedDecision)
	if err != nil {
		return nil, pipeline.NewStatus(pipeline.Unschedulable, "error committing scheduling decision", err.Error())
	}
	return result, pipeline.NewSuccessStatus()
}

func (bp *DefaultBindingPipeline) setTimings(result *client.CommitSchedulingDecisionSuccess, stopwatches *BindingPipelineStopwatches) {
	result.Timings.QueueTime = stopwatches.QueueTime.Duration().Milliseconds()
	result.Timings.NodeLockTime = stopwatches.NodeLockTime.Duration().Milliseconds()
	result.Timings.FetchNodeInfo = stopwatches.FetchNodeInfo.Duration().Milliseconds()
	result.Timings.BindingPipeline = stopwatches.BindingPipeline.Duration().Milliseconds()
	result.Timings.CommitDecision = stopwatches.CommitDecision.Duration().Milliseconds()
}
