package runtime

import (
	"math/rand"
	"sort"
	"time"

	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/collections"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/pipeline"
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
	random         *rand.Rand
}

// Creates a new instance of the DefaultDecisionPipeline.
func NewDefaultDecisionPipeline(id int, plugins *pipeline.DecisionPipelinePlugins, scheduler pipeline.PolarisScheduler) *DefaultDecisionPipeline {
	decisionPipeline := DefaultDecisionPipeline{
		id:             id,
		plugins:        plugins,
		pipelineHelper: NewPipelineHelper(),
		scheduler:      scheduler,
		random:         rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	return &decisionPipeline
}

func (dp *DefaultDecisionPipeline) SchedulePod(podInfo *pipeline.SampledPodInfo) (*pipeline.SchedulingDecision, pipeline.Status) {
	schedCtx := podInfo.Ctx

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

	status = dp.pipelineHelper.RunPreScorePlugins(schedCtx, dp.plugins.PreScore, podInfo.PodInfo, eligibleNodes)
	if !pipeline.IsSuccessStatus(status) {
		return nil, status
	}

	allScores, status := dp.pipelineHelper.RunScorePlugins(schedCtx, dp.plugins.Score, podInfo.PodInfo, eligibleNodes)
	if !pipeline.IsSuccessStatus(status) {
		return nil, status
	}
	finalScores := dp.combineScores(dp.plugins.Score, allScores, eligibleNodes)

	targetNode := dp.pickBestNode(finalScores)
	status = dp.runReservePlugins(schedCtx, podInfo.PodInfo, targetNode)
	if !pipeline.IsSuccessStatus(status) {
		dp.runUnreservePlugins(schedCtx, podInfo.PodInfo, targetNode)
		return nil, status
	}

	decision := pipeline.SchedulingDecision{
		Pod:        podInfo.PodInfo,
		TargetNode: targetNode,
	}
	return &decision, pipeline.NewSuccessStatus()
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

func (dp *DefaultDecisionPipeline) pickBestNode(finalScores []pipeline.NodeScore) *pipeline.NodeInfo {
	sort.Slice(
		finalScores,
		func(i int, j int) bool {
			return finalScores[i].Score < finalScores[j].Score
		},
	)

	topScore := finalScores[0].Score
	topScoreCount := 0
	for i := range finalScores {
		if finalScores[i].Score == topScore {
			topScoreCount++
		} else {
			break
		}
	}

	// If more than one node have the same top score, we pick a random one from them.
	selectedIndex := 0
	if topScoreCount > 1 {
		selectedIndex = dp.random.Intn(topScoreCount)
	}

	return finalScores[selectedIndex].Node
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
