package serviceplacement

// Used to transform a string slice into a new string slice.
type StringSliceTransformFn func(curr []string) []string

// ServiceGraphPlacementMap allows caching and looking up the Kubernetes nodes, on which
// the pods of a ServiceGraphNode have been placed.
//
// It provides the following thread safe operations:
// - lookup of Kubernetes node names for a service graph node
// - update of Kubernetes node names for a service graph node
//
// The node names for each ServiceGraphNode are accessible as an array slice (instead of a Set),
// because the common use case is iterating through the list of nodes (e.g., the NetworkQosPlugin in the scheduler).
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
	//
	// Each K8s node name should appear only once in the list.
	SetKubernetesNodes(svcGraphNodeLabel string, updateFn StringSliceTransformFn)

	// Returns true if this map was created for the initial placement of the ServiceGraph (i.e., if the placement map was initially empty).
	IsInitialPlacement() bool
}

// Creates a new ServicePlacementMap.
// isInitialPlacement indicates if this is the first time that pods for this ServiceGraph are placed
// (i.e., if the placement map will be initially empty)
func NewServicePlacementMap(isInitialPlacement bool) ServiceGraphPlacementMap {
	return newServicePlacementMapImpl(isInitialPlacement)
}
