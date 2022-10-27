package runtime

import (
	"math"

	"k8s.io/kubernetes/pkg/apis/core"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/collections"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/pipeline"
)

var (
	_ pipeline.SamplingPipeline = (*DefaultSamplingPipeline)(nil)
)

// Default implementation of the Polaris SamplingPipeline
type DefaultSamplingPipeline struct {
	id          int
	plugins     *pipeline.SamplingPipelinePlugins
	nodeSampler pipeline.PolarisNodeSampler
}

func NewDefaultSamplingPipeline(id int, plugins *pipeline.SamplingPipelinePlugins, nodeSampler pipeline.PolarisNodeSampler) *DefaultSamplingPipeline {
	samplingPipeline := &DefaultSamplingPipeline{
		id:          id,
		plugins:     plugins,
		nodeSampler: nodeSampler,
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
	reqNodesCount := sp.calcRequiredNodesCount(nodesToSampleBp)

}

func (sp *DefaultSamplingPipeline) calcRequiredNodesCount(nodesToSampleBp int) int {
	storeReader := sp.nodeSampler.NodesCache().Nodes().ReadLock()
	defer storeReader.Unlock()
	return calcRequiredNodesCountFromStore(nodesToSampleBp, storeReader)
}

// Calculates the number of nodes that a sample for the specified request needs to contain.
func calcRequiredNodesCountFromStore(
	nodesToSampleBp int,
	storeReader collections.ConcurrentObjectStoreReader[*core.Node],
) int {
	percentageOfNodesToSample := float64(nodesToSampleBp) / 10000.0
	reqNodes := percentageOfNodesToSample * float64(storeReader.Len())
	return int(math.Max(reqNodes, 1))
}
