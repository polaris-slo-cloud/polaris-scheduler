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

	maxRetrySchedulingCount = 10

	// State key for the stopwatch that measures the time from the arrival of the pod through the API until
	// it enters the SampleNodes stage of the scheduling queue.
	queueStopwatchStateKey = util.StopwatchStateKey + ".queue"

	// State key for the stopwatch that measures the time taken by the SampleNodes stage.
	sampleNodesStopwatchStateKey = util.StopwatchStateKey + ".pipeline"

	// State key for the stopwatch that measures the time from the beginning of the SampleNodes stage until the end of the scheduling pipeline.
	schedulingPipelineStopwatchStateKey = util.StopwatchStateKey + ".pipeline"

	// State key for the stopwatch that measures the time form the arrival of the pod through the API until
	// its scheduling decision has been committed or the commit has failed.
	endToEndStopwatchStateKey = util.StopwatchStateKey + ".e2e"
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
				ps.addPodToQueue(pod, 0)
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
// The retryCount indicates how many times this pod has been re-added to the queue after failing scheduling due to a commit failure (count starts at 0 for a newly submitted pod).
func (ps *DefaultPolarisScheduler) addPodToQueue(pod *pipeline.IncomingPod, schedulingRetryCount int) {
	schedCtx := pipeline.NewSchedulingContext(ps.ctx)
	queuedPod := pipeline.NewQueuedPodInfo(pod.Pod, schedCtx, schedulingRetryCount)

	ps.createAndStartStopwatch(schedCtx, queueStopwatchStateKey, &pod.ReceivedAt)
	ps.createAndStartStopwatch(schedCtx, endToEndStopwatchStateKey, &pod.ReceivedAt)

	ps.schedQueue.Enqueue(queuedPod)
}

// Continuously retrieves pods from the scheduling queue, samples nodes for each pod,
// and then adds the pod to the decisionPipelineQueue.
// Entry point for the sampling goroutines.
func (ps *DefaultPolarisScheduler) executeSamplingLoop(id int, sampler pipeline.SampleNodesPlugin) {
	ps.logger.Info("starting SampleNodesPlugin", "id", id)

	for {
		pod := ps.schedQueue.Dequeue()
		if pod == nil {
			break
		}

		ps.stopStopwatch(pod.Ctx, queueStopwatchStateKey)
		ps.createAndStartStopwatch(pod.Ctx, schedulingPipelineStopwatchStateKey, nil)

		atomic.AddInt32(&ps.podsInSampling, 1)
		ps.sampleNodesForPod(pod, sampler)
		atomic.AddInt32(&ps.podsInSampling, -1)
	}

	ps.logger.Info("stopped DefaultPolarisScheduler SampleNodesPlugin", "id", id)
}

// Uses the SampleNodesPlugin (sampler) to sample nodes for the specified pod and, if successful, adds the sampled pod info
// to the decisionPipelineQueue. If an error occurs, the error information is committed to the Pod object in the cluster.
func (ps *DefaultPolarisScheduler) sampleNodesForPod(pod *pipeline.QueuedPodInfo, sampler pipeline.SampleNodesPlugin) error {
	sampleNodesStopwatch := ps.createAndStartStopwatch(pod.Ctx, sampleNodesStopwatchStateKey, nil)
	candidateNodes, status := sampler.SampleNodes(pod.Ctx, pod.PodInfo)
	sampleNodesStopwatch.Stop()

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
// Entry point for the decision pipeline goroutines.
func (ps *DefaultPolarisScheduler) executeDecisionPipelinePump(id int, decisionPipeline pipeline.DecisionPipeline) {
	ps.logger.Info("starting DefaultPolarisScheduler DecisionPipeline", "id", id)

	for {
		select {
		case pod := <-ps.decisionPipelineQueue:
			atomic.AddInt32(&ps.podsInDecisionPipeline, 1)

			candidateDecisions, status := decisionPipeline.DecideCommitCandidates(pod, int(ps.config.CommitCandidateNodes))
			ps.stopStopwatch(pod.Ctx, schedulingPipelineStopwatchStateKey)

			if pipeline.IsSuccessStatus(status) {
				go ps.commitFirstPossibleSchedulingDecision(pod.Ctx, candidateDecisions)
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

// Commits the decision of the scheduling pipeline to the orchestrator, starting with the first candidate decision and returning if it succeeds.
// If the first candidate decision fails, we iteratively try the next decision, continuing until one decision commits successfully or we run out of decisions.
// If we run out of decisions, we try rescheduling the pod completely, if it has not reached it max retries limit.
func (ps *DefaultPolarisScheduler) commitFirstPossibleSchedulingDecision(schedCtx pipeline.SchedulingContext, candidateDecisions []*pipeline.SchedulingDecision) {
	podInfo := candidateDecisions[0].Pod
	fullPodName := podInfo.Pod.Namespace + "." + podInfo.Pod.Name

	queueStopwatch := ps.getStopwatch(schedCtx, queueStopwatchStateKey)
	sampleNodesStopwatch := ps.getStopwatch(schedCtx, sampleNodesStopwatchStateKey)
	pipelineStopwatch := ps.getStopwatch(schedCtx, schedulingPipelineStopwatchStateKey)
	commitStopwatch := util.NewStopwatch()

	commitStopwatch.Start()
	result, commitErrors := ps.tryCommitFirstPossibleSchedulingDecision(schedCtx, candidateDecisions)
	commitStopwatch.Stop()
	e2eStopwatch := ps.stopStopwatch(schedCtx, endToEndStopwatchStateKey)

	if result != nil {
		ps.logger.Info(
			"SchedulingSuccess",
			"pod", fullPodName,
			"targetNode", result.NodeName,
			"unixTimestampMs", time.Now().UnixMilli(),
			"queueTimeMs", queueStopwatch.Duration().Milliseconds(),
			"samplingDurationMs", sampleNodesStopwatch.Duration().Milliseconds(),
			"pipelineDurationMs", pipelineStopwatch.Duration().Milliseconds(),
			"commitDurationMs", commitStopwatch.Duration().Milliseconds(),
			"e2eDurationMs", e2eStopwatch.Duration().Milliseconds(),
			"agentQueueTimeMs", result.Timings.QueueTime,
			"agentNodeLockTimeMs", result.Timings.NodeLockTime,
			"agentFetchNodeInfoMs", result.Timings.FetchNodeInfo,
			"agentBindingPipelineMs", result.Timings.BindingPipeline,
			"agentCreatePodMs", result.Timings.CreatePod,
			"agentCreateBindingMs", result.Timings.CreateBinding,
			"agentCommitDecisionMs", result.Timings.CommitDecision,
			"commitRetries", len(commitErrors),
		)
	} else {
		retryScheduling := podInfo.SchedulingRetryCount < maxRetrySchedulingCount

		ps.logger.Info(
			"FailedScheduling",
			"pod", fullPodName,
			"queueTimeMs", queueStopwatch.Duration().Milliseconds(),
			"samplingDurationMs", sampleNodesStopwatch.Duration().Milliseconds(),
			"pipelineDurationMs", pipelineStopwatch.Duration().Milliseconds(),
			"commitDurationMs", commitStopwatch.Duration().Milliseconds(),
			"e2eDurationMs", e2eStopwatch.Duration().Milliseconds(),
			"reasons", commitErrors,
			"retryCount", podInfo.SchedulingRetryCount,
			"retryingScheduling", retryScheduling,
		)

		if retryScheduling {
			podToRetry := &pipeline.IncomingPod{
				Pod:        podInfo.Pod,
				ReceivedAt: time.Now(),
			}
			ps.addPodToQueue(podToRetry, podInfo.SchedulingRetryCount+1)
		}
	}
}

// Tries to commit the first possible candidate decision.
// Returns the name of the targetNode that the pod was committed to and the list of errors that occurred on failed commits.
func (ps *DefaultPolarisScheduler) tryCommitFirstPossibleSchedulingDecision(schedCtx pipeline.SchedulingContext, candidateDecisions []*pipeline.SchedulingDecision) (*client.CommitSchedulingDecisionSuccess, []error) {
	var commitErrors []error
	for _, decision := range candidateDecisions {
		if result, err := ps.commitSchedulingDecision(schedCtx, decision); err == nil {
			result.NodeName = decision.TargetNode.ClusterName + "." + result.NodeName
			return result, commitErrors
		} else {
			if commitErrors == nil {
				commitErrors = make([]error, 0, len(candidateDecisions))
			}
			commitErrors = append(commitErrors, err)
		}
	}
	return nil, commitErrors
}

func (ps *DefaultPolarisScheduler) commitSchedulingDecision(
	schedCtx pipeline.SchedulingContext,
	decision *pipeline.SchedulingDecision,
) (*client.CommitSchedulingDecisionSuccess, error) {
	clusterClient, err := ps.clusterClientsMgr.GetClusterClient(decision.TargetNode.ClusterName)
	if err != nil {
		ps.logger.Error(err, "commitSchedulingDecision() could not obtain ClusterClient")
		return nil, err
	}

	clusterSchedDecision := &client.ClusterSchedulingDecision{
		Pod:      decision.Pod.Pod,
		NodeName: decision.TargetNode.Node.Name,
	}

	return clusterClient.CommitSchedulingDecision(schedCtx.Context(), clusterSchedDecision)
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

func (ps *DefaultPolarisScheduler) getStopwatch(schedCtx pipeline.SchedulingContext, stateKey string) *util.Stopwatch {
	stopwatch, ok, err := pipeline.ReadTypedStateData[*util.Stopwatch](schedCtx, stateKey)
	if !ok || err != nil {
		panic("could not read Stopwatch from SchedulingContext")
	}
	return stopwatch
}

func (ps *DefaultPolarisScheduler) stopStopwatch(schedCtx pipeline.SchedulingContext, stateKey string) *util.Stopwatch {
	stopwatch := ps.getStopwatch(schedCtx, stateKey)
	stopwatch.Stop()
	return stopwatch
}
