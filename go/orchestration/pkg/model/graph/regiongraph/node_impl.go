package regiongraph

import (
	lg "k8s.rainbow-h2020.eu/rainbow/orchestration/pkg/model/graph/labeledgraph"
)

var (
	_nodeImpl *nodeImpl

	_ Node = _nodeImpl
)

// The default implementation of a regiongraph.Node.
type nodeImpl struct {
	lg.LabeledNode
}

// NewNode is the factory function for creating a new regiongraph.Node
var NewNode lg.LabeledNodeFactoryFn = func(id int64, label string) lg.LabeledNode {
	return &nodeImpl{
		LabeledNode: lg.NewDefaultLabeledNode(id, label),
	}
}
