package pipeline

import "context"

// Generic interface for the state data that can be stored in a SchedulingContext.
type StateData interface{}

// SchedulingContext is used to carry state information between stages of the scheduling pipeline.
// All plugins can access the information in the SchedulingContext - they are all assumed to be trusted.
type SchedulingContext interface {

	// Gets the context.Context that this SchedulingContext is associated with.
	//
	// This may change between plugin executions, so it should always be read directly
	// from the SchedulingContext.
	//
	// This method is thread-safe.
	Context() context.Context

	// Reads state data from the SchedulingContext.
	// Returns the StateData stored under the given key and a boolean indicating if the key was found.
	//
	// This method is thread-safe.
	Read(key string) (StateData, bool)

	// Writes state data to the SchedulingContext and stores it under the given key.
	//
	// This method is thread-safe.
	Write(key string, data StateData)
}
