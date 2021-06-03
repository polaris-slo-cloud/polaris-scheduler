package regionmanager

import (
	"fmt"

	"gonum.org/v1/gonum/graph"
	"k8s.rainbow-h2020.eu/rainbow/orchestration/pkg/model/graph/regiongraph"
)

const (
	cloudNodesPerType = 10
)

var (
	_regionManagerImpl *regionManagerImpl

	_ RegionManager = _regionManagerImpl
)

type regionManagerImpl struct {
	regionGraph *regiongraph.RegionGraph
}

func newRegionManagerImpl() *regionManagerImpl {
	return &regionManagerImpl{}
}

func (me *regionManagerImpl) RegionGraph() *regiongraph.RegionGraph {
	if me.regionGraph == nil {
		me.regionGraph = me.buildRegionGraph()
	}
	return me.regionGraph
}

func (me *regionManagerImpl) buildRegionGraph() *regiongraph.RegionGraph {
	region := regiongraph.NewRegionGraph()

	// We create a hardcoded mocked graph for now.

	regionHead := region.AddNewNode("kind-control-plane", &regiongraph.KubernetesNodeInfo{
		Roles: []string{"control-plane", "fog", "fog-region-head", "master"},
	})
	fogWorkerNodes := []*regiongraph.Node{
		region.AddNewNode("kind-worker", &regiongraph.KubernetesNodeInfo{
			Roles: []string{"fog", "worker"},
		}),
		region.AddNewNode("kind-worker2", &regiongraph.KubernetesNodeInfo{
			Roles: []string{"fog", "worker"},
		}),
		region.AddNewNode("kind-worker3", &regiongraph.KubernetesNodeInfo{
			Roles: []string{"fog", "worker"},
		}),
		region.AddNewNode("kind-worker4", &regiongraph.KubernetesNodeInfo{
			Roles: []string{"fog", "worker"},
		}),
		region.AddNewNode("kind-worker5", &regiongraph.KubernetesNodeInfo{
			Roles: []string{"fog", "worker"},
		}),
		region.AddNewNode("kind-worker6", &regiongraph.KubernetesNodeInfo{
			Roles: []string{"fog", "worker"},
		}),
	}

	firstCloudNodeID := len(fogWorkerNodes) + 1
	var cloudWorkerNodes []*regiongraph.Node = make([]*regiongraph.Node, 0)
	cloudWorkerNodes = me.addCloudNodes(cloudWorkerNodes, region, firstCloudNodeID, cloudNodesPerType, "small")
	cloudWorkerNodes = me.addCloudNodes(cloudWorkerNodes, region, firstCloudNodeID+len(cloudWorkerNodes), cloudNodesPerType, "medium")
	cloudWorkerNodes = me.addCloudNodes(cloudWorkerNodes, region, firstCloudNodeID+len(cloudWorkerNodes), cloudNodesPerType, "large")

	region.SetRegionHead(regionHead)

	edges := []graph.WeightedEdge{
		region.NewWeightedEdge(regionHead, fogWorkerNodes[0], 15),
		region.NewWeightedEdge(fogWorkerNodes[0], fogWorkerNodes[1], 80),
		region.NewWeightedEdge(regionHead, fogWorkerNodes[2], 20),
		region.NewWeightedEdge(regionHead, fogWorkerNodes[3], 10),
		region.NewWeightedEdge(fogWorkerNodes[3], fogWorkerNodes[4], 20),
		region.NewWeightedEdge(fogWorkerNodes[2], fogWorkerNodes[3], 10),
		region.NewWeightedEdge(fogWorkerNodes[4], fogWorkerNodes[5], 75),
	}

	for i, cloudNodeI := range cloudWorkerNodes {
		// Every cloud node has a 40 ms edge to the regioHead
		edges = append(edges, region.NewWeightedEdge(regionHead, cloudNodeI, 40))

		// Every cloud node has a 0 ms edge to every other cloud node
		for j := i + 1; j < len(cloudWorkerNodes); j++ {
			edges = append(edges, region.NewWeightedEdge(cloudNodeI, cloudWorkerNodes[j], 0))
		}
	}

	for _, edge := range edges {
		region.SetWeightedEdge(edge)
	}

	return region
}

func (me *regionManagerImpl) addCloudNodes(cloudWorkerNodes []*regiongraph.Node, region *regiongraph.RegionGraph, startIndex, count int, nodeSize string) []*regiongraph.Node {
	upperBound := startIndex + count

	for i := startIndex; i < upperBound; i++ {
		cloudWorkerNodes = append(cloudWorkerNodes,
			region.AddNewNode(fmt.Sprintf("kind-worker%d", i), &regiongraph.KubernetesNodeInfo{
				Roles: []string{"cloud", "worker", nodeSize},
			}),
		)
	}
	return cloudWorkerNodes
}
