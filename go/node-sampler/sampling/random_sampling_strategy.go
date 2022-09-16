package sampling

import (
	"fmt"
	"math/rand"

	core "k8s.io/api/core/v1"

	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/pipeline"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/remotesampling"
)

const (
	randomPoolSize = 100
)

var (
	_ SamplingStrategy            = (*RandomSamplingStrategy)(nil)
	_ SamplingStrategyFactoryFunc = NewRandomSamplingStrategy
)

const (
	RandomSamplingStrategyName = "random"
)

type RandomSamplingStrategy struct {
	polarisNodeSampler PolarisNodeSampler

	// A pool of rand.Rand objects, each of them to be used by a single goroutine.
	// rand.Rand is not thread-safe and the global rand.Int() function uses a mutex to sync access to a single Rand.
	randPool chan *rand.Rand
}

func NewRandomSamplingStrategy(polarisNodeSampler PolarisNodeSampler) (SamplingStrategy, error) {
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
	return RandomSamplingStrategyName
}

func (rs *RandomSamplingStrategy) SampleNodes(request *remotesampling.RemoteNodesSamplerRequest) (*remotesampling.RemoteNodesSamplerResponse, error) {
	random := <-rs.randPool
	nodes := rs.sampleNodesInternal(request, random)
	rs.randPool <- random

	clusterName := rs.polarisNodeSampler.ClusterClient().ClusterName()
	nodeInfos := make([]*pipeline.NodeInfo, len(nodes))
	for i, node := range nodes {
		nodeInfos[i] = pipeline.NewNodeInfo(clusterName, node)
	}

	response := &remotesampling.RemoteNodesSamplerResponse{
		Nodes: nodeInfos,
	}
	return response, nil
}

func (rs *RandomSamplingStrategy) sampleNodesInternal(request *remotesampling.RemoteNodesSamplerRequest, random *rand.Rand) []*core.Node {
	storeReader := rs.polarisNodeSampler.NodesCache().Nodes().ReadLock()
	defer storeReader.Unlock()

	reqNodesCount := calcRequiredNodesCount(request, storeReader)
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
