package servicegraph

import (
	fogappsCRDs "k8s.rainbow-h2020.eu/rainbow/orchestration/apis/fogapps/v1"
	lg "k8s.rainbow-h2020.eu/rainbow/orchestration/pkg/model/graph/labeledgraph"
)

var (
	_serviceGraphImpl *serviceGraphImpl

	_ ServiceGraph = _serviceGraphImpl
)

type serviceGraphImpl struct {
	crdInstance fogappsCRDs.ServiceGraph
	graph       lg.LabeledDirectedGraph
	userNodes   []Node
}

func serviceGraphImplFromCRDInstance(crdInstance *fogappsCRDs.ServiceGraph) *serviceGraphImpl {
	svcGraph := serviceGraphImpl{
		crdInstance: *crdInstance.DeepCopy(),
		graph:       lg.NewLabeledDirectedGraph(NewNode, NewEdge),
		userNodes:   make([]Node, 0),
	}
	svcGraph.buildGraph()
	return &svcGraph
}

func (me *serviceGraphImpl) CRDInstance() *fogappsCRDs.ServiceGraph {
	return &me.crdInstance
}

func (me *serviceGraphImpl) Graph() lg.LabeledDirectedGraph {
	return me.graph
}

func (me *serviceGraphImpl) NodeByLabel(label string) Node {
	if node := me.graph.NodeByLabel(label); node != nil {
		return node.(Node)
	}
	return nil
}

func (me *serviceGraphImpl) Edge(fromLabel, toLabel string) Edge {
	return me.graph.EdgeByLabels(fromLabel, toLabel).(Edge)
}

func (me *serviceGraphImpl) UserNodes() []Node {
	return me.userNodes
}

// Constructs the graph data structure from the CRD instance.
func (me *serviceGraphImpl) buildGraph() {
	for i := range me.crdInstance.Spec.Nodes {
		crdNode := &me.crdInstance.Spec.Nodes[i]
		graphNode := me.graph.NewNode(crdNode.Name)
		graphNode.SetPayload(crdNode)
		me.graph.AddNode(graphNode)

		if crdNode.NodeType == fogappsCRDs.UserNode {
			me.userNodes = append(me.userNodes, graphNode)
		}
	}

	for i := range me.crdInstance.Spec.Links {
		crdLink := &me.crdInstance.Spec.Links[i]
		source := me.graph.NodeByLabel(crdLink.Source)
		target := me.graph.NodeByLabel(crdLink.Target)
		graphEdge := me.graph.NewWeightedEdge(source, target, newServiceLinkWeightImpl(crdLink))
		me.graph.SetWeightedEdge(graphEdge)
	}
}
