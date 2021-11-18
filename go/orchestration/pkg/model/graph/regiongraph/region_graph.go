package regiongraph

import (
	"k8s.rainbow-h2020.eu/rainbow/orchestration/pkg/model/graph/labeledgraph"
)

// RegionGraph is a representation of a RAINBOW region as a weighted undirected graph.
// The weight of an edge is the number of milliseconds it takes to send a request between the
// two nodes that it connects.
type RegionGraph struct {
	labeledgraph.LabeledUndirectedGraph
	regionHead *Node
}

// NewRegionGraph creates a new instance of the default RegionGraph type.
func NewRegionGraph() *RegionGraph {
	return &RegionGraph{
		LabeledUndirectedGraph: labeledgraph.NewLabeledUndirectedGraph(NewNode),
	}
}

// Node gets the node with the specified ID.
func (me *RegionGraph) Node(id int64) *Node {
	if node := me.LabeledUndirectedGraph.Node(id); node != nil {
		return node.(*Node)
	}
	return nil
}

// NodeByLabel gets the node with the spcified label.
func (me *RegionGraph) NodeByLabel(label string) *Node {
	if node := me.LabeledUndirectedGraph.NodeByLabel(label); node != nil {
		return node.(*Node)
	}
	return nil
}

// AddNewNode creates a new node, adds it to the graph, and returns it.
func (me *RegionGraph) AddNewNode(label string, info *KubernetesNodeInfo) *Node {
	node := me.LabeledUndirectedGraph.NewNode(label).(*Node)
	node.SetKubernetesNodeInfo(info)
	me.LabeledUndirectedGraph.AddNode(node)
	return node
}

// RegionHead returns the cluster head node for the region, or nil if none has been defined.
func (me *RegionGraph) RegionHead() *Node {
	return me.regionHead
}

// SetRegionHead sets the cluster head node for the region, or panics if the
// specified node is not part of the graph.
func (me *RegionGraph) SetRegionHead(node *Node) {
	if existingNode := me.Node(node.ID()); existingNode != node {
		panic("The specified node is not part of this graph.")
	}
	me.regionHead = node
}
