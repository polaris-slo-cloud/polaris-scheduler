package runtime

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/go-logr/logr"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/client"
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

	decisionPipelineQueueSize = 400
)

// The default implementation of the PolarisScheduler.
type DefaultPolarisScheduler struct {
	config            *config.SchedulerConfig
	clusterClientsMgr client.ClusterClientsManager
	pluginsFactory    pipeline.PluginsFactory
	podSource         pipeline.PodSource

	// The scheduling queue, which is sorted by the SortPlugin.
	schedQueue queue.SchedulingQueue

	// Contains pods, for which nodes have been sampled and which are now waiting to enter the decision pipeline.
	decisionPipelineQueue chan *pipeline.SampledPodInfo

	// The number of pods currently in the SampleNodes stage.
	// This field must be read/written using atomic operations.
	podsInSampling int32

	// The number of pods currently waiting to enter the decision pipeline.
	// This field must be read/written using atomic operations.
	podsWaitingForDecisionPipeline int32

	// The number of pods currently in the decision pipeline.
	// This field must be read/written using atomic operations.
	podsInDecisionPipeline int32

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
func NewDefaultPolarisScheduler(
	conf *config.SchedulerConfig,
	pluginsRegistry *pipeline.PluginsRegistry,
	podSource pipeline.PodSource,
	clusterClientsMgr client.ClusterClientsManager,
	logger *logr.Logger,
) *DefaultPolarisScheduler {
	config.SetDefaultsSchedulerConfig(conf)
	log := logger.WithName("DefaultPolarisScheduler")

	scheduler := DefaultPolarisScheduler{
		config:                conf,
		clusterClientsMgr:     clusterClientsMgr,
		pluginsFactory:        NewDefaultPluginsFactory(pluginsRegistry),
		podSource:             podSource,
		decisionPipelineQueue: make(chan *pipeline.SampledPodInfo, decisionPipelineQueueSize),
		stopCh:                make(chan bool, 1),
		state:                 pristine,
		logger:                &log,
	}
	return &scheduler
}

func (ps *DefaultPolarisScheduler) Config() *config.SchedulerConfig {
	return ps.config
}

func (ps *DefaultPolarisScheduler) ClusterClientsManager() client.ClusterClientsManager {
	return ps.clusterClientsMgr
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

	// Start the scheduling queue.
	sortPlugin, err := ps.pluginsFactory.NewSortPlugin(ps)
	if err != nil {
		return err
	}
	ps.schedQueue = queue.NewPrioritySchedulingQueue(sortPlugin.Less)
	go ps.pumpIntoQueue(ps.podSource)

	// Start the sampling goroutines.
	for i := 0; i < int(ps.config.ParallelNodeSamplers); i++ {
		sampleNodesPlugin, err := ps.pluginsFactory.NewSampleNodesPlugin(ps)
		if err != nil {
			ps.Stop()
			return err
		}
		go ps.executeSamplingLoop(i, sampleNodesPlugin)
	}

	// Start the decision pipelines.
	for i := 0; i < int(ps.config.ParallelDecisionPipelines); i++ {
		decisionPipelinePlugins, err := ps.pluginsFactory.NewDecisionPipelinePlugins(ps)
		if err != nil {
			ps.Stop()
			return err
		}
		decisionPipeline := NewDefaultDecisionPipeline(i, decisionPipelinePlugins, ps)
		go ps.executeDecisionPipelinePump(i, decisionPipeline)
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

func (ps *DefaultPolarisScheduler) PodsInQueueCount() int {
	return ps.schedQueue.Len()
}

func (ps *DefaultPolarisScheduler) PodsInNodeSamplingCount() int {
	return int(atomic.LoadInt32(&ps.podsInSampling))
}

func (ps *DefaultPolarisScheduler) PodsWaitingForDecisionPipelineCount() int {
	return int(atomic.LoadInt32(&ps.podsWaitingForDecisionPipeline))
}

func (ps *DefaultPolarisScheduler) PodsInDecisionPipelineCount() int {
	return int(atomic.LoadInt32(&ps.podsInDecisionPipeline))
}

func (ps *DefaultPolarisScheduler) validateConfig() error {
	pluginsList := ps.config.Plugins

	if pluginsList.Sort == nil {
		return fmt.Errorf("cannot start DefaultPolarisScheduler, because no SortPlugin is configured")
	}

	if pluginsList.SampleNodes == nil {
		return fmt.Errorf("cannot start DefaultPolarisScheduler, because no SampleNodesPlugin is configured")
	}

	if len(pluginsList.Filter) == 0 {
		return fmt.Errorf("cannot start DefaultPolarisScheduler, because no FilterPlugin is configured")
	}

	if len(pluginsList.Score) == 0 {
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

// Continuously retrieves pods from the scheduling queue, samples nodes for each pod,
// and then adds the pod to the decisionPipelineQueue.
func (ps *DefaultPolarisScheduler) executeSamplingLoop(id int, sampler pipeline.SampleNodesPlugin) {
	ps.logger.Info("starting SampleNodesPlugin", "id", id)

	for {
		pod := ps.schedQueue.Dequeue()
		if pod == nil {
			break
		}

		atomic.AddInt32(&ps.podsInSampling, 1)
		ps.sampleNodesForPod(pod, sampler)
		atomic.AddInt32(&ps.podsInSampling, -1)
	}

	ps.logger.Info("stopped DefaultPolarisScheduler SampleNodesPlugin", "id", id)
}

// Uses the SampleNodesPlugin (sampler) to sample nodes for the specified pod and, if successful, adds the sampled pod info
// to the decisionPipelineQueue. If an error occurs, the error information is committed to the Pod object in the cluster.
func (ps *DefaultPolarisScheduler) sampleNodesForPod(pod *pipeline.QueuedPodInfo, sampler pipeline.SampleNodesPlugin) error {
	candidateNodes, status := sampler.SampleNodes(pod.Ctx, pod.PodInfo, ps.config)
	if !pipeline.IsSuccessStatus(status) {
		return ps.handleFailureStatus(pipeline.SampleNodesStage, sampler, pod.Ctx, pod.PodInfo, status)
	}
	if len(candidateNodes) == 0 {
		status := pipeline.NewStatus(pipeline.Unschedulable, "the SampleNodesPlugin returned 0 nodes")
		return ps.handleFailureStatus(pipeline.SampleNodesStage, sampler, pod.Ctx, pod.PodInfo, status)
	}

	sampledPod := pipeline.SampledPodInfo{
		QueuedPodInfo: pod,
		SampledNodes:  candidateNodes,
	}
	atomic.AddInt32(&ps.podsWaitingForDecisionPipeline, 1)
	ps.decisionPipelineQueue <- &sampledPod
	return nil
}

// Continuously retrieves pods from the decision pipeline queue and pumps each of them through
// a single decision pipeline.
func (ps *DefaultPolarisScheduler) executeDecisionPipelinePump(id int, decisionPipeline pipeline.DecisionPipeline) {
	ps.logger.Info("starting DefaultPolarisScheduler DecisionPipeline", "id", id)

	for {
		select {
		case pod := <-ps.decisionPipelineQueue:
			atomic.AddInt32(&ps.podsInDecisionPipeline, 1)
			decision, status := decisionPipeline.SchedulePod(pod)
			if pipeline.IsSuccessStatus(status) {
				ps.commitSchedulingDecision(pod.Ctx, decision)
			} else {
				ps.handleFailureStatus(status.FailedStage(), status.FailedPlugin(), pod.Ctx, pod.PodInfo, status)
			}
			atomic.AddInt32(&ps.podsInDecisionPipeline, -1)

		case <-ps.stopCh:
			// Stop signal received, so we stop the scheduler.
			ps.logger.Info("stopped DefaultPolarisScheduler DecisionPipeline", "id", id)
			return
		}
	}
}

func (ps *DefaultPolarisScheduler) handleFailureStatus(stage string, plugin pipeline.Plugin, schedCtx pipeline.SchedulingContext, podInfo *pipeline.PodInfo, status pipeline.Status) error {
	pod := podInfo.Pod

	// ToDo: which cluster client should we get - the pod might not have been assigned to a cluster yet.
	clusterClient, err := ps.clusterClientsMgr.GetClusterClient("ToDo")
	if err != nil {
		ps.logger.Error(err, "could not obtain ClusterClient")
		return err
	}
	eventRecorder := clusterClient.EventRecorder()

	msg := status.Message()
	eventRecorder.Eventf(pod, nil, core.EventTypeWarning, "FailedScheduling", "Scheduling", msg)
	return nil
}

// Commits the decision of the scheduling pipeline to the orchestrator.
func (ps *DefaultPolarisScheduler) commitSchedulingDecision(schedCtx pipeline.SchedulingContext, decision *pipeline.SchedulingDecision) {
	pod := decision.Pod.Pod
	pod.Spec.NodeName = decision.SelectedNode.Node.Name

	// ToDo: get client for correct cluster.
	clusterClient, err := ps.clusterClientsMgr.GetClusterClient("ToDo")
	if err != nil {
		ps.logger.Error(err, "could not obtain ClusterClient")
		return
	}

	go ps.savePod(schedCtx, clusterClient, pod)
}

func (ps *DefaultPolarisScheduler) savePod(schedCtx pipeline.SchedulingContext, clusterClient client.ClusterClient, pod *core.Pod) {
	_, err := clusterClient.ClientSet().CoreV1().Pods(pod.Namespace).Update(schedCtx.Context(), pod, meta.UpdateOptions{})
	if err != nil {
		ps.logger.Error(err, "could not update Pod", "pod", pod)
	}
}
