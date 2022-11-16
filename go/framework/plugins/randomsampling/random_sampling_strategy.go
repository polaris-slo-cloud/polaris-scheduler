package randomsampling

import (
	"fmt"
	"math/rand"

	core "k8s.io/api/core/v1"

	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/config"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/pipeline"
)

const (
	randomPoolSize = 100
)

var (
	_ pipeline.SamplingStrategyPlugin    = (*RandomSamplingStrategy)(nil)
	_ pipeline.SamplingPluginFactoryFunc = NewRandomSamplingStrategy
)

const (
	PluginName                 = "RandomSamplingStrategy"
	RandomSamplingStrategyName = "random"
)

type RandomSamplingStrategy struct {
	polarisNodeSampler pipeline.PolarisNodeSampler

	// A pool of rand.Rand objects, each of them to be used by a single goroutine.
	// rand.Rand is not thread-safe and the global rand.Int() function uses a mutex to sync access to a single Rand.
	randPool chan *rand.Rand
}

func NewRandomSamplingStrategy(pluginConfig config.PluginConfig, polarisNodeSampler pipeline.PolarisNodeSampler) (pipeline.Plugin, error) {
	rs := &RandomSamplingStrategy{
		polarisNodeSampler: polarisNodeSampler,
		randPool:           make(chan *rand.Rand, randomPoolSize),
	}

	for i := 0; i < randomPoolSize; i++ {
		seed := rand.Int63()
		rs.randPool <- rand.New(rand.NewSource(seed))
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

	clusterName := rs.polarisNodeSampler.ClusterClient().ClusterName()
	nodeInfos := make([]*pipeline.NodeInfo, len(nodes))
	for i, node := range nodes {
		nodeInfos[i] = pipeline.NewNodeInfo(clusterName, node)
	}

	return nodeInfos, pipeline.NewSuccessStatus()
}

func (rs *RandomSamplingStrategy) sampleNodesInternal(podInfo *pipeline.PodInfo, reqNodesCount int, random *rand.Rand) []*core.Node {
	storeReader := rs.polarisNodeSampler.NodesCache().Nodes().ReadLock()
	defer storeReader.Unlock()

	totalNodesCount := storeReader.Len()
	if totalNodesCount == 0 {
		return make([]*core.Node, 0)
	}

	sampledNodes := make([]*core.Node, reqNodesCount)
	chosenIndices := make(map[int]bool, reqNodesCount)

	for i := 0; i < reqNodesCount; i++ {
		var randIndex int
		for {
			randIndex = random.Intn(totalNodesCount)
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
