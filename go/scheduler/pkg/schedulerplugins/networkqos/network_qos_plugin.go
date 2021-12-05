package networkqos

import (
	"context"
	"fmt"
	"math"

	"gonum.org/v1/gonum/graph"
	graphpath "gonum.org/v1/gonum/graph/path"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	framework "k8s.io/kubernetes/pkg/scheduler/framework"

	fogappsCRDs "k8s.rainbow-h2020.eu/rainbow/orchestration/apis/fogapps/v1"
	"k8s.rainbow-h2020.eu/rainbow/orchestration/pkg/model/graph/regiongraph"
	"k8s.rainbow-h2020.eu/rainbow/orchestration/pkg/model/graph/servicegraph"
	"k8s.rainbow-h2020.eu/rainbow/orchestration/pkg/serviceplacement"
	"k8s.rainbow-h2020.eu/rainbow/orchestration/pkg/services/regionmanager"
	"k8s.rainbow-h2020.eu/rainbow/orchestration/pkg/services/servicegraphmanager"
	"k8s.rainbow-h2020.eu/rainbow/scheduler/internal/util"
)

const (
	// PluginName is the name of this scheduler plugin.
	PluginName = "NetworkQoS"
)

var (
	_networkQosPlugin *NetworkQosPlugin

	_ framework.Plugin          = _networkQosPlugin
	_ framework.PreFilterPlugin = _networkQosPlugin
	_ framework.FilterPlugin    = _networkQosPlugin
)

// NetworkQosPlugin is a Filter plugin that filters out nodes that violate the network QoS constraints of the application.
type NetworkQosPlugin struct {
	regionManager regionmanager.RegionManager
}

var _ framework.FilterPlugin = &NetworkQosPlugin{}

// New creates a new NetworkQosPlugin instance.
func New(obj runtime.Object, handle framework.Handle) (framework.Plugin, error) {
	return &NetworkQosPlugin{
		regionManager: regionmanager.GetRegionManager(),
	}, nil
}

// Name returns the name of this scheduler plugin.
func (me *NetworkQosPlugin) Name() string {
	return PluginName
}

// PreScore finds incoming links in ServiceGraph and caches them in the networkQosStateData with the following information:
// - the ServiceLink itself
// - the network QoS requirements
// - SRC = { K8s nodes that have the Service Link’s source pod scheduled on them and the shortest paths from them }
func (me *NetworkQosPlugin) PreFilter(ctx context.Context, cycleState *framework.CycleState, pod *core.Pod) *framework.Status {
	svcGraphState, noSvcGraphStatus := util.GetServiceGraphFromCycleStateOrStatus(cycleState)
	if noSvcGraphStatus != nil {
		return noSvcGraphStatus
	}

	svcGraph := svcGraphState.ServiceGraph()
	podSvcNode, _ := util.GetServiceGraphNode(svcGraph, pod)
	region := me.regionManager.RegionGraph()
	incomingLinks, err := me.getIncomingSvcLinksWithQoS(svcGraphState, podSvcNode, region)
	if err != nil {
		return framework.AsStatus(err)
	}

	if len(incomingLinks) > 0 {
		qosState := networkQosStateData{
			svcGraphState: svcGraphState,
			regionGraph:   region,
			podSvcNode:    podSvcNode,
			incomingLinks: incomingLinks,
		}
		cycleState.Lock()
		cycleState.Write(networkQosStateKey, &qosState)
		cycleState.Unlock()
	}

	return framework.NewStatus(framework.Success)
}

// Returns the PreFilterExtensions, if this plugin implements them.
func (me *NetworkQosPlugin) PreFilterExtensions() framework.PreFilterExtensions {
	return nil
}

// Filter checks if the current K8s node meets the NetworkQoS requirements defined for the pod in the service graph.
// If the nodes does not meet the requirements, Filter() returns an unschedulable status.
//
// Filter() performs the following steps FOR EACH incoming service link:
// 1. Compute the shortest paths (latency-wise) from all SRC nodes (see PreFilter) to the candidate K8s node.
// 2. Pick shortest path that meets the network QoS requirements of the Service Link. If there is none, the candidate node is not suitable.
// 3. TODO: If the candidate node is suitable, store the path’s highest bandwidth and latency variance values in the networkQosStateData.
func (me *NetworkQosPlugin) Filter(ctx context.Context, cycleState *framework.CycleState, pod *core.Pod, candidateK8sNodeInfo *framework.NodeInfo) *framework.Status {
	qosState, noSvcGraphStatus := getNetworkQosStateDataOrStatus(cycleState)
	if noSvcGraphStatus != nil {
		return noSvcGraphStatus
	}

	region := qosState.regionGraph
	candidateK8sNode, err := me.getK8sNodeFromRegion(region, candidateK8sNodeInfo.Node().Name)
	if err != nil {
		return framework.AsStatus(err)
	}

	for _, incomingServiceLink := range qosState.incomingLinks {
		shortestCompliantPath := me.findShortestCompliantPath(incomingServiceLink, candidateK8sNode, region)
		if shortestCompliantPath == nil {
			return framework.NewStatus(
				framework.Unschedulable,
				fmt.Sprintf("Node %s with does not meet the NetworkQoS requirements for ServiceLink from %s.", candidateK8sNode.Label(), incomingServiceLink.link.ServiceLink().Source),
			)
		}
	}

	return framework.NewStatus(framework.Success)
}

// Finds the incoming service links to the ServiceGraph node that corresponds to the pod to be scheduled.
// Only service links that have NetworkQosRequirements set are returned.
func (me *NetworkQosPlugin) getIncomingSvcLinksWithQoS(
	svcGraphState servicegraphmanager.ServiceGraphState,
	podSvcNode servicegraph.Node,
	region regiongraph.RegionGraph,
) ([]*incomingServiceLink, error) {
	incomingLinks := make([]*incomingServiceLink, 0)
	destSvcNodeId := podSvcNode.ID()
	svcGraph := svcGraphState.ServiceGraph()
	placementMap, err := svcGraphState.PlacementMap()
	if err != nil {
		return nil, err
	}

	svcNodeIterator := svcGraph.Graph().Nodes()
	for svcNodeIterator.Next() {
		currSvcNode := svcNodeIterator.Node().(servicegraph.Node)
		currSvcNodeId := currSvcNode.ID()
		if currSvcNodeId == destSvcNodeId {
			continue
		}

		incomingLink := svcGraph.Graph().Edge(currSvcNodeId, destSvcNodeId).(servicegraph.Edge)
		if incomingLink != nil {
			if qosReqs := incomingLink.ServiceLink().QosRequirements; me.checkLinkQosRequirementsPresent(qosReqs) {
				nodeAndLinkPair := incomingServiceLink{
					link:            incomingLink,
					qosRequirements: qosReqs,
					k8sSrcNodes:     me.buildK8sSourceNodeInfosForLink(incomingLink, placementMap, region),
				}
				incomingLinks = append(incomingLinks, &nodeAndLinkPair)
			}
		}
	}

	return incomingLinks, nil
}

// Returns true if any LinkQoSRequirements are set.
func (me *NetworkQosPlugin) checkLinkQosRequirementsPresent(requirements *fogappsCRDs.LinkQosRequirements) bool {
	ok := requirements != nil
	return ok && (requirements.Throughput != nil || requirements.Latency != nil || requirements.PacketLoss != nil)
}

func (me *NetworkQosPlugin) buildK8sSourceNodeInfosForLink(
	link servicegraph.Edge,
	placementMap serviceplacement.ServiceGraphPlacementMap,
	region regiongraph.RegionGraph,
) []k8sSourceNode {
	k8sSrcNodeNames := placementMap.GetKubernetesNodes(link.From().(servicegraph.Node).Label())
	nodeInfos := make([]k8sSourceNode, len(k8sSrcNodeNames))

	for i, nodeName := range k8sSrcNodeNames {
		k8sNode := region.NodeByLabel(nodeName)
		shortestPaths := graphpath.DijkstraFrom(k8sNode, region.Graph())
		nodeInfos[i] = k8sSourceNode{
			k8sNode:              k8sNode,
			shortestNetworkPaths: &shortestPaths,
		}
	}

	return nodeInfos
}

func (me *NetworkQosPlugin) getK8sNodeFromRegion(region regiongraph.RegionGraph, nodeName string) (regiongraph.Node, error) {
	k8sNode := region.NodeByLabel(nodeName)
	if k8sNode == nil {
		return nil, fmt.Errorf("The node %s was not found in the region graph", nodeName)
	}
	return k8sNode, nil
}

// Returns the shortest path between for the incoming service link that meets the QoS requirements or nil if none can be found.
func (me *NetworkQosPlugin) findShortestCompliantPath(
	incomingSvcLink *incomingServiceLink,
	candidateK8sNode regiongraph.Node,
	region regiongraph.RegionGraph,
) *networkPathInfo {
	var shortestPath *networkPathInfo

	for _, k8sSrcNode := range incomingSvcLink.k8sSrcNodes {
		path, _ := k8sSrcNode.shortestNetworkPaths.To(candidateK8sNode.ID())
		pathInfo := me.computeNetworkPathInfo(path, region)

		if me.checkPathMeetsRequirements(&pathInfo, incomingSvcLink.qosRequirements) {
			if shortestPath == nil || pathInfo.totalPacketDelayMsec < shortestPath.totalPacketDelayMsec {
				shortestPath = &pathInfo
			}
		}
	}

	return shortestPath
}

func (me *NetworkQosPlugin) computeNetworkPathInfo(path []graph.Node, region regiongraph.RegionGraph) networkPathInfo {
	pathInfo := networkPathInfo{
		lowestBandwithKbps: math.MaxInt64,
	}

	pathLength := len(path)
	for i := 0; i < pathLength-1; i++ {
		startNode := path[i]
		endNode := path[i+1]
		link := region.Graph().Edge(startNode.ID(), endNode.ID()).(regiongraph.Edge)
		linkQos := link.NetworkLinkQoS()

		// Throughput
		if linkQos.Throughput.BandwidthKbps < pathInfo.lowestBandwithKbps {
			pathInfo.lowestBandwithKbps = linkQos.Throughput.BandwidthKbps
		}
		if linkQos.Throughput.BandwidthVariance > pathInfo.highestBandwidthVariance {
			pathInfo.highestBandwidthVariance = linkQos.Throughput.BandwidthVariance
		}

		// Latency
		pathInfo.totalPacketDelayMsec += int64(linkQos.Latency.PacketDelayMsec)
		if linkQos.Latency.PacketDelayVariance > pathInfo.highestPacketDelayVariance {
			pathInfo.highestPacketDelayVariance = linkQos.Latency.PacketDelayVariance
		}

		// Packet loss
		if linkQos.PacketLoss.PacketLossBp > pathInfo.highestPacketLossBp {
			pathInfo.highestPacketLossBp = linkQos.PacketLoss.PacketLossBp
		}
	}

	return pathInfo
}

func (me *NetworkQosPlugin) checkPathMeetsRequirements(pathInfo *networkPathInfo, requirements *fogappsCRDs.LinkQosRequirements) bool {
	ok := true

	// Throughput
	if req := requirements.Throughput; req != nil {
		ok = ok && pathInfo.lowestBandwithKbps >= req.MinBandwidthKbps
		if req.MaxBandwidthVariance != nil {
			ok = ok && pathInfo.highestBandwidthVariance <= *req.MaxBandwidthVariance
		}
	}

	// Latency
	if req := requirements.Latency; req != nil {
		ok = ok && pathInfo.totalPacketDelayMsec <= int64(req.MaxPacketDelayMsec)
		if req.MaxPacketDelayVariance != nil {
			ok = ok && pathInfo.highestPacketDelayVariance <= *req.MaxPacketDelayVariance
		}
	}

	// Packet loss
	if req := requirements.PacketLoss; req != nil {
		ok = ok && pathInfo.highestPacketLossBp <= req.MaxPacketLossBp
	}

	return ok
}
