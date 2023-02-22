package randomsampling

import (
	"fmt"

	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/client"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/config"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/pipeline"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/util"
)

const (
	randomPoolSize = 100
)

var (
	_ pipeline.SamplingStrategyPlugin        = (*RandomSamplingStrategy)(nil)
	_ pipeline.ClusterAgentPluginFactoryFunc = NewRandomSamplingStrategy
)

const (
	PluginName                 = "RandomSamplingStrategy"
	RandomSamplingStrategyName = "random"
)

type RandomSamplingStrategy struct {
	clusterAgentServices pipeline.ClusterAgentServices

	// A pool of util.Random objects, each of them to be used by a single goroutine,
	// because util.Random is not guaranteed to be thread-safe.
	randPool chan util.Random
}

func NewRandomSamplingStrategy(pluginConfig config.PluginConfig, clusterAgentServices pipeline.ClusterAgentServices) (pipeline.Plugin, error) {
	rs := &RandomSamplingStrategy{
		clusterAgentServices: clusterAgentServices,
		randPool:             make(chan util.Random, randomPoolSize),
	}

	for i := 0; i < randomPoolSize; i++ {
		rs.randPool <- util.NewDefaultRandom()
	}

	return rs, nil
}

func (rs *RandomSamplingStrategy) Name() string {
	return PluginName
}

func (rs *RandomSamplingStrategy) StrategyName() string {
	return RandomSamplingStrategyName
}

func (rs *RandomSamplingStrategy) SampleNodes(ctx pipeline.SchedulingContext, podInfo *pipeline.PodInfo, sampleSize int) ([]*pipeline.NodeInfo, pipeline.Status) {
	random := <-rs.randPool
	nodes := rs.sampleNodesInternal(podInfo, sampleSize, random)
	rs.randPool <- random

	clusterName := rs.clusterAgentServices.ClusterClient().ClusterName()
	nodeInfos := make([]*pipeline.NodeInfo, len(nodes))
	for i, node := range nodes {
		nodeInfos[i] = pipeline.NewNodeInfo(clusterName, node)
	}

	return nodeInfos, pipeline.NewSuccessStatus()
}

func (rs *RandomSamplingStrategy) sampleNodesInternal(podInfo *pipeline.PodInfo, reqNodesCount int, random util.Random) []*client.ClusterNode {
	storeReader := rs.clusterAgentServices.NodesCache().Nodes().ReadLock()
	defer storeReader.Unlock()

	totalNodesCount := storeReader.Len()
	if totalNodesCount == 0 {
		return make([]*client.ClusterNode, 0)
	}

	sampledNodes := make([]*client.ClusterNode, reqNodesCount)
	chosenIndices := make(map[int]bool, reqNodesCount)

	for i := 0; i < reqNodesCount; i++ {
		var randIndex int
		for {
			randIndex = random.Int(totalNodesCount)
			if _, exists := chosenIndices[randIndex]; !exists {
				break
			}
		}
		chosenIndices[randIndex] = true

		if _, node, ok := storeReader.GetByIndex(randIndex); ok {
			sampledNodes[i] = node
		} else {
			panic(fmt.Errorf("index %v not found in NodesCache", randIndex))
		}
	}

	return sampledNodes
}
