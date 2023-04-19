package runtime

import (
	"math"
	"math/rand"
	"sort"

	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/collections"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/pipeline"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/util"
)

var (
	_ pipeline.DecisionPipeline = (*DefaultDecisionPipeline)(nil)
)

// Default implementation of the Polaris DecisionPipeline
type DefaultDecisionPipeline struct {
	id             int
	plugins        *pipeline.DecisionPipelinePlugins
	pipelineHelper *PipelineHelper
	scheduler      pipeline.PolarisScheduler
}

// Creates a new instance of the DefaultDecisionPipeline.
func NewDefaultDecisionPipeline(id int, plugins *pipeline.DecisionPipelinePlugins, scheduler pipeline.PolarisScheduler) *DefaultDecisionPipeline {
	// ToDo: ReservePlugin_MultiBind - Remove this check once we have modified ReservePlugin for the MultiBind mechanism.
	if len(plugins.Reserve) > 0 {
		panic("ToDo: Modify ReservePlugin to account for the new MultiBinding mechanism. Currently the ReserveStage is disabled. Search for 'ReservePlugin_MultiBind' in the code.")
	}

	decisionPipeline := DefaultDecisionPipeline{
		id:             id,
		plugins:        plugins,
		pipelineHelper: NewPipelineHelper(),
		scheduler:      scheduler,
	}
	return &decisionPipeline
}

func (dp *DefaultDecisionPipeline) DecideCommitCandidates(podInfo *pipeline.SampledPodInfo, commitCandidatesCount int) ([]*pipeline.SchedulingDecision, pipeline.Status) {
	schedCtx := podInfo.Ctx
	nodeEligibilityStats := &util.NodeEligibilityStats{
		SampledNodesCount: len(podInfo.SampledNodes),
	}

	status := dp.pipelineHelper.RunPreFilterPlugins(schedCtx, dp.plugins.PreFilter, podInfo.PodInfo)
	if !pipeline.IsSuccessStatus(status) {
		return nil, status
	}

	candidateNodesList := collections.ConvertToLinkedList(podInfo.SampledNodes)
	podInfo.SampledNodes = nil // Allow reclaiming memory.
	status = dp.pipelineHelper.RunFilterPlugins(schedCtx, dp.plugins.Filter, podInfo.PodInfo, candidateNodesList)
	if !pipeline.IsSuccessStatus(status) {
		return nil, status
	}
	eligibleNodes := collections.ConvertToSlice[*pipeline.NodeInfo](candidateNodesList)
	candidateNodesList = nil
	nodeEligibilityStats.EligibleNodesCount = len(eligibleNodes)
	schedCtx.Write(util.NodeEligibilityStatsInfoStateKey, nodeEligibilityStats)

	status = dp.pipelineHelper.RunPreScorePlugins(schedCtx, dp.plugins.PreScore, podInfo.PodInfo, eligibleNodes)
	if !pipeline.IsSuccessStatus(status) {
		return nil, status
	}

	allScores, status := dp.pipelineHelper.RunScorePlugins(schedCtx, dp.plugins.Score, podInfo.PodInfo, eligibleNodes)
	if !pipeline.IsSuccessStatus(status) {
		return nil, status
	}
	finalScores := dp.combineScores(dp.plugins.Score, allScores, eligibleNodes)

	commitCandidateNodes := dp.pickBestNodes(finalScores, commitCandidatesCount)
	// ToDo: ReservePlugin_MultiBind
	// status = dp.runReservePlugins(schedCtx, podInfo.PodInfo, commitCandidateNodes)
	// if !pipeline.IsSuccessStatus(status) {
	// 	dp.runUnreservePlugins(schedCtx, podInfo.PodInfo, commitCandidateNodes)
	// 	return nil, status
	// }

	commitCandidateDecisions := dp.createSchedulingDecisions(podInfo.PodInfo, commitCandidateNodes)

	return commitCandidateDecisions, pipeline.NewSuccessStatus()
}

// Aggregates the scores for each node and computes an average.
// The aggregation considers the accumulated score computed by the sampling score plugins (stored in each eligibleNode) and
// the scores computed by the scheduler's score plugins (stored in allSchedulerScores - allSchedulerScores[i] contains a list of scores for all eligible nodes computed by the scheduler score plugin i).
func (dp *DefaultDecisionPipeline) combineScores(scorePlugins []*pipeline.ScorePluginWithExtensions, allSchedulerScores [][]pipeline.NodeScore, eligibleNodes []*pipeline.NodeInfo) []pipeline.NodeScore {
	nodeScores := make([]pipeline.NodeScore, len(eligibleNodes))
	for i := range nodeScores {
		node := eligibleNodes[i]

		var initialScore int64 = 0
		if node.SamplingScore != nil {
			initialScore = node.SamplingScore.AccumulatedScore
		}

		nodeScores[i] = pipeline.NodeScore{
			Node:  node,
			Score: initialScore,
		}
	}

	// Aggregate the scores from the scheduler's score plugins.
	for pluginIndex, plugin := range scorePlugins {
		weight := int64(plugin.Weight)
		pluginScores := allSchedulerScores[pluginIndex]
		for nodeIndex := range pluginScores {
			nodeScores[nodeIndex].Score += pluginScores[nodeIndex].Score * weight
		}
	}

	schedulerScorePluginsCount := int64(len(scorePlugins))
	for i := range nodeScores {
		accumulatedScore := &nodeScores[i]
		accumulatedScoreComponents := schedulerScorePluginsCount
		if accumulatedScore.Node.SamplingScore != nil {
			accumulatedScoreComponents += int64(accumulatedScore.Node.SamplingScore.ScorePluginsCount)
		}
		accumulatedScore.Score = accumulatedScore.Score / accumulatedScoreComponents
	}

	return nodeScores
}

// Picks the 'numNodesToPick' top nodes from the list.
func (dp *DefaultDecisionPipeline) pickBestNodes(finalScores []pipeline.NodeScore, numNodesToPick int) []*pipeline.NodeInfo {
	numNodesToPick = int(math.Min(float64(numNodesToPick), float64(len(finalScores))))
	pickedNodes := make([]*pipeline.NodeInfo, numNodesToPick)

	sort.Slice(
		finalScores,
		func(i int, j int) bool {
			return finalScores[i].Score < finalScores[j].Score
		},
	)

	topScore := finalScores[0].Score
	topScoreCount := 0
	for i := 0; i < numNodesToPick; i++ {
		currNodeScore := finalScores[i]
		pickedNodes[i] = currNodeScore.Node
		if currNodeScore.Score == topScore {
			topScoreCount++
		}
	}

	// If more than one node have the same top score, shuffle these nodes in the slice.
	if topScoreCount > 1 {
		rand.Shuffle(topScoreCount, func(i int, j int) {
			temp := pickedNodes[i]
			pickedNodes[i] = pickedNodes[j]
			pickedNodes[j] = temp
		})
	}

	return pickedNodes
}

func (dp *DefaultDecisionPipeline) runReservePlugins(ctx pipeline.SchedulingContext, podInfo *pipeline.PodInfo, targetNode *pipeline.NodeInfo) pipeline.Status {
	for _, plugin := range dp.plugins.Reserve {
		status := plugin.Reserve(ctx, podInfo, targetNode)
		if !pipeline.IsSuccessStatus(status) {
			status.SetFailedPlugin(plugin, pipeline.ReserveStage)
			return status
		}
	}
	return pipeline.NewSuccessStatus()
}

func (dp *DefaultDecisionPipeline) runUnreservePlugins(ctx pipeline.SchedulingContext, podInfo *pipeline.PodInfo, targetNode *pipeline.NodeInfo) {
	for _, plugin := range dp.plugins.Reserve {
		plugin.Unreserve(ctx, podInfo, targetNode)
	}
}

func (dp *DefaultDecisionPipeline) createSchedulingDecisions(podInfo *pipeline.PodInfo, commitCandidateNodes []*pipeline.NodeInfo) []*pipeline.SchedulingDecision {
	commitCandidateDecisions := make([]*pipeline.SchedulingDecision, len(commitCandidateNodes))
	for i, commitCandidateNode := range commitCandidateNodes {
		commitCandidateDecisions[i] = &pipeline.SchedulingDecision{
			Pod:        podInfo,
			TargetNode: commitCandidateNode,
		}
	}
	return commitCandidateDecisions
}
