package pipeline

import (
	"context"
	"sync"
)

var (
	schedulingCtxImpl *schedulingContextImpl = nil

	_ SchedulingContext = schedulingCtxImpl
)

type schedulingContextImpl struct {
	mutex     sync.RWMutex
	ctx       context.Context
	scheduler PolarisSchedulerService
	state     map[string]StateData
}

// Creates a new SchedulingContext
func NewSchedulingContext(ctx context.Context, scheduler PolarisSchedulerService) SchedulingContext {
	schedulingCtx := schedulingContextImpl{
		mutex:     sync.RWMutex{},
		ctx:       ctx,
		scheduler: scheduler,
		state:     make(map[string]StateData),
	}
	return &schedulingCtx
}

func (schedCtx *schedulingContextImpl) Context() context.Context {
	schedCtx.mutex.RLock()
	defer schedCtx.mutex.RUnlock()
	return schedCtx.ctx
}

func (schedCtx *schedulingContextImpl) Scheduler() PolarisSchedulerService {
	return schedCtx.scheduler
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
