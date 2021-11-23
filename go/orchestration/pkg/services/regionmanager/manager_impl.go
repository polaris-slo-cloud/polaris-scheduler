package regionmanager

import (
	"sync/atomic"

	"k8s.io/klog/v2"
	cluster "k8s.rainbow-h2020.eu/rainbow/orchestration/apis/cluster/v1"
	"k8s.rainbow-h2020.eu/rainbow/orchestration/pkg/kubeutil"
	"k8s.rainbow-h2020.eu/rainbow/orchestration/pkg/model/graph/regiongraph"
	"k8s.rainbow-h2020.eu/rainbow/orchestration/pkg/services/configmanager"
)

var (
	_regionManagerImpl *regionManagerImpl

	_ RegionManager = _regionManagerImpl
)

type regionManagerImpl struct {
	regionGraph atomic.Value
	watcher     kubeutil.ListWatcher
}

func newRegionManagerImpl() *regionManagerImpl {
	configMgr := configmanager.GetConfigManager()
	watcher, err := kubeutil.StartListWatcher(&cluster.NetworkLinkList{}, configMgr.RestConfig(), configMgr.Scheme())
	if err != nil {
		panic(err)
	}
	regionMgr := &regionManagerImpl{
		watcher: watcher,
	}

	// Build the initial region graph
	networkLinks := (<-watcher.WatchChan()).(*cluster.NetworkLinkList)
	regionGraph := regionMgr.buildRegionGraph(networkLinks)
	regionMgr.regionGraph.Store(regionGraph)

	go regionMgr.watchNetworkLinks()
	return regionMgr
}

func (me *regionManagerImpl) RegionGraph() regiongraph.RegionGraph {
	return me.regionGraph.Load().(regiongraph.RegionGraph)
}

func (me *regionManagerImpl) watchNetworkLinks() {
	watchChan := me.watcher.WatchChan()
	for networkLinks := range watchChan {
		updatedGraph := me.buildRegionGraph(networkLinks.(*cluster.NetworkLinkList))
		me.regionGraph.Store(updatedGraph)
	}
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

	klog.Infof("Successfully built a RegionGraph with %v links.", len(networkLinks.Items))
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
