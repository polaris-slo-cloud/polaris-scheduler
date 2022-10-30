package runtime

import (
	"container/list"

	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/pipeline"
)

// Handles functionality that is common to both the scheduling and the sampling pipeline.
type PipelineHelper struct {
}

func NewPipelineHelper() *PipelineHelper {
	ph := &PipelineHelper{}
	return ph
}

// Runs the specified PreFilter plugins for the specified pod.
func (ph *PipelineHelper) RunPreFilterPlugins(ctx pipeline.SchedulingContext, preFilterPlugins []pipeline.PreFilterPlugin, podInfo *pipeline.PodInfo) pipeline.Status {
	for _, plugin := range preFilterPlugins {
		status := plugin.PreFilter(ctx, podInfo)
		if !pipeline.IsSuccessStatus(status) {
			status.SetFailedPlugin(plugin, pipeline.PreFilterStage)
			return status
		}
	}
	return pipeline.NewSuccessStatus()
}

// Runs the specified Filter plugins on the candidate nodes list and removes all nodes that are not eligible from that list.
//
// Returns a Success status, if candidate nodes are left after filtering, or an Unschedulable status if no candidate nodes are left after filtering.
func (ph *PipelineHelper) RunFilterPlugins(
	ctx pipeline.SchedulingContext,
	filterPlugins []pipeline.FilterPlugin,
	podInfo *pipeline.PodInfo,
	candidateNodes *list.List,
) pipeline.Status {
	for _, plugin := range filterPlugins {
		status := ph.runFilterPlugin(ctx, plugin, podInfo, candidateNodes)

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

func (ph *PipelineHelper) runFilterPlugin(ctx pipeline.SchedulingContext, plugin pipeline.FilterPlugin, podInfo *pipeline.PodInfo, candidateNodes *list.List) pipeline.Status {
	for currNode := candidateNodes.Front(); currNode != nil; {
		status := plugin.Filter(ctx, podInfo, currNode.Value.(*pipeline.NodeInfo))

		// We get the next element already now, because the current one might be removed from the list.
		nextNode := currNode.Next()

		if !pipeline.IsSuccessStatus(status) {
			switch status.Code() {
			case pipeline.Unschedulable:
				candidateNodes.Remove(currNode)
			case pipeline.InternalError:
				return status
			}
		}

		currNode = nextNode
	}
	return pipeline.NewSuccessStatus()
}

// Runs the specified PreScore plugins on the eligible nodes.
func (ph *PipelineHelper) RunPreScorePlugins(
	ctx pipeline.SchedulingContext,
	preScorePlugins []pipeline.PreScorePlugin,
	podInfo *pipeline.PodInfo,
	eligibleNodes []*pipeline.NodeInfo,
) pipeline.Status {
	for _, plugin := range preScorePlugins {
		status := plugin.PreScore(ctx, podInfo, eligibleNodes)
		if !pipeline.IsSuccessStatus(status) {
			status.SetFailedPlugin(plugin, pipeline.PreScoreStage)
			return status
		}
	}
	return pipeline.NewSuccessStatus()
}

// Runs the specified Score plugins and their extensions on the eligible nodes.
//
// Returns a nested array of allScores (type [][]pipeline.NodeScore), where allScores[i] stores the scores for all eligible nodes computed by score plugin i.
func (ph *PipelineHelper) RunScorePlugins(
	ctx pipeline.SchedulingContext,
	scorePlugins []*pipeline.ScorePluginWithExtensions,
	podInfo *pipeline.PodInfo,
	eligibleNodes []*pipeline.NodeInfo,
) ([][]pipeline.NodeScore, pipeline.Status) {
	// allScores[i] stores the scores for all eligible nodes computed by score plugin i.
	allScores := make([][]pipeline.NodeScore, len(scorePlugins))

	for i, plugin := range scorePlugins {
		scores, status := ph.runScorePlugin(ctx, plugin, podInfo, eligibleNodes)
		if !pipeline.IsSuccessStatus(status) {
			status.SetFailedPlugin(plugin, pipeline.ScoreStage)
			return nil, status
		}
		allScores[i] = scores
	}

	return allScores, pipeline.NewSuccessStatus()
}

func (ph *PipelineHelper) runScorePlugin(
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
