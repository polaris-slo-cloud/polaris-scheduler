package roundrobinsampling

import (
	"fmt"
	"sync"

	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/client"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/collections"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/config"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/pipeline"
)

var (
	_ pipeline.SamplingStrategyPlugin        = (*RoundRobinSamplingStrategy)(nil)
	_ pipeline.ClusterAgentPluginFactoryFunc = NewRoundRobinSamplingStrategy
)

const (
	PluginName                     = "RoundRobinSamplingStrategy"
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
	clusterAgentServices pipeline.ClusterAgentServices

	// Stores the last index from which a node was taken for a sample.
	// Access to this is controller by the mutex.
	lastNodeIndex int
	mutex         *sync.Mutex
}

func NewRoundRobinSamplingStrategy(pluginConfig config.PluginConfig, clusterAgentServices pipeline.ClusterAgentServices) (pipeline.Plugin, error) {
	rr := &RoundRobinSamplingStrategy{
		clusterAgentServices: clusterAgentServices,
		lastNodeIndex:        -1,
		mutex:                &sync.Mutex{},
	}
	return rr, nil
}

func (rr *RoundRobinSamplingStrategy) Name() string {
	return PluginName
}

func (rr *RoundRobinSamplingStrategy) StrategyName() string {
	return RoundRobinSamplingStrategyName
}

func (rr *RoundRobinSamplingStrategy) SampleNodes(ctx pipeline.SchedulingContext, podInfo *pipeline.PodInfo, sampleSize int) ([]*pipeline.NodeInfo, pipeline.Status) {
	storeReader := rr.clusterAgentServices.NodesCache().Nodes().ReadLock()
	defer storeReader.Unlock()

	sampleRange := rr.computeSampleRange(sampleSize, storeReader)
	nodes := rr.getNodesSample(sampleRange, storeReader)

	return nodes, pipeline.NewSuccessStatus()
}

func (rr *RoundRobinSamplingStrategy) computeSampleRange(
	sampleSize int,
	storeReader collections.ConcurrentObjectStoreReader[*client.ClusterNode],
) roundRobinSampleRange {
	totalNodesCount := storeReader.Len()

	// Ensure that we do not need to sample more nodes that there are in the cache (which might have changed in the meantime).
	if sampleSize > totalNodesCount {
		sampleSize = totalNodesCount
	}

	ret := roundRobinSampleRange{
		requiredNodesCount: sampleSize,
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
	storeReader collections.ConcurrentObjectStoreReader[*client.ClusterNode],
) []*pipeline.NodeInfo {
	clusterName := rr.clusterAgentServices.ClusterClient().ClusterName()
	sampledNodes := make([]*pipeline.NodeInfo, sampleRange.requiredNodesCount)
	sampleIndex := 0

	var firstLoopEnd int
	if sampleRange.endIndex < sampleRange.startIndex {
		// We need to wrap around to the beginning of the list at some point.
		firstLoopEnd = storeReader.Len() - 1
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
