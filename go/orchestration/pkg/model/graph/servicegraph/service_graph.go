package servicegraph

import (
	fogappsCRDs "k8s.rainbow-h2020.eu/rainbow/orchestration/apis/fogapps/v1"
	lg "k8s.rainbow-h2020.eu/rainbow/orchestration/pkg/model/graph/labeledgraph"
)

// Node represents a node in a ServiceGraph.
type Node interface {
	lg.LabeledNode
}

// Edge represents an edge in a ServiceGraph.
// Its weight is determined by a ServiceLink object.
type Edge interface {
	lg.WeightedEdge

	// Gets the fogappsCRDs.ServiceLink that describes this edge.
	ServiceLink() *fogappsCRDs.ServiceLink
}

// ServiceLinkWeight wraps a fogappsCRDs.ServiceLink in a ComplexEdgeWeight object.
type ServiceLinkWeight interface {
	lg.ComplexEdgeWeight

	// Gets the fogappsCRDs.ServiceLink stored by this weight.
	ServiceLink() *fogappsCRDs.ServiceLink
}

// ServiceGraph is a representation of a RAINBOW application as a weighted directed graph.
type ServiceGraph interface {
	// Gets the CRD instance object, from which this ServiceGraph was constructed.
	CRDInstance() *fogappsCRDs.ServiceGraph

	// Graph returns the LabeledDirectedGraph that models this application.
	Graph() lg.LabeledDirectedGraph

	// NodeByLabel gets the node with the specified label.
	NodeByLabel(label string) Node

	// Edge returns the servicegraph.Edge from the node with
	// the label fromNodeLabel to the node with the label toNodeLabel.
	// If such an edge does not exist, the method returns nil.
	Edge(fromLabel, toLabel string) Edge

	// Gets the array of nodes of the graph that have NodeType = User
	UserNodes() []Node
}

// FromCRDInstance creates a new ServiceGraph from a ServiceGraph CRD.
func FromCRDInstance(crdInstance *fogappsCRDs.ServiceGraph) ServiceGraph {
	return serviceGraphImplFromCRDInstance(crdInstance)
}
