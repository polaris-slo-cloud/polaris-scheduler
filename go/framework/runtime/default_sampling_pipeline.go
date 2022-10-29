package runtime

import (
	"container/list"
	"fmt"
	"math"

	"github.com/go-logr/logr"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/collections"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/pipeline"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/util"
)

const (
	// The maximum number of iterations of the sampling loop that first obtains a set of nodes and then filters them.
	// The loop ends if enough nodes are eligible or after the maximum number of iterations.
	maxSamplingLoopIterations = 4
)

var (
	_ pipeline.SamplingPipeline = (*DefaultSamplingPipeline)(nil)
)

// Default implementation of the Polaris SamplingPipeline
type DefaultSamplingPipeline struct {
	id             int
	plugins        *pipeline.SamplingPipelinePlugins
	pipelineHelper *PipelineHelper
	nodeSampler    pipeline.PolarisNodeSampler
	logger         *logr.Logger
}

func NewDefaultSamplingPipeline(id int, plugins *pipeline.SamplingPipelinePlugins, nodeSampler pipeline.PolarisNodeSampler) *DefaultSamplingPipeline {
	samplingPipeline := &DefaultSamplingPipeline{
		id:             id,
		plugins:        plugins,
		pipelineHelper: NewPipelineHelper(),
		nodeSampler:    nodeSampler,
		logger:         nodeSampler.Logger(),
	}
	return samplingPipeline
}

// SampleNodes implements pipeline.SamplingPipeline
func (sp *DefaultSamplingPipeline) SampleNodes(
	ctx pipeline.SchedulingContext,
	samplingStrategy pipeline.SamplingStrategyPlugin,
	podInfo *pipeline.PodInfo,
	nodesToSampleBp int,
) ([]*pipeline.NodeInfo, pipeline.Status) {
	var status pipeline.Status
	var eligibleNodes []*pipeline.NodeInfo

	stopwatch := util.NewStopwatch()
	stopwatch.Start()
	defer sp.stopAndLogStopwatch(stopwatch, podInfo, status, eligibleNodes)

	sampleSize := sp.calcRequiredNodesCount(nodesToSampleBp)

	status = sp.pipelineHelper.RunPreFilterPlugins(ctx, sp.plugins.PreFilter, podInfo)
	if !pipeline.IsSuccessStatus(status) {
		return nil, status
	}

	eligibleNodes, status = sp.sampleAndFilterNodes(ctx, samplingStrategy, podInfo, sampleSize)
	if !pipeline.IsSuccessStatus(status) {
		return nil, status
	}

	status = sp.pipelineHelper.RunPreScorePlugins(ctx, sp.plugins.PreScore, podInfo, eligibleNodes)
	if !pipeline.IsSuccessStatus(status) {
		return nil, status
	}

	allScores, status := sp.pipelineHelper.RunScorePlugins(ctx, sp.plugins.Score, podInfo, eligibleNodes)
	if !pipeline.IsSuccessStatus(status) {
		return nil, status
	}
	sp.combineAndAssignScores(sp.plugins.Score, allScores, eligibleNodes)

	return eligibleNodes, pipeline.NewSuccessStatus()
}

func (sp *DefaultSamplingPipeline) calcRequiredNodesCount(nodesToSampleBp int) int {
	storeReader := sp.nodeSampler.NodesCache().Nodes().ReadLock()
	defer storeReader.Unlock()

	percentageOfNodesToSample := float64(nodesToSampleBp) / 10000.0
	reqNodes := percentageOfNodesToSample * float64(storeReader.Len())
	return int(math.Max(reqNodes, 1))
}

func (sp *DefaultSamplingPipeline) sampleAndFilterNodes(
	ctx pipeline.SchedulingContext,
	samplingStrategy pipeline.SamplingStrategyPlugin,
	podInfo *pipeline.PodInfo,
	sampleSize int,
) ([]*pipeline.NodeInfo, pipeline.Status) {
	eligibleNodes := list.New()

	for i := 0; eligibleNodes.Len() < sampleSize && i <= maxSamplingLoopIterations; i++ {
		missingNodesCount := sampleSize - eligibleNodes.Len()

		currSample, status := samplingStrategy.SampleNodes(ctx, podInfo, missingNodesCount)
		if !pipeline.IsSuccessStatus(status) {
			return nil, status
		}

		currSampleList := collections.ConvertToLinkedList(currSample)
		currSample = nil // Allow reclaiming memory
		status = sp.pipelineHelper.RunFilterPlugins(ctx, sp.plugins.Filter, podInfo, currSampleList)
		if !pipeline.IsSuccessStatus(status) && status != nil && status.Code() != pipeline.Unschedulable {
			return nil, status
		}
		collections.AppendToLinkedList(eligibleNodes, currSampleList)
	}

	if eligibleNodes.Len() == 0 {
		return nil, pipeline.NewStatus(pipeline.Unschedulable, "no candidates left after Filter plugins")
	}
	return collections.ConvertToSlice[*pipeline.NodeInfo](eligibleNodes), nil
}

// Aggregates all scores for each node and assigns them to the SamplingScore field of the node.
func (sp *DefaultSamplingPipeline) combineAndAssignScores(scorePlugins []*pipeline.ScorePluginWithExtensions, allScores [][]pipeline.NodeScore, eligibleNodes []*pipeline.NodeInfo) {
	scorePluginsCount := len(scorePlugins)
	if scorePluginsCount == 0 {
		return
	}

	for _, node := range eligibleNodes {
		node.SamplingScore = &pipeline.SamplingScore{
			ScorePluginsCount: scorePluginsCount,
		}
	}

	// Aggregate the scores from the score plugins.
	for pluginIndex, plugin := range scorePlugins {
		weight := int64(plugin.Weight)
		pluginScores := allScores[pluginIndex]
		for nodeIndex := range pluginScores {
			eligibleNodes[nodeIndex].SamplingScore.AccumulatedScore += pluginScores[nodeIndex].Score * weight
		}
	}
}

func (sp *DefaultSamplingPipeline) stopAndLogStopwatch(stopwatch *util.Stopwatch, podInfo *pipeline.PodInfo, status pipeline.Status, eligibleNodes []*pipeline.NodeInfo) {
	stopwatch.Stop()
	fullPodName := fmt.Sprintf("%s.%s", podInfo.Pod.Namespace, podInfo.Pod.Name)
	sp.logger.Info(
		"Pod traversed sampling pipeline",
		"pod", fullPodName,
		"success", pipeline.IsSuccessStatus(status),
		"eligibleNodes", len(eligibleNodes),
	)
}
