package servicegraph

import (
	"k8s.rainbow-h2020.eu/rainbow/orchestration/pkg/model/graph/labeledgraph"
	"k8s.rainbow-h2020.eu/rainbow/orchestration/pkg/model/graph/regiongraph"
)

// MicroserviceNodeInfo provides information about a microservice in a RAINBOW application.
type MicroserviceNodeInfo struct {
	MicroserviceType           string
	MaxLatencyToMessageQueueMs int64
	ScheduledOnNode            *regiongraph.Node
}

// MicroserviceNode represents a node in a ServiceGraph.
type MicroserviceNode struct {
	labeledgraph.LabeledNode
}

// NewMicroserviceNode is the factory function for creating a new servicegraph.Node
var NewMicroserviceNode labeledgraph.LabeledNodeFactoryFn = func(id int64, label string) labeledgraph.LabeledNode {
	return &MicroserviceNode{
		LabeledNode: labeledgraph.NewDefaultLabeledNode(id, label),
	}
}

// MicroserviceNodeInfo gets the information about the Kubernetes node.
func (me *MicroserviceNode) MicroserviceNodeInfo() *MicroserviceNodeInfo {
	return me.Payload().(*MicroserviceNodeInfo)
}

// SetMicroserviceNodeInfo sets the information about the Kubernetes node.
func (me *MicroserviceNode) SetMicroserviceNodeInfo(info *MicroserviceNodeInfo) {
	me.SetPayload(info)
}
