package runtime

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/go-logr/logr"

	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/client"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/config"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/pipeline"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/runtime/pluginfactories"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/runtime/queue"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/util"
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

	endToEndStopwatchStateKey           = util.StopwatchStateKey + ".e2e"
	schedulingPipelineStopwatchStateKey = util.StopwatchStateKey + ".pipeline"
)

// The default implementation of the PolarisScheduler.
type DefaultPolarisScheduler struct {
	config            *config.SchedulerConfig
	clusterClientsMgr client.ClusterClientsManager
	pluginsFactory    pipeline.SchedulingPluginsFactory
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
	pluginsRegistry *pipeline.PluginsRegistry[pipeline.PolarisScheduler],
	podSource pipeline.PodSource,
	clusterClientsMgr client.ClusterClientsManager,
	logger *logr.Logger,
) *DefaultPolarisScheduler {
	config.SetDefaultsSchedulerConfig(conf)
	log := logger.WithName("DefaultPolarisScheduler")

	scheduler := &DefaultPolarisScheduler{
		config:                conf,
		clusterClientsMgr:     clusterClientsMgr,
		pluginsFactory:        pluginfactories.NewDefaultSchedulingPluginsFactory(pluginsRegistry),
		podSource:             podSource,
		decisionPipelineQueue: make(chan *pipeline.SampledPodInfo, decisionPipelineQueueSize),
		stopCh:                make(chan bool, 1),
		state:                 pristine,
		logger:                &log,
	}

	return scheduler
}

// Gets the scheduler configuration.
func (ps *DefaultPolarisScheduler) Config() *config.SchedulerConfig {
	return ps.config
}

// Gets the ClusterClientsManager for communicating with the node clusters.
func (ps *DefaultPolarisScheduler) ClusterClientsManager() client.ClusterClientsManager {
	return ps.clusterClientsMgr
}

// Starts the scheduling goroutines and then returns nil
// or an error, if any occurred.
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

// Stops the scheduling goroutines.
func (ps *DefaultPolarisScheduler) Stop() error {
	if atomic.CompareAndSwapInt32(&ps.state, started, stopped) {
		close(ps.stopCh)
		ps.schedQueue.Close()
	}
	return nil
}

// Gets the logger used by this scheduler.
func (ps *DefaultPolarisScheduler) Logger() *logr.Logger {
	return ps.logger
}

// Returns true if the scheduling process has been started.
func (ps *DefaultPolarisScheduler) IsActive() bool {
	return atomic.LoadInt32(&ps.state) == started
}

// Returns the number of queued pods.
func (ps *DefaultPolarisScheduler) PodsInQueueCount() int {
	return ps.schedQueue.Len()
}

// Returns the number of pods, for which nodes are currently being sampled.
func (ps *DefaultPolarisScheduler) PodsInNodeSamplingCount() int {
	return int(atomic.LoadInt32(&ps.podsInSampling))
}

// Returns the number of pods, for which nodes have been sampled, and which are
// now waiting to enter the decision pipeline.
func (ps *DefaultPolarisScheduler) PodsWaitingForDecisionPipelineCount() int {
	return int(atomic.LoadInt32(&ps.podsWaitingForDecisionPipeline))
}

// Returns the number of pods currently in the decision pipeline.
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

	// Since the cluster-agent can do filtering and scoring, we do not necessarily need a filter or a score plugin in the scheduler.
	// if len(pluginsList.Filter) == 0 {
	// 	return fmt.Errorf("cannot start DefaultPolarisScheduler, because no FilterPlugin is configured")
	// }

	// if len(pluginsList.Score) == 0 {
	// 	return fmt.Errorf("cannot start DefaultPolarisScheduler, because no ScorePlugin is configured")
	// }

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
func (ps *DefaultPolarisScheduler) addPodToQueue(pod *pipeline.IncomingPod) {
	schedCtx := pipeline.NewSchedulingContext(ps.ctx)
	queuedPod := pipeline.NewQueuedPodInfo(pod.Pod, schedCtx)

	ps.createAndStartStopwatch(schedCtx, schedulingPipelineStopwatchStateKey, nil)
	ps.createAndStartStopwatch(schedCtx, endToEndStopwatchStateKey, &pod.ReceivedAt)

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
	candidateNodes, status := sampler.SampleNodes(pod.Ctx, pod.PodInfo)
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
	fullyQualifiedPodName := pod.Namespace + "." + pod.Name
	ps.logger.Info("FailedScheduling", "pod", fullyQualifiedPodName, "reason", status.Message())

	return nil
}

// Commits the decision of the scheduling pipeline to the orchestrator.
func (ps *DefaultPolarisScheduler) commitSchedulingDecision(schedCtx pipeline.SchedulingContext, decision *pipeline.SchedulingDecision) {
	clusterClient, err := ps.clusterClientsMgr.GetClusterClient(decision.TargetNode.ClusterName)
	if err != nil {
		ps.logger.Error(err, "commitSchedulingDecision() could not obtain ClusterClient")
		return
	}

	pipelineStopwatch := ps.stopStopwatch(schedCtx, schedulingPipelineStopwatchStateKey)
	e2eStopwatch := ps.stopStopwatch(schedCtx, endToEndStopwatchStateKey)

	go ps.commitSchedulingDecisionUsingClient(schedCtx, clusterClient, decision, pipelineStopwatch, e2eStopwatch)
}

func (ps *DefaultPolarisScheduler) commitSchedulingDecisionUsingClient(
	schedCtx pipeline.SchedulingContext,
	clusterClient client.ClusterClient,
	decision *pipeline.SchedulingDecision,
	stoppedPipelineStopwatch *util.Stopwatch,
	stoppedE2EStopwatch *util.Stopwatch,
) {
	clusterSchedDecision := &client.ClusterSchedulingDecision{
		Pod:      decision.Pod.Pod,
		NodeName: decision.TargetNode.Node.Name,
	}

	fullPodName := decision.Pod.Pod.Namespace + "." + decision.Pod.Pod.Name
	targetNode := decision.TargetNode.ClusterName + "." + decision.TargetNode.Node.Name
	pipelineDurationMs := stoppedPipelineStopwatch.Duration().Milliseconds()
	submitPodApiQueueTimeMs := stoppedE2EStopwatch.Duration().Milliseconds() - pipelineDurationMs

	if err := clusterClient.CommitSchedulingDecision(schedCtx.Context(), clusterSchedDecision); err == nil {
		// Stop the stopwatch again to get the full E2E scheduling time.
		stoppedE2EStopwatch.Stop()

		ps.logger.Info(
			"SchedulingSuccess",
			"pod", fullPodName,
			"targetNode", targetNode,
			"submitPodApiQueueTimeMs", submitPodApiQueueTimeMs,
			"pipelineDurationMs", pipelineDurationMs,
			"e2eDurationMs", stoppedE2EStopwatch.Duration().Milliseconds(),
		)
	} else {
		// Stop the stopwatch again to get the full E2E scheduling time.
		stoppedE2EStopwatch.Stop()

		ps.logger.Info(
			"FailedScheduling",
			"pod", fullPodName,
			"targetNode", targetNode,
			"submitPodApiQueueTimeMs", submitPodApiQueueTimeMs,
			"pipelineDurationMs", pipelineDurationMs,
			"e2eDurationMs", stoppedE2EStopwatch.Duration().Milliseconds(),
			"reason", err,
		)
	}
}

// Creates a new stopwatch, starts it (using the specified startTime or the current time, if startTime is nil), and adds it to the SchedulingContext.
func (ps *DefaultPolarisScheduler) createAndStartStopwatch(schedCtx pipeline.SchedulingContext, stateKey string, startTime *time.Time) *util.Stopwatch {
	stopwatch := util.NewStopwatch()
	if startTime != nil {
		stopwatch.StartAt(*startTime)
	} else {
		stopwatch.Start()
	}
	schedCtx.Write(stateKey, stopwatch)
	return stopwatch
}

func (ps *DefaultPolarisScheduler) stopStopwatch(schedCtx pipeline.SchedulingContext, stateKey string) *util.Stopwatch {
	stopwatch, ok, err := pipeline.ReadTypedStateData[*util.Stopwatch](schedCtx, stateKey)
	if !ok || err != nil {
		panic("could not read Stopwatch from SchedulingContext")
	}

	stopwatch.Stop()
	return stopwatch
}
