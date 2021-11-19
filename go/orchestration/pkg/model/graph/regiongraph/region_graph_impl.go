package regiongraph

import (
	"gonum.org/v1/gonum/graph"
	cluster "k8s.rainbow-h2020.eu/rainbow/orchestration/apis/cluster/v1"
	lg "k8s.rainbow-h2020.eu/rainbow/orchestration/pkg/model/graph/labeledgraph"
)

var (
	_regionGraphImpl *regionGraphImpl

	_ RegionGraph = _regionGraphImpl
)

// regionGraphImpl is the default implementation of regionGraphImpl
type regionGraphImpl struct {
	graph lg.LabeledUndirectedGraph
}

// NewRegionGraph creates a new instance of the default RegionGraph type.
func NewRegionGraph() RegionGraph {
	return &regionGraphImpl{
		graph: lg.NewLabeledUndirectedGraph(NewNode, NewEdge),
	}
}

func (me *regionGraphImpl) Graph() lg.LabeledUndirectedGraph {
	return me.graph
}

func (me *regionGraphImpl) NodeByLabel(label string) Node {
	if node := me.graph.NodeByLabel(label); node != nil {
		return node.(Node)
	}
	return nil
}

func (me *regionGraphImpl) Edge(fromLabel, toLabel string) Edge {
	var fromNode, toNode Node
	if fromNode = me.graph.NodeByLabel(fromLabel); fromNode == nil {
		return nil
	}
	if toNode = me.graph.NodeByLabel(toLabel); toNode == nil {
		return nil
	}

	var edge graph.WeightedEdge
	if edge = me.graph.WeightedEdge(fromNode.ID(), toNode.ID()); edge == nil {
		return nil
	}
	return edge.(Edge)
}

func (me *regionGraphImpl) NewNode(label string) Node {
	return me.graph.NewNode(label).(Node)
}

func (me *regionGraphImpl) AddNode(node Node) {
	me.graph.AddNode(node)
}

func (me *regionGraphImpl) NewEdge(from, to Node, qos *cluster.NetworkLinkQoS) Edge {
	weight := newNetworkLinkQosWeightImpl(qos)
	return me.graph.NewWeightedEdge(from, to, weight).(Edge)
}

func (me *regionGraphImpl) SetEdge(edge Edge) {
	me.graph.SetWeightedEdge(edge)
}
