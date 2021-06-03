package regiongraph

import (
	"k8s.rainbow-h2020.eu/rainbow/orchestration/pkg/model/graph/labeledgraph"
)

// KubernetesNodeInfo provides information about a Kubernetes node in the cluster.
type KubernetesNodeInfo struct {
	Roles []string
}

// Node represents a node in a RegionGraph.
type Node struct {
	labeledgraph.LabeledNode
}

// NewNode is the factory function for creating a new regiongraph.Node
var NewNode labeledgraph.LabeledNodeFactoryFn = func(id int64, label string) labeledgraph.LabeledNode {
	return &Node{
		LabeledNode: labeledgraph.NewDefaultLabeledNode(id, label),
	}
}

// KubernetesNodeInfo gets the information about the Kubernetes node.
func (me *Node) KubernetesNodeInfo() *KubernetesNodeInfo {
	return me.Payload().(*KubernetesNodeInfo)
}

// SetKubernetesNodeInfo sets the information about the Kubernetes node.
func (me *Node) SetKubernetesNodeInfo(info *KubernetesNodeInfo) {
	me.SetPayload(info)
}
