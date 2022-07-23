package runtime

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/go-logr/logr"
	core "k8s.io/api/core/v1"

	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/config"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/pipeline"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/runtime/queue"
)

var (
	_ pipeline.PolarisScheduler = (*DefaultPolarisScheduler)(nil)

	schedulerStates = [...]string{"pristine", "started", "stopped"}
)

const (
	pristine int32 = 0
	started  int32 = 1
	stopped  int32 = 2
)

// The default implementation of the PolarisScheduler.
type DefaultPolarisScheduler struct {
	config     *config.SchedulerConfig
	plugins    *pipeline.PluginRegistry
	podSource  pipeline.PodSource
	schedQueue queue.SchedulingQueue

	// The number of pods currently in the pipeline.
	// This field must be read/written using atomic operations.
	podsInPipeline int32

	// The context.Context used by the scheduler.
	ctx context.Context

	// Used to stop the scheduler's goroutines.
	stopCh chan bool

	// Describes the current state of the scheduler (pristine, started, stopped).
	// This field must be read/written using atomic operations.
	state int32

	// The logger used by the scheduler.
	logger *logr.Logger
}

// Creates a new instance of the default implementation of the PolarisScheduler.
func NewDefaultPolarisScheduler(conf *config.SchedulerConfig, podSource pipeline.PodSource, logger *logr.Logger) *DefaultPolarisScheduler {
	config.SetDefaultsSchedulerConfig(conf)
	log := logger.WithName("DefaultPolarisScheduler")

	scheduler := DefaultPolarisScheduler{
		config:    conf,
		podSource: podSource,
		stopCh:    make(chan bool, 1),
		state:     pristine,
		logger:    &log,
	}
	return &scheduler
}

func (ps *DefaultPolarisScheduler) Config() *config.SchedulerConfig {
	return ps.config
}

func (ps *DefaultPolarisScheduler) Start(ctx context.Context) error {
	if !atomic.CompareAndSwapInt32(&ps.state, pristine, started) {
		state := atomic.LoadInt32(&ps.state)
		return fmt.Errorf("cannot start scheduler, because its current state is: %s", schedulerStates[state])
	}

	ps.ctx = ctx

	if err := ps.validateConfig(); err != nil {
		return err
	}

	ps.schedQueue = queue.NewPrioritySchedulingQueue(ps.plugins.Sort.Less)
	go ps.pumpIntoQueue(ps.podSource)

	for i := 0; i < int(ps.config.ParallelSchedulingPipelines); i++ {
		go ps.executePipelinePump(i)
	}

	return nil
}

func (ps *DefaultPolarisScheduler) Stop() error {
	if atomic.CompareAndSwapInt32(&ps.state, started, stopped) {
		close(ps.stopCh)
		ps.schedQueue.Close()
	}
	return nil
}

func (ps *DefaultPolarisScheduler) IsActive() bool {
	return atomic.LoadInt32(&ps.state) == started
}

func (ps *DefaultPolarisScheduler) PodsInPipelineCount() int {
	return int(atomic.LoadInt32(&ps.podsInPipeline))
}

func (ps *DefaultPolarisScheduler) PodsInQueueCount() int {
	return ps.schedQueue.Len()
}

func (ps *DefaultPolarisScheduler) validateConfig() error {
	if ps.plugins.Sort == nil {
		return fmt.Errorf("cannot start DefaultPolarisScheduler, because no SortPlugin is configured")
	}

	if len(ps.plugins.SampleNodes) == 0 {
		return fmt.Errorf("cannot start DefaultPolarisScheduler, because no SampleNodesPlugin is configured")
	}

	if len(ps.plugins.Filter) == 0 {
		return fmt.Errorf("cannot start DefaultPolarisScheduler, because no FilterPlugin is configured")
	}

	if len(ps.plugins.Score) == 0 {
		return fmt.Errorf("cannot start DefaultPolarisScheduler, because no ScorePlugin is configured")
	}

	return nil
}

// Listens to the specified PodSource and pumps its emitted pods into the scheduling queue.
func (ps *DefaultPolarisScheduler) pumpIntoQueue(src pipeline.PodSource) {
	incomingPods := src.IncomingPods()

	for {
		select {
		case pod, ok := <-incomingPods:
			if ok {
				ps.addPodToQueue(pod)
			} else {
				ps.Stop()
			}
		case <-ps.stopCh:
			// Stop signal received, so we stop the scheduler.
			return
		}
	}
}

// Creates a SchedulingContext for the pod and adds it to the scheduling queue.
func (ps *DefaultPolarisScheduler) addPodToQueue(pod *core.Pod) {
	schedCtx := pipeline.NewSchedulingContext(ps.ctx)
	queuedPod := pipeline.NewQueuedPodInfo(pod, schedCtx)
	ps.schedQueue.Enqueue(queuedPod)
}

// Continuously retrieves pods from the scheduling queue and pumps each of them through
// a single scheduling pipeline.
func (ps *DefaultPolarisScheduler) executePipelinePump(id int) {
	ps.logger.Info("starting DefaultPolarisScheduler pipeline", "id", id)

	for {
		pod := ps.schedQueue.Dequeue()
		if pod == nil {
			break
		}

		atomic.AddInt32(&ps.podsInPipeline, 1)
		ps.schedulePod(pod.PodInfo)
		atomic.AddInt32(&ps.podsInPipeline, -1)
	}

	ps.logger.Info("stopped DefaultPolarisScheduler pipeline", "id", id)
}

// Executes the scheduling pipeline for the specified pod.
func (ps *DefaultPolarisScheduler) schedulePod(podInfo *pipeline.PodInfo) error {

}
