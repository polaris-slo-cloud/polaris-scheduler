package serviceplacement

// Used to transform a string slice into a new string slice.
type StringSliceTransformFn func(curr []string) []string

// ServiceGraphPlacementMap allows caching and looking up the Kubernetes nodes, on which
// the pods of a ServiceGraphNode have been placed.
//
// It provides the following thread safe operations:
// - lookup of Kubernetes node names for a service graph node
// - update of Kubernetes node names for a service graph node
type ServiceGraphPlacementMap interface {

	// Gets the list of Kubernetes node names, to which at least one pod of the
	// ServiceGraphNode has been assigned.
	//
	// If no pods have been assigned to any nodes, the returned list is empty.
	// If the service graph node is unknown, nil is returned.
	GetKubernetesNodes(svcGraphNodeLabel string) []string

	// Sets the list of Kubernetes node names for the ServiceGraphNode with the specified label.
	//
	// The updateFn is called after locking the respective node and receives the current list of nodes
	// (nil if the ServiceGraphNode has not yet been added). It must return a new list that should
	// be stored for the ServiceGraphNode.
	SetKubernetesNodes(svcGraphNodeLabel string, updateFn StringSliceTransformFn)
}

// Creates a new ServicePlacementMap
func NewServicePlacementMap() ServiceGraphPlacementMap {
	return newServicePlacementMapImpl()
}
