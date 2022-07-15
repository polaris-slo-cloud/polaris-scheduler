package pipeline

// Plugin is the parent interface for all Polaris scheduling pipeline plugins
type Plugin interface {
	Name() string
}

// A SortPlugin is used to establish the order, in which incoming pods will be handled by the scheduling pipeline.
type SortPlugin interface {
	Plugin

	// Less returns true if podA should be scheduled before podB.
	// Otherwise, it returns false.
	Less(podA *QueuedPodInfo, schedCtxA SchedulingContext, podB *QueuedPodInfo, schedCtxB SchedulingContext) bool
}
