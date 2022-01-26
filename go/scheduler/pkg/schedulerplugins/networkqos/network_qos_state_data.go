package networkqos

import (
	"sync"

	graphpath "gonum.org/v1/gonum/graph/path"
	"k8s.io/kubernetes/pkg/scheduler/framework"

	fogappsCRDs "k8s.rainbow-h2020.eu/rainbow/orchestration/apis/fogapps/v1"
	"k8s.rainbow-h2020.eu/rainbow/orchestration/pkg/model/graph/regiongraph"
	"k8s.rainbow-h2020.eu/rainbow/orchestration/pkg/model/graph/servicegraph"
	"k8s.rainbow-h2020.eu/rainbow/orchestration/pkg/services/servicegraphmanager"
)

const (
	networkQosStateKey = "NetworkQosPlugin.networkQosStateData"
)

var (
	_ framework.StateData = (*networkQosStateData)(nil)
)

// Stores information about a K8s node from the RegionGraph that hosts an incomingServiceLink's source pod.
type k8sSourceNode struct {
	// The K8s node in the RegionGraph.
	k8sNode regiongraph.Node

	// The shortest paths starting from the k8sNode.
	shortestNetworkPaths *graphpath.Shortest
}

// Stores info about a ServiceGraph link that enters the current pod's ServiceGraph node.
type incomingServiceLink struct {
	// The ServiceGraphLink that comes from the srcNode to the ServiceGraphNode of the current pod.
	link servicegraph.Edge

	// The QoS requirements of the link.
	qosRequirements *fogappsCRDs.LinkQosRequirements

	// The K8s nodes, to which the pods, corresponding to the source Service Node of `link`, have been deployed.
	k8sSrcNodes []k8sSourceNode
}

// Used to cache information about the incoming links to a pod's ServiceGraphNode.
type networkQosStateData struct {
	svcGraphState servicegraphmanager.ServiceGraphState

	// The region graph at the time of PreFilter() - all Filter() invocations must use the same version of this graph.
	regionGraph regiongraph.RegionGraph

	// The ServiceGraph node that corresponds to the pod to be scheduled.
	podSvcNode servicegraph.Node

	// The incoming ServiceGraph links to the podSvcNode.
	incomingLinks []*incomingServiceLink

	// The minimum network QoS requirements that the region graph node must fulfill, based on the
	// ServiceGraph node's outgoing links.
	minNetworkRequirements *fogappsCRDs.LinkQosRequirements

	// Stores the score (int64) for each K8s node that passes the Filter phase.
	// The node name is used as the key.
	k8sNodeScores sync.Map
}

// Stores Information about a path between two nodes in the RegionGraph.
type networkPathInfo struct {
	// The lowest bandwidth of any link along the path.
	lowestBandwithKbps int64

	// The highest bandwidth variance of any link along the path.
	highestBandwidthVariance int64

	// The sum of the packet delays over the entire path.
	totalPacketDelayMsec int64

	// The highest packet delay variance of any link along the path.
	highestPacketDelayVariance int32

	// The highest packet loss in basis points of any link along the path.
	highestPacketLossBp int32

	// The lowest QualityClass (in Kbps) of any network link in the path.
	lowestNetworkQualityClassKbps int64
}

func (me *networkQosStateData) Clone() framework.StateData {
	return &networkQosStateData{
		svcGraphState: me.svcGraphState,
		regionGraph:   me.regionGraph,
		podSvcNode:    me.podSvcNode,
		incomingLinks: me.incomingLinks,
		k8sNodeScores: me.k8sNodeScores,
	}
}

// Gets the networkQosStateData from the CycleState or returns a framework.Success state if the current pod is not associated with a ServiceGraph
// and, thus, does not have any networkQosStateData.
func getNetworkQosStateDataOrStatus(cycleState *framework.CycleState) (*networkQosStateData, *framework.Status) {
	cycleState.RLock()
	stateData, err := cycleState.Read(networkQosStateKey)
	cycleState.RUnlock()
	if err == nil {
		return stateData.(*networkQosStateData), nil
	}
	return nil, framework.NewStatus(framework.Success, "Skipping this pod, because it is not associated with a ServiceGraph.")
}
