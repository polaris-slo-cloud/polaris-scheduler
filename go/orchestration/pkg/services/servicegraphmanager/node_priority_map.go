package servicegraphmanager

import (
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/traverse"

	fogappsCRDs "k8s.rainbow-h2020.eu/rainbow/orchestration/apis/fogapps/v1"
	"k8s.rainbow-h2020.eu/rainbow/orchestration/pkg/model/graph/servicegraph"
)

var (
	_ NodePriorityMap = (*nodePriorityMapImpl)(nil)
)

// Used to store the scheduling priorities of the nodes of a ServiceGraph.
type NodePriorityMap interface {
	// Returns the priority of the node with the specified label.
	// Lower numbers indicate higher priority, 0 is the highest.
	// If the specified node cannot be found, the return value is -1.
	NodePriority(nodeLabel string) int
}

type nodePriorityMapImpl struct {
	nodePriorities map[string]int
}

// Creates a new NodePriorityMap from a ServiceGraph
func NewNodePriorityMapFromServiceGraph(svcGraph servicegraph.ServiceGraph) NodePriorityMap {
	priorityMap := nodePriorityMapImpl{
		nodePriorities: make(map[string]int, svcGraph.Graph().Nodes().Len()),
	}
	priorityMap.build(svcGraph)

	return &priorityMap
}

func (me *nodePriorityMapImpl) NodePriority(nodeLabel string) int {
	if priority, ok := me.nodePriorities[nodeLabel]; ok {
		return priority
	}
	return -1
}

// Performs a breadth-first traversal on the ServiceGraph to build the priorities map.
func (me *nodePriorityMapImpl) build(svcGraph servicegraph.ServiceGraph) {
	priority := 0

	bfTraversal := traverse.BreadthFirst{
		Visit: func(n graph.Node) {
			svcGraphNode := n.(servicegraph.Node)
			if svcGraphNode.ServiceGraphNode().NodeType != fogappsCRDs.UserNode {
				me.nodePriorities[svcGraphNode.Label()] = priority
				priority++
			}
		},
	}

	var startNode servicegraph.Node = nil
	userNodes := svcGraph.UserNodes()
	if len(userNodes) == 0 {
		startNode = userNodes[0]
	} else {
		allNodes := svcGraph.CRDInstance().Spec.Nodes
		if len(allNodes) == 0 {
			// There are no nodes in the graph
			return
		}
		startNode = svcGraph.NodeByLabel(allNodes[0].Name)
	}

	bfTraversal.Walk(svcGraph.Graph(), startNode, nil)
}
