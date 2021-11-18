package servicegraph

import (
	"sync"

	"k8s.rainbow-h2020.eu/rainbow/orchestration/pkg/model/graph/labeledgraph"
)

// ServiceGraph is a representation of an application as an undirected graph.
type ServiceGraph struct {
	labeledgraph.LabeledGraph
	messageQueueNode *MicroserviceNode
	namespace        string
	appName          string
	maxDelayMs       int64 // ToDo: We can probably delete this, because it is not needed if every pod has its own delay value.

	// Mutex is used to synchronize access to this graph.
	Mutex *sync.RWMutex
}

// NewServiceGraph creates a new instance of ServiceGraph.
func NewServiceGraph(namespace, appName string) *ServiceGraph {
	return &ServiceGraph{
		LabeledGraph: labeledgraph.NewLabeledUndirectedGraph(NewMicroserviceNode),
		namespace:    namespace,
		appName:      appName,
		Mutex:        &sync.RWMutex{},
	}
}

// Node gets the node with the specified ID.
func (me *ServiceGraph) Node(id int64) *MicroserviceNode {
	if node := me.LabeledGraph.Node(id); node != nil {
		return node.(*MicroserviceNode)
	}
	return nil
}

// NodeByLabel gets the node with the spcified label.
func (me *ServiceGraph) NodeByLabel(label string) *MicroserviceNode {
	if node := me.LabeledGraph.NodeByLabel(label); node != nil {
		return node.(*MicroserviceNode)
	}
	return nil
}

// AddNewNode creates a new node, adds it to the graph, and returns it.
func (me *ServiceGraph) AddNewNode(label string, info *MicroserviceNodeInfo) *MicroserviceNode {
	node := me.LabeledGraph.NewNode(label).(*MicroserviceNode)
	node.SetMicroserviceNodeInfo(info)
	me.LabeledGraph.AddNode(node)
	return node
}

// MessageQueueNode returns the node that represents the message queue, or nil if none has been set.
func (me *ServiceGraph) MessageQueueNode() *MicroserviceNode {
	return me.messageQueueNode
}

// SetMessageQueueNode sets the node that represents the message queue.
func (me *ServiceGraph) SetMessageQueueNode(node *MicroserviceNode) {
	if existingNode := me.Node(node.ID()); existingNode != node {
		panic("The specified node is not part of this graph.")
	}
	me.messageQueueNode = node
}

// Namespace returns the namespace, where the graph should be deployed.
func (me *ServiceGraph) Namespace() string {
	return me.namespace
}

// AppName returns the name of the application that the graph represents.
func (me *ServiceGraph) AppName() string {
	return me.appName
}

// MaxDelayMs returns the maximum allowed delay between any node and the fog-region-head.
func (me *ServiceGraph) MaxDelayMs() int64 {
	return me.maxDelayMs
}

// SetMaxDelayMs sets the maximum allowed delay between any node and the fog-region-head.
func (me *ServiceGraph) SetMaxDelayMs(maxDelayMs int64) {
	me.maxDelayMs = maxDelayMs
}
