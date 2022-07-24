package pipeline

import (
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/config"
)

const (
	// The minimum node score that may be returned by a ScorePlugin (after NormalizeScore).
	MinNodeScore int64 = 0

	// The maximum node score that may be returned by a ScorePlugin (after NormalizeScore).
	MaxNodeScore int64 = 100

	SortStage        = "Sort"
	SampleNodesStage = "SampleNodes"
	PreFilterStage   = "PreFilter"
	FilterStage      = "Filter"
	PreScoreStage    = "PreScore"
	ScoreStage       = "Score"
	ReserveStage     = "Reserve"
)

// Plugin is the parent interface for all Polaris scheduling pipeline plugins
//
// The Polaris scheduling pipeline consists of the following stages:
// - Sort (one plugin only)
// - SampleNodes (one plugin only)
// - PreFilter
// - Filter
// - PreScore
// - Score
// - Reserve
//
// The stages from PreFilter up to (including) Reserve are called the "Decision Pipeline".
// For Decision Pipeline plugins it is common to tie into multiple stages of the pipeline.
// For each pipeline instance only a single instance of each plugin will be created, even if it ties
// into multiple stages (e.g., PreFilter, Filter, and Score).
//
// Multiple Decision Pipeline instances may execute in parallel, but each instance will
// execute on a single goroutine and only be traversed by a single pod at a time.
type Plugin interface {
	Name() string
}

// A SortPlugin is used to establish the order, in which incoming pods will be handled by the scheduling pipeline.
type SortPlugin interface {
	Plugin

	// Less returns true if podA should be scheduled before podB.
	// Otherwise, it returns false.
	Less(podA *QueuedPodInfo, podB *QueuedPodInfo) bool
}

// A SampleNodesPlugin is used to obtain a sample of nodes from the entire supercluster as hosting candidates for the pod.
// This plugin is called when a pod enters the scheduling pipeline.
type SampleNodesPlugin interface {
	Plugin

	// Samples nodes across the entire supercluster to act has hosting candidates for the pod.
	//
	// Returns an array of NodeInfos that describe the sampled nodes and a Status.
	SampleNodes(ctx SchedulingContext, podInfo *PodInfo, config *config.SchedulerConfig) ([]*NodeInfo, Status)
}

// A PreFilterPlugin is called once per Pod and can be used to pre-compute information that will be needed by a FilterPlugin.
// PreFilterPlugins are called after nodes have been sampled.
type PreFilterPlugin interface {
	Plugin

	// PreFilter is called once per Pod and can be used to pre-compute information that will be needed by a FilterPlugin.
	//
	// All PreFilterPlugins must return Success, otherwise the pod is marked as Unschedulable.
	PreFilter(ctx SchedulingContext, podInfo *PodInfo) Status
}

// A FilterPlugin determines if a particular node is suitable for hosting a pod.
// FilterPlugins are called after the PreFilterState.
// At the beginning of the Filter stage all nodes from the SampleNodes stage are used. This list
// may be reduced by every FilterPlugin. Once a node is deemed to be unsuitable to host a pod,
// it is not passed to any other FilterPlugin.
type FilterPlugin interface {
	Plugin

	// Filter is called to determine if the pod described by podInfo can be hosted on the node described by NodeInfo.
	//
	// Returns a "Success" Status is the node can host the pod, an "Unschedulable" Status if this is not the case,
	// or an "InternalError" Status if an unexpected error occurred during evaluation.
	Filter(ctx SchedulingContext, podInfo *PodInfo, nodeInfo *NodeInfo) Status
}

// A PreFilterPlugin is called once per Pod, after the Filter stage has completed, and can be used to
// pre-compute information that will be needed by a ScorePlugin.
type PreScorePlugin interface {
	Plugin

	// PreScore is called once per Pod and can be used to pre-compute information that will be needed by a ScorePlugin.
	// eligibleNodes contains all nodes that have been deemed suitable to host the pod by the Filter stage plugins.
	//
	// All PreScorePlugins must return Success, otherwise the pod is marked as Unschedulable.
	PreScore(ctx SchedulingContext, podInfo *PodInfo, eligibleNodes []*NodeInfo) Status
}

// Allows defining optional actions supported by a ScorePlugin
type ScoreExtensions interface {
	// Called to normalize the node scores returned by the associated ScorePlugin to a range between MinNodeScore and MaxNodeScore.
	// This method should updated the scores list and return a Success Status.
	NormalizeScores(ctx SchedulingContext, podInfo *PodInfo, scores []NodeScore) Status
}

// A ScorePlugin has to assign a score to every node that came out of the Filter stage.
// The scores from all ScorePlugins are accumulated by the scheduling pipeline and used to rank the eligible nodes.
//
// The node with the highest score is assigned to host the pod.
// If multiple nodes have the same high score, a random node is picked from this set of winners.
type ScorePlugin interface {
	Plugin

	// Score needs to compute a score for the node that describes "how suitable" it is to host the pod.
	// These scores are used to rank the nodes.
	// All ScorePlugins must return a Success Status, otherwise the pod is rejected.
	Score(ctx SchedulingContext, podInfo *PodInfo, nodeInfo *NodeInfo) (int64, Status)

	// Returns the ScoreExtensions, if they are implemented by this plugin, otherwise nil.
	ScoreExtensions() ScoreExtensions
}

// A ReservePlugin is called after the scheduling pipeline has chosen the final target node after the Score stage.
// It may be used to update 3rd party data structures.
type ReservePlugin interface {
	Plugin

	// Reserve is called after the scheduling pipeline has chosen the final target node after the Score stage.
	// It may be used to update 3rd party data structures.
	// If any ReservePlugin returns a non Success Status, the pod will not be scheduled to that node and
	// Unreserve will be called on all ReservePlugins.
	Reserve(ctx SchedulingContext, podInfo *PodInfo, targetNode *NodeInfo) Status

	// Unreserve is called if an error occurs during the Reserve stage or if another ReservePlugin rejects the pod.
	// It may be used to update 3rd party data structures.
	// This method must be idempotent and may be called by the scheduling pipeline even if Reserve() was not
	// previously called.
	Unreserve(ctx SchedulingContext, podInfo *PodInfo, targetNode *NodeInfo)
}

// Represents a scheduling decision made by the Decision Pipeline.
type SchedulingDecision struct {

	// The node that has been selected for the pod.
	SelectedNode *NodeInfo
}

// Represents an instance of the Polaris Scheduler Decision Pipeline,
// encompassing all stages from PreFilter until (including) Reserve.
//
// A Decision Pipeline executes on a single goroutine and there is only
// a single pod traversing the pipeline at a time.
type DecisionPipeline interface {

	// Executes the Decision Pipeline and returns a SchedulingDecision and a Status.
	// The SchedulingDecision is nil in case the pod could not be scheduled or if an error occurred.
	SchedulePod(podInfo *SampledPodInfo) (*SchedulingDecision, Status)
}
