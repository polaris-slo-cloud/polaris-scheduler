package sampling

import (
	"fmt"
	"sync"

	core "k8s.io/api/core/v1"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/collections"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/pipeline"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/remotesampling"
)

var (
	_ SamplingStrategy            = (*RandomSamplingStrategy)(nil)
	_ SamplingStrategyFactoryFunc = NewRoundRobinSamplingStrategy
)

const (
	RoundRobinSamplingStrategyName = "round-robin"
)

type roundRobinSampleRange struct {

	// The index, where to start copying nodes to the sample.
	startIndex int

	// The index (inclusive) of the last node to be copied to the sample.
	endIndex int

	requiredNodesCount int
}

type RoundRobinSamplingStrategy struct {
	polarisNodeSampler PolarisNodeSampler

	// Stores the last index from which a node was taken for a sample.
	// Access to this is controller by the mutex.
	lastNodeIndex int
	mutex         *sync.Mutex
}

func NewRoundRobinSamplingStrategy(polarisNodeSampler PolarisNodeSampler) (SamplingStrategy, error) {
	rr := &RoundRobinSamplingStrategy{
		polarisNodeSampler: polarisNodeSampler,
		lastNodeIndex:      -1,
		mutex:              &sync.Mutex{},
	}
	return rr, nil
}

func (rr *RoundRobinSamplingStrategy) Name() string {
	return RoundRobinSamplingStrategyName
}

func (rr *RoundRobinSamplingStrategy) SampleNodes(request *remotesampling.RemoteNodesSamplerRequest) (*remotesampling.RemoteNodesSamplerResponse, error) {
	response := &remotesampling.RemoteNodesSamplerResponse{}

	storeReader := rr.polarisNodeSampler.NodesCache().Nodes().ReadLock()
	defer storeReader.Unlock()

	sampleRange := rr.computeSampleRange(request, storeReader)
	response.Nodes = rr.getNodesSample(sampleRange, storeReader)

	return response, nil
}

func (rr *RoundRobinSamplingStrategy) computeSampleRange(
	request *remotesampling.RemoteNodesSamplerRequest,
	storeReader collections.ConcurrentObjectStoreReader[*core.Node],
) roundRobinSampleRange {
	totalNodesCount := storeReader.Len()
	ret := roundRobinSampleRange{
		requiredNodesCount: calcRequiredNodesCount(request, storeReader),
	}
	if totalNodesCount == 0 {
		ret.requiredNodesCount = 0
		return ret
	}

	rr.mutex.Lock()
	defer rr.mutex.Unlock()

	ret.startIndex = rr.lastNodeIndex + 1
	if ret.startIndex == totalNodesCount {
		ret.startIndex = 0
	}

	ret.endIndex = ret.startIndex + ret.requiredNodesCount - 1
	if ret.endIndex >= totalNodesCount {
		// We take all the nodes until the end of the list and then wrap around to the beginning of the list.
		nodesUntilListEnd := totalNodesCount - ret.startIndex
		remainingNodes := ret.requiredNodesCount - nodesUntilListEnd
		ret.endIndex = remainingNodes - 1
	}

	rr.lastNodeIndex = ret.endIndex
	return ret
}

func (rr *RoundRobinSamplingStrategy) getNodesSample(
	sampleRange roundRobinSampleRange,
	storeReader collections.ConcurrentObjectStoreReader[*core.Node],
) []*pipeline.NodeInfo {
	clusterName := rr.polarisNodeSampler.ClusterClient().ClusterName()
	sampledNodes := make([]*pipeline.NodeInfo, sampleRange.requiredNodesCount)
	sampleIndex := 0

	var firstLoopEnd int
	if sampleRange.endIndex < sampleRange.startIndex {
		// We need to wrap around to the beginning of the list at some point.
		firstLoopEnd = storeReader.Len()
	} else {
		// We can go straight from firstIndex to lastIndex
		firstLoopEnd = sampleRange.endIndex
	}

	// Copy the first segment, from startIndex to max(endIndex, nodesCache.Len() - 1)
	for i := sampleRange.startIndex; i <= firstLoopEnd; i++ {
		_, node, ok := storeReader.GetByIndex(i)
		if !ok {
			panic(fmt.Errorf("index %v not found in NodesCache", i))
		}
		sampledNodes[sampleIndex] = pipeline.NewNodeInfo(clusterName, node)
		sampleIndex++
	}

	if sampleIndex == sampleRange.requiredNodesCount {
		return sampledNodes
	}

	// Copy the second segment (after wrapping around to the start of the cache list).
	for i := 0; i <= sampleRange.endIndex; i++ {
		_, node, ok := storeReader.GetByIndex(i)
		if !ok {
			panic(fmt.Errorf("index %v not found in NodesCache", i))
		}
		sampledNodes[sampleIndex] = pipeline.NewNodeInfo(clusterName, node)
		sampleIndex++
	}

	return sampledNodes
}
