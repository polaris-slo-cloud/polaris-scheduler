package regionmanager

import (
	cluster "k8s.rainbow-h2020.eu/rainbow/orchestration/apis/cluster/v1"
	"k8s.rainbow-h2020.eu/rainbow/orchestration/pkg/model/graph/regiongraph"
)

var (
	_regionManagerImpl *regionManagerImpl

	_ RegionManager = _regionManagerImpl
)

type regionManagerImpl struct {
	regionGraph regiongraph.RegionGraph
}

func newRegionManagerImpl() *regionManagerImpl {
	return &regionManagerImpl{}
}

func (me *regionManagerImpl) RegionGraph() regiongraph.RegionGraph {
	if me.regionGraph == nil {
		// ToDo
		me.regionGraph = me.buildRegionGraph(nil)
	}
	return me.regionGraph
}

func (me *regionManagerImpl) buildRegionGraph(networkLinks *cluster.NetworkLinkList) regiongraph.RegionGraph {
	region := regiongraph.NewRegionGraph()

	for i := range networkLinks.Items {
		networkLink := &networkLinks.Items[i]
		fromNode := getOrCreateNode(region, networkLink.Spec.NodeA)
		toNode := getOrCreateNode(region, networkLink.Spec.NodeB)

		edge := region.NewEdge(fromNode, toNode, &networkLink.Spec.QoS)
		region.SetEdge(edge)
	}

	return region
}

func getOrCreateNode(region regiongraph.RegionGraph, nodeName string) regiongraph.Node {
	node := region.NodeByLabel(nodeName)
	if node == nil {
		node = region.NewNode(nodeName)
		region.AddNode(node)
	}
	return node
}
