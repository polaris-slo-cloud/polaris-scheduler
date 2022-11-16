package client

import (
	"context"

	core "k8s.io/api/core/v1"

	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/collections"
)

// Tracks and caches the full list of nodes in a cluster.
//
// The node objects in the cache are always the most recent ones received through the watch.
type NodesCache interface {
	// Starts watching the nodes.
	// This method returns once the initial list has been retrieved and added to the store.
	//
	// The passed context can be used to stop the watch.
	StartWatch(ctx context.Context) error

	// Gets the cache of all nodes.
	Nodes() collections.ConcurrentObjectStore[*core.Node]
}
