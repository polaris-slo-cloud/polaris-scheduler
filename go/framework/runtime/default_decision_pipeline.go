package runtime

import (
	"container/list"
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
	id        int
	plugins   *pipeline.DecisionPipelinePlugins
	scheduler pipeline.PolarisScheduler
	random    *rand.Rand
}

// Creates a new instance of the DefaultDecisionPipeline.
func NewDefaultDecisionPipeline(id int, plugins *pipeline.DecisionPipelinePlugins, scheduler pipeline.PolarisScheduler) *DefaultDecisionPipeline {
	decisionPipeline := DefaultDecisionPipeline{
		id:        id,
		plugins:   plugins,
		scheduler: scheduler,
		random:    rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	return &decisionPipeline
}

func (dp *DefaultDecisionPipeline) SchedulePod(podInfo *pipeline.SampledPodInfo) (*pipeline.SchedulingDecision, pipeline.Status) {
	schedCtx := podInfo.Ctx

	status := dp.runPreFilterPlugins(schedCtx, podInfo.PodInfo)
	if !pipeline.IsSuccessStatus(status) {
		return nil, status
	}

	candidateNodesList := collections.ConvertToLinkedList(podInfo.SampledNodes)
	podInfo.SampledNodes = nil // Allow reclaiming memory.
	status = dp.runFilterPlugins(schedCtx, podInfo.PodInfo, candidateNodesList)
	if !pipeline.IsSuccessStatus(status) {
		return nil, status
	}
	eligibleNodes := collections.ConvertToSlice[*pipeline.NodeInfo](candidateNodesList)
	candidateNodesList = nil

	status = dp.runPreScorePlugins(schedCtx, podInfo.PodInfo, eligibleNodes)
	if !pipeline.IsSuccessStatus(status) {
		return nil, status
	}

	finalScores, status := dp.runScorePlugins(schedCtx, podInfo.PodInfo, eligibleNodes)
	if !pipeline.IsSuccessStatus(status) {
		return nil, status
	}

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

func (dp *DefaultDecisionPipeline) runPreFilterPlugins(ctx pipeline.SchedulingContext, podInfo *pipeline.PodInfo) pipeline.Status {
	for _, plugin := range dp.plugins.PreFilter {
		status := plugin.PreFilter(ctx, podInfo)
		if !pipeline.IsSuccessStatus(status) {
			status.SetFailedPlugin(plugin, pipeline.PreFilterStage)
			return status
		}
	}
	return pipeline.NewSuccessStatus()
}

func (dp *DefaultDecisionPipeline) runFilterPlugins(ctx pipeline.SchedulingContext, podInfo *pipeline.PodInfo, candidateNodes *list.List) pipeline.Status {
	for _, plugin := range dp.plugins.Filter {
		status := dp.runFilterPlugin(ctx, plugin, podInfo, candidateNodes)

		if !pipeline.IsSuccessStatus(status) {
			status.SetFailedPlugin(plugin, pipeline.FilterStage)
			return status
		}

		if candidateNodes.Len() == 0 {
			unschedulableStatus := pipeline.NewStatus(pipeline.Unschedulable, "no candidates left after Filter plugins")
			unschedulableStatus.SetFailedPlugin(plugin, pipeline.FilterStage)
			return unschedulableStatus
		}
	}
	return pipeline.NewSuccessStatus()
}

func (dp *DefaultDecisionPipeline) runFilterPlugin(ctx pipeline.SchedulingContext, plugin pipeline.FilterPlugin, podInfo *pipeline.PodInfo, candidateNodes *list.List) pipeline.Status {
	for currNode := candidateNodes.Front(); currNode != nil; {
		status := plugin.Filter(ctx, podInfo, currNode.Value.(*pipeline.NodeInfo))

		// We get the next element already now, because the current one might be removed from the list.
		nextNode := currNode.Next()

		if !pipeline.IsSuccessStatus(status) {
			switch status.Code() {
			case pipeline.Unschedulable:
				candidateNodes.Remove(currNode)
				break
			case pipeline.InternalError:
				return status
			}
		}

		currNode = nextNode
	}
	return pipeline.NewSuccessStatus()
}

func (dp *DefaultDecisionPipeline) runPreScorePlugins(ctx pipeline.SchedulingContext, podInfo *pipeline.PodInfo, eligibleNodes []*pipeline.NodeInfo) pipeline.Status {
	for _, plugin := range dp.plugins.PreScore {
		status := plugin.PreScore(ctx, podInfo, eligibleNodes)
		if !pipeline.IsSuccessStatus(status) {
			status.SetFailedPlugin(plugin, pipeline.PreScoreStage)
			return status
		}
	}
	return pipeline.NewSuccessStatus()
}

func (dp *DefaultDecisionPipeline) runScorePlugins(ctx pipeline.SchedulingContext, podInfo *pipeline.PodInfo, eligibleNodes []*pipeline.NodeInfo) ([]pipeline.NodeScore, pipeline.Status) {
	allScores := make([][]pipeline.NodeScore, len(eligibleNodes))

	for i, plugin := range dp.plugins.Score {
		scores, status := dp.runScorePlugin(ctx, plugin, podInfo, eligibleNodes)
		if !pipeline.IsSuccessStatus(status) {
			status.SetFailedPlugin(plugin, pipeline.ScoreStage)
			return nil, status
		}
		allScores[i] = scores
	}

	finalScores := dp.combineScores(dp.plugins.Score, allScores, eligibleNodes)
	return finalScores, pipeline.NewSuccessStatus()
}

func (dp *DefaultDecisionPipeline) runScorePlugin(
	ctx pipeline.SchedulingContext,
	plugin *pipeline.ScorePluginWithExtensions,
	podInfo *pipeline.PodInfo,
	eligibleNodes []*pipeline.NodeInfo,
) ([]pipeline.NodeScore, pipeline.Status) {
	scores := make([]pipeline.NodeScore, len(eligibleNodes))

	for i, node := range eligibleNodes {
		score, status := plugin.Score(ctx, podInfo, node)
		if !pipeline.IsSuccessStatus(status) {
			return nil, status
		}
		scores[i] = pipeline.NodeScore{
			Node:  node,
			Score: score,
		}
	}

	if plugin.ScoreExtensions != nil {
		status := plugin.ScoreExtensions.NormalizeScores(ctx, podInfo, scores)
		if !pipeline.IsSuccessStatus(status) {
			return nil, status
		}
	}

	return scores, pipeline.NewSuccessStatus()
}

func (dp *DefaultDecisionPipeline) combineScores(scorePlugins []*pipeline.ScorePluginWithExtensions, allScores [][]pipeline.NodeScore, eligibleNodes []*pipeline.NodeInfo) []pipeline.NodeScore {
	nodeScores := make([]pipeline.NodeScore, len(eligibleNodes))
	for i := range nodeScores {
		nodeScores[i] = pipeline.NodeScore{
			Node:  eligibleNodes[i],
			Score: 0,
		}
	}

	for pluginIndex, plugin := range scorePlugins {
		weight := int64(plugin.Weight)
		pluginScores := allScores[pluginIndex]
		for nodeIndex := range pluginScores {
			nodeScores[nodeIndex].Score += pluginScores[nodeIndex].Score * weight
		}
	}

	pluginsCount := int64(len(scorePlugins))
	for i := range nodeScores {
		accumulatedScore := &nodeScores[i]
		accumulatedScore.Score = accumulatedScore.Score / pluginsCount
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
