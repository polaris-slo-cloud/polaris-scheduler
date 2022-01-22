package networkqos

import (
	"context"
	"fmt"
	"math"
	"sync"

	"gonum.org/v1/gonum/graph"
	graphpath "gonum.org/v1/gonum/graph/path"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	framework "k8s.io/kubernetes/pkg/scheduler/framework"

	clusterCRDs "k8s.rainbow-h2020.eu/rainbow/orchestration/apis/cluster/v1"
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
	_ framework.ScorePlugin     = _networkQosPlugin
	_ framework.ScoreExtensions = _networkQosPlugin
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
	incomingLinks, err := me.getIncomingSvcLinks(svcGraphState, podSvcNode, region)
	if err != nil {
		return framework.AsStatus(err)
	}
	minNetworkQosReqs := me.getMinNetworkRequirements(svcGraphState, podSvcNode, incomingLinks)

	if len(incomingLinks) > 0 {
		qosState := networkQosStateData{
			svcGraphState:          svcGraphState,
			regionGraph:            region,
			podSvcNode:             podSvcNode,
			incomingLinks:          incomingLinks,
			minNetworkRequirements: minNetworkQosReqs,
			k8sNodeScores:          sync.Map{},
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
// If the node does not meet the requirements, Filter() returns an unschedulable status.
//
// Filter() performs the following operations:
// 1. Check if the candidate K8s node's network links support the minNetworkRequirements.
// 2. FOR EACH incoming service link:
//    2.1. Compute the shortest paths (latency-wise) from all SRC nodes (see PreFilter) to the candidate K8s node.
//    2.2. Pick shortest path that meets the network QoS requirements of the Service Link. If there is none, the candidate node is not suitable.
//    2.3. If the candidate node is suitable, store the path’s highest bandwidth and latency variance values in the networkQosStateData.
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

	// Check if any of the candidate node's network connections meets the minimum requirements from all incoming and outgoing service links.
	if !me.checkNodeMeetsMinRequirements(region, candidateK8sNode, qosState.minNetworkRequirements) {
		return framework.NewStatus(
			framework.Unschedulable,
			fmt.Sprintf("None of node %s's network links meets the minimum NetworkQoS requirements.", candidateK8sNode.Label()),
		)
	}

	// Loop through the incoming service links and if the candidate node fails to meet the QoS requirements
	// even for a single link, return Unschedulable.
	shortestPaths := make([]*networkPathInfo, len(qosState.incomingLinks))
	for i, incomingServiceLink := range qosState.incomingLinks {
		shortestCompliantPath := me.findShortestCompliantPath(incomingServiceLink, candidateK8sNode, region)
		if shortestCompliantPath == nil {
			return framework.NewStatus(
				framework.Unschedulable,
				fmt.Sprintf("Node %s does not meet the NetworkQoS requirements for ServiceLink from %s.", candidateK8sNode.Label(), incomingServiceLink.link.ServiceLink().Source),
			)
		}
		shortestPaths[i] = shortestCompliantPath
	}

	// Compute the node's score and store it for the Score phase.
	// We do the computation now to not require us to store the networkPathInfos for all nodes.
	nodeScore := me.computeNodeScore(shortestPaths)
	qosState.k8sNodeScores.Store(candidateK8sNodeInfo.Node().Name, nodeScore)

	return framework.NewStatus(framework.Success)
}

// ScoreExtensions returns a ScoreExtensions interface if the plugin implements one, or nil if does not.
func (me *NetworkQosPlugin) ScoreExtensions() framework.ScoreExtensions {
	return me
}

// Score is called on each filtered node. It must return success and an integer
// indicating the rank of the node. All scoring plugins must return success or
// the pod will be rejected.
func (me *NetworkQosPlugin) Score(ctx context.Context, cycleState *framework.CycleState, pod *core.Pod, nodeName string) (int64, *framework.Status) {
	qosState, noSvcGraphStatus := getNetworkQosStateDataOrStatus(cycleState)
	if noSvcGraphStatus != nil {
		return 100, noSvcGraphStatus
	}

	score, ok := qosState.k8sNodeScores.Load(nodeName)
	if !ok {
		return 0, framework.AsStatus(fmt.Errorf("the node %s has no score in qosState.k8sNodeScores", nodeName))
	}
	intScore, ok := score.(int64)
	if !ok {
		return 0, framework.AsStatus(fmt.Errorf("could not convert qosState.k8sNodeScores[%s] into an int64", nodeName))
	}
	return intScore, framework.NewStatus(framework.Success)
}

// NormalizeScore normalizes all scores to a range between 0 and 100.
func (me *NetworkQosPlugin) NormalizeScore(ctx context.Context, cycleState *framework.CycleState, pod *core.Pod, scores framework.NodeScoreList) *framework.Status {
	_, noSvcGraphStatus := getNetworkQosStateDataOrStatus(cycleState)
	if noSvcGraphStatus != nil {
		return noSvcGraphStatus
	}

	util.NormalizeNodeScores(scores)

	// Purge qosState from CycleState to allow freeing memory.
	cycleState.Lock()
	cycleState.Delete(networkQosStateKey)
	cycleState.Unlock()

	return framework.NewStatus(framework.Success)
}

// Finds the incoming service links to the ServiceGraph node that corresponds to the pod to be scheduled.
// All incoming service links are returned, regardless of whether they have QosRequirements set or not.
func (me *NetworkQosPlugin) getIncomingSvcLinks(
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

	incomingLinkIterator := svcGraph.Graph().To(destSvcNodeId)
	for incomingLinkIterator.Next() {
		srcSvcNode := incomingLinkIterator.Node().(servicegraph.Node)
		srcSvcNodeId := srcSvcNode.ID()

		incomingLink := svcGraph.Graph().Edge(srcSvcNodeId, destSvcNodeId).(servicegraph.Edge)
		nodeAndLinkPair := incomingServiceLink{
			link:            incomingLink,
			qosRequirements: incomingLink.ServiceLink().QosRequirements,
			k8sSrcNodes:     me.buildK8sSourceNodeInfosForLink(incomingLink, placementMap, region),
		}
		incomingLinks = append(incomingLinks, &nodeAndLinkPair)
	}

	return incomingLinks, nil
}

// Computes the minimum network QoS requirements, based on all outgoing and incoming ServiceLinks to the ServiceGraph node.
func (me *NetworkQosPlugin) getMinNetworkRequirements(
	svcGraphState servicegraphmanager.ServiceGraphState,
	podSvcNode servicegraph.Node,
	incomingLinks []*incomingServiceLink,
) *fogappsCRDs.LinkQosRequirements {
	svcGraph := svcGraphState.ServiceGraph()
	srcNodeId := podSvcNode.ID()
	minReqs := newNetworkQosRequirements()

	adaptMinReqs := func(linkQos *fogappsCRDs.LinkQosRequirements) {
		if linkQos == nil {
			return
		}
		if linkQos.Throughput != nil {
			if linkQos.Throughput != nil {
				util.SetIfGreaterThanInt64(&minReqs.Throughput.MinBandwidthKbps, &linkQos.Throughput.MinBandwidthKbps)
				util.SetIfLessThanInt64(minReqs.Throughput.MaxBandwidthVariance, linkQos.Throughput.MaxBandwidthVariance)
			}
			if linkQos.Latency != nil {
				util.SetIfLessThanInt32(&minReqs.Latency.MaxPacketDelayMsec, &linkQos.Latency.MaxPacketDelayMsec)
				util.SetIfLessThanInt32(minReqs.Latency.MaxPacketDelayVariance, linkQos.Latency.MaxPacketDelayVariance)
			}
			if linkQos.PacketLoss != nil {
				util.SetIfLessThanInt32(&minReqs.PacketLoss.MaxPacketLossBp, &linkQos.PacketLoss.MaxPacketLossBp)
			}
		}
	}

	outgoingLinksIterator := svcGraph.Graph().From(podSvcNode.ID())
	for outgoingLinksIterator.Next() {
		destNodeId := outgoingLinksIterator.Node().ID()
		outgoingLink := svcGraph.Graph().Edge(srcNodeId, destNodeId).(servicegraph.Edge)
		adaptMinReqs(outgoingLink.ServiceLink().QosRequirements)
	}

	for _, incomingLink := range incomingLinks {
		adaptMinReqs(incomingLink.qosRequirements)
	}

	return minReqs
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
		return nil, fmt.Errorf("the node %s was not found in the region graph", nodeName)
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
		lowestBandwithKbps:            math.MaxInt64,
		lowestNetworkQualityClassKbps: math.MaxInt64,
	}

	pathLength := len(path)
	for i := 0; i < pathLength-1; i++ {
		startNode := path[i]
		endNode := path[i+1]
		link := region.Graph().Edge(startNode.ID(), endNode.ID()).(regiongraph.Edge)
		linkQos := link.NetworkLinkQoS()

		// Quality class
		if linkQos.QualityClass != "" {
			linkQualityClassKbps := clusterCRDs.NetworkQualitClassToKbps(linkQos.QualityClass)
			if linkQualityClassKbps < pathInfo.lowestNetworkQualityClassKbps {
				pathInfo.lowestNetworkQualityClassKbps = linkQualityClassKbps
			}
		}

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
	// If there are no requirements, the path definitely fulfills them.
	if requirements == nil {
		return true
	}

	ok := true

	// QualityClass
	if req := requirements.LinkType; req != nil && req.MinQualityClass != nil {
		requestedQualityClassKbps := clusterCRDs.NetworkQualitClassToKbps(*req.MinQualityClass)
		ok = ok && pathInfo.lowestNetworkQualityClassKbps >= requestedQualityClassKbps
	}

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

func (me *NetworkQosPlugin) checkNodeMeetsMinRequirements(
	region regiongraph.RegionGraph,
	candidateK8sNode regiongraph.Node,
	requirements *fogappsCRDs.LinkQosRequirements,
) bool {
	// If there are no requirements, the path definitely fulfills them.
	if requirements == nil {
		return true
	}

	ok := true

	candidateNodeId := candidateK8sNode.ID()
	networkLinksIterator := region.Graph().From(candidateNodeId)
	for networkLinksIterator.Next() && ok {
		link := region.Graph().Edge(candidateNodeId, networkLinksIterator.Node().ID()).(regiongraph.Edge)
		linkQos := link.NetworkLinkQoS()
		if linkQos == nil {
			continue
		}

		// QualityClass
		if req := requirements.LinkType; req != nil && req.MinQualityClass != nil {
			requestedQualityClassKbps := clusterCRDs.NetworkQualitClassToKbps(*req.MinQualityClass)
			linkQualitClassKbps := clusterCRDs.NetworkQualitClassToKbps(linkQos.QualityClass)
			ok = ok && linkQualitClassKbps >= requestedQualityClassKbps
		}

		// Throughput
		if req := requirements.Throughput; req != nil {
			ok = ok && linkQos.Throughput.BandwidthKbps >= req.MinBandwidthKbps
			if req.MaxBandwidthVariance != nil {
				ok = ok && linkQos.Throughput.BandwidthVariance <= *req.MaxBandwidthVariance
			}
		}

		// Latency
		if req := requirements.Latency; req != nil {
			ok = ok && linkQos.Latency.PacketDelayMsec <= req.MaxPacketDelayMsec
			if req.MaxPacketDelayVariance != nil {
				ok = ok && linkQos.Latency.PacketDelayVariance <= *req.MaxPacketDelayVariance
			}
		}

		// Packet loss
		if req := requirements.PacketLoss; req != nil {
			ok = ok && linkQos.PacketLoss.PacketLossBp <= req.MaxPacketLossBp
		}

	}

	return ok
}

// Computes a node score, based on the highest bandwidth and latency variance values of the shortestPaths.
func (me *NetworkQosPlugin) computeNodeScore(shortestPaths []*networkPathInfo) int64 {
	var bandwidthVarianceSum int64 = 0
	var packetDelayVarianceSum int64 = 0
	for _, path := range shortestPaths {
		bandwidthVarianceSum += path.highestBandwidthVariance
		packetDelayVarianceSum += int64(path.highestPacketDelayVariance)
	}

	pathsCount := len(shortestPaths)
	var avgBandwidthVariance float64
	var avgPacketDelayVariance float64
	if pathsCount > 0 {
		avgBandwidthVariance = float64(bandwidthVarianceSum) / float64(pathsCount)
		avgPacketDelayVariance = float64(packetDelayVarianceSum) / float64(pathsCount)
	} else {
		// There were no incoming service links, so we set the variances to 0 to get the best score.
		avgBandwidthVariance = 0
		avgPacketDelayVariance = 0
	}

	var highestScore float64 = 1000000
	bandwidthScore := highestScore / (avgBandwidthVariance + 1)
	packetDelayScore := highestScore / (avgPacketDelayVariance + 1)
	overallScore := math.Round((bandwidthScore + packetDelayScore) / 2)
	return int64(overallScore)
}

// Creates a new LinkQosRequirements object filled with the most lenient values.
func newNetworkQosRequirements() *fogappsCRDs.LinkQosRequirements {
	var maxBandwidthVar int64 = math.MaxInt64
	var maxPacketDelayVar int32 = math.MaxInt32
	minReqs := fogappsCRDs.LinkQosRequirements{
		Throughput: &fogappsCRDs.NetworkThroughputRequirements{
			MinBandwidthKbps:     0,
			MaxBandwidthVariance: &maxBandwidthVar,
		},
		Latency: &fogappsCRDs.NetworkLatencyRequirements{
			MaxPacketDelayMsec:     math.MaxInt32,
			MaxPacketDelayVariance: &maxPacketDelayVar,
		},
		PacketLoss: &fogappsCRDs.NetworkPacketLossRequirements{
			MaxPacketLossBp: 10000,
		},
	}
	return &minReqs
}
