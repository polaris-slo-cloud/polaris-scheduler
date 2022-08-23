package randomsampler

import (
	"math"
	"math/rand"
	"time"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/client"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/config"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/pipeline"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/util"
)

var (
	_ pipeline.SampleNodesPlugin = (*RandomNodesSamplerPlugin)(nil)
	_ pipeline.PluginFactoryFunc = NewRandomNodesSamplerPlugin
)

const (
	PluginName = "RandomNodesSampler"
)

type RandomNodesSamplerPlugin struct {
	clusterMgr                client.ClusterClientsManager
	scheduler                 pipeline.PolarisScheduler
	percentageOfNodesToSample float64
	random                    *rand.Rand
}

func NewRandomNodesSamplerPlugin(config config.PluginConfig, scheduler pipeline.PolarisScheduler) (pipeline.Plugin, error) {
	percentageOfNodesToSample := float64(scheduler.Config().NodesToSampleBp) / 10000.0

	plugin := RandomNodesSamplerPlugin{
		clusterMgr:                scheduler.ClusterClientsManager(),
		scheduler:                 scheduler,
		percentageOfNodesToSample: percentageOfNodesToSample,
		random:                    rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	return &plugin, nil
}

func (rsp *RandomNodesSamplerPlugin) Name() string {
	return PluginName
}

func (rsp *RandomNodesSamplerPlugin) SampleNodes(ctx pipeline.SchedulingContext, podInfo *pipeline.PodInfo) ([]*pipeline.NodeInfo, pipeline.Status) {
	// ToDo: improve this to avoid fetching all nodes every time.
	allNodes, totalNodesCount, err := rsp.getAllNodes(ctx)
	if err != nil {
		return nil, pipeline.NewInternalErrorStatus(err)
	}

	clusterSampleCounts, requiredSamples := rsp.calculateClusterNodeCounts(allNodes, totalNodesCount)
	nodeInfos := make([]*pipeline.NodeInfo, requiredSamples)

	nextSampleIndex := 0
	for i, samplesFromCluster := range clusterSampleCounts {
		rsp.sampleNodesFromCluster(allNodes[i], samplesFromCluster, nodeInfos, nextSampleIndex)
		nextSampleIndex += samplesFromCluster
	}

	return nodeInfos, pipeline.NewSuccessStatus()
}

func (rsp *RandomNodesSamplerPlugin) getAllNodes(ctx pipeline.SchedulingContext) ([][]*core.Node, int, error) {
	allNodes := make([][]*core.Node, 0, rsp.clusterMgr.ClustersCount())
	totalNodesCount := 0

	err := rsp.clusterMgr.ForEach(func(clusterName string, clusterClient client.ClusterClient) error {
		nodes, err := rsp.getClusterNodes(ctx, clusterClient)
		if err != nil {
			return err
		}
		allNodes = append(allNodes, nodes)
		totalNodesCount += len(nodes)
		return nil
	})
	if err != nil {
		return nil, -1, err
	}

	return allNodes, totalNodesCount, nil
}

func (rsp *RandomNodesSamplerPlugin) getClusterNodes(ctx pipeline.SchedulingContext, clusterClient client.ClusterClient) ([]*core.Node, error) {
	nodeList, err := clusterClient.ClientSet().CoreV1().Nodes().List(ctx.Context(), meta.ListOptions{})
	if err != nil {
		return nil, err
	}

	ret := make([]*core.Node, len(nodeList.Items))
	for i := range nodeList.Items {
		ret[i] = &nodeList.Items[i]
	}
	return ret, nil
}

func (rsp *RandomNodesSamplerPlugin) calculateClusterNodeCounts(allNodes [][]*core.Node, totalNodesCount int) ([]int, int) {
	clustersCount := len(allNodes)
	totalNodesFloat := float64(totalNodesCount)
	totalSamples := math.Ceil(totalNodesFloat * rsp.percentageOfNodesToSample)
	clusterNodeCounts := make([]int, clustersCount)

	// Due to the ceil operation for every cluster, the final number of total samples may be higher than the calculated totalSamples.
	correctedTotalSamples := 0

	for i, clusterNodes := range allNodes {
		clusterPercentage := float64(len(clusterNodes)) / totalNodesFloat
		samplesFromCluster := int(math.Ceil(clusterPercentage * totalSamples))
		clusterNodeCounts[i] = samplesFromCluster
		correctedTotalSamples += samplesFromCluster
	}

	return clusterNodeCounts, correctedTotalSamples
}

func (rsp *RandomNodesSamplerPlugin) sampleNodesFromCluster(clusterNodes []*core.Node, count int, dest []*pipeline.NodeInfo, destStartIndex int) {
	destIndex := destStartIndex
	availNodesCount := len(clusterNodes)

	for i := 0; i < count; i++ {
		index := rsp.random.Intn(availNodesCount)
		dest[destIndex] = createNodeInfo(clusterNodes[index])
		destIndex++
		availNodesCount--
		util.Swap(clusterNodes, index, availNodesCount)
	}
}

func createNodeInfo(node *core.Node) *pipeline.NodeInfo {
	return &pipeline.NodeInfo{
		Node:                 node,
		AllocatableResources: util.NewResourcesFromList(node.Status.Allocatable),
		TotalResources:       util.NewResourcesFromList(node.Status.Capacity),
	}
}
