package pipeline

import (
	"context"
	"fmt"
	"sync"
)

var (
	_ SchedulingContext = (*schedulingContextImpl)(nil)
)

type schedulingContextImpl struct {
	mutex sync.RWMutex
	ctx   context.Context
	state map[string]StateData
}

// Creates a new SchedulingContext
func NewSchedulingContext(ctx context.Context) SchedulingContext {
	schedulingCtx := schedulingContextImpl{
		mutex: sync.RWMutex{},
		ctx:   ctx,
		state: make(map[string]StateData),
	}
	return &schedulingCtx
}

// Convenience function to read StateData from a SchedulingContext and casting it to a specific type.
// Returns
//   - (stateData as T, true, nil) if the key was found
//   - (zeroValue(T), false, nil) if the key was not found
//   - (zeroValue(T), true, err) if the key was found, but the value could not be converted to T
func ReadTypedStateData[T StateData](schedCtx SchedulingContext, key string) (T, bool, error) {
	data, ok := schedCtx.Read(key)
	if !ok {
		var nilT T
		return nilT, false, nil
	}

	dataT, ok := data.(T)
	if !ok {
		var nilT T
		return nilT, true, fmt.Errorf("invalid object stored as %s", key)
	}

	return dataT, true, nil
}

func (schedCtx *schedulingContextImpl) Context() context.Context {
	schedCtx.mutex.RLock()
	defer schedCtx.mutex.RUnlock()
	return schedCtx.ctx
}

func (schedCtx *schedulingContextImpl) Read(key string) (StateData, bool) {
	schedCtx.mutex.RLock()
	defer schedCtx.mutex.RUnlock()

	if data, ok := schedCtx.state[key]; ok {
		return data, true
	}
	return nil, false
}

func (schedCtx *schedulingContextImpl) Write(key string, data StateData) {
	schedCtx.mutex.Lock()
	defer schedCtx.mutex.Unlock()

	schedCtx.state[key] = data
}

// Sets the context.Context of the SchedulingContext.
func (schedCtx *schedulingContextImpl) setContext(ctx context.Context) {
	schedCtx.mutex.Lock()
	defer schedCtx.mutex.Unlock()
	schedCtx.ctx = ctx
}
