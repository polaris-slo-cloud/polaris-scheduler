package clusteragent

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/go-logr/logr"

	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/client"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/collections"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/config"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/pipeline"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/runtime"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/runtime/pluginfactories"
)

const (
	// The prefix of all agent endpoints.
	// Thus, the agent endpoints are reachable under "/agent/*"
	ClusterAgentEndpointsPrefix = "agent"

	// The endpoint name for committing a scheduling decision.
	CommitSchedulingDecisionEndpoint = "scheduling-decision"
)

var (
	_ PolarisClusterAgent = (*DefaultPolarisClusterAgent)(nil)
)

// ToDo: Extract all APIs (cluster agent, remote sampling, pod submission) into distinct classes (like we have for the pod submission).
// Then we can move all of them into a distinct api package.

// Default, orchestrator-independent implementation of the PolarisClusterAgent.
//
// This service will expose a REST API at /agent/*
// All orchestrator-specific cluster interactions are handled by the clusterClient.
type DefaultPolarisClusterAgent struct {
	// The Context that was used to start the agent and which can be used to stop it.
	ctx context.Context

	// The configuration of the ClusterAgent.
	config *config.ClusterAgentConfig

	// The client used to access the cluster.
	clusterClient client.LocalClusterClient

	// The currently available binding pipelines.
	// To avoid copying the gin.Context into a queue, we have decided to create a
	// queue of idle binding pipelines, which can be obtained by a gin handler function.
	bindingPipelines chan pipeline.BindingPipeline

	// Used by the binding pipelines to obtain mutually exclusive access to a node.
	nodesLocker collections.EntityLocker

	// The global cache of nodes in the cluster.
	nodesCache client.NodesCache

	// The Gin engine used to run the REST API.
	ginEngine *gin.Engine

	// The factory for creating the binding plugins.
	pluginsFactory pipeline.BindingPluginsFactory

	// The logger.
	logger *logr.Logger
}

func NewDefaultPolarisClusterAgent(
	clusterAgentConfig *config.ClusterAgentConfig,
	ginEngine *gin.Engine,
	clusterClient client.LocalClusterClient,
	nodesCache client.NodesCache,
	pluginRegistry *pipeline.PluginsRegistry[pipeline.ClusterAgentServices],
	logger *logr.Logger,
) *DefaultPolarisClusterAgent {
	cb := &DefaultPolarisClusterAgent{
		config:           clusterAgentConfig,
		ginEngine:        ginEngine,
		clusterClient:    clusterClient,
		bindingPipelines: make(chan pipeline.BindingPipeline, clusterAgentConfig.ParallelBindingPipelines),
		nodesLocker:      collections.NewEntityLockerImpl(),
		nodesCache:       nodesCache,
		pluginsFactory:   pluginfactories.NewDefaultBindingPluginsFactory(pluginRegistry),
		logger:           logger,
	}

	return cb
}

func (ca *DefaultPolarisClusterAgent) ClusterClient() client.LocalClusterClient {
	return ca.clusterClient
}

func (ca *DefaultPolarisClusterAgent) Config() *config.ClusterAgentConfig {
	return ca.config
}

func (ca *DefaultPolarisClusterAgent) NodesCache() client.NodesCache {
	return ca.nodesCache
}

func (ca *DefaultPolarisClusterAgent) Logger() *logr.Logger {
	return ca.logger
}

func (ca *DefaultPolarisClusterAgent) Start(ctx context.Context) error {
	if ca.ctx != nil {
		return fmt.Errorf("this DefaultPolarisClusterAgent is already running")
	}
	ca.ctx = ctx

	if err := ca.createBindingPipelines(); err != nil {
		return err
	}

	apiPath, err := url.JoinPath(ClusterAgentEndpointsPrefix, CommitSchedulingDecisionEndpoint)
	if err != nil {
		return err
	}
	ca.ginEngine.POST(apiPath, ca.handlePostSchedulingDecision)

	return nil
}

func (ca *DefaultPolarisClusterAgent) createBindingPipelines() error {
	for i := 0; i < int(ca.config.ParallelBindingPipelines); i++ {
		bindingPlugins, err := ca.pluginsFactory.NewBindingPipelinePlugins(ca)
		if err != nil {
			return err
		}
		bindingPipeline := runtime.NewDefaultBindingPipeline(i, bindingPlugins, ca, ca.nodesLocker)
		ca.bindingPipelines <- bindingPipeline
	}
	return nil
}

func (ca *DefaultPolarisClusterAgent) handlePostSchedulingDecision(c *gin.Context) {
	var schedDecision client.ClusterSchedulingDecision

	if err := c.Bind(&schedDecision); err != nil {
		agentError := &PolarisClusterAgentError{Error: client.NewPolarisErrorDto(err)}
		c.JSON(http.StatusBadRequest, agentError)
		return
	}

	status := ca.runBindingPipeline(&schedDecision)
	if !pipeline.IsSuccessStatus(status) {
		err := status.Error()
		if err == nil {
			err = fmt.Errorf("SchedulingDecisionCommitFailed %v", status.Reasons())
		}
		ca.logger.Info("SchedulingDecisionCommitFailed", "reason", status.Reasons())
		agentError := &PolarisClusterAgentError{Error: client.NewPolarisErrorDto(err)}
		c.JSON(http.StatusInternalServerError, agentError)
		return
	}

	c.JSON(http.StatusCreated, &schedDecision)
}

func (ca *DefaultPolarisClusterAgent) runBindingPipeline(schedDecision *client.ClusterSchedulingDecision) pipeline.Status {
	stopwatches := runtime.NewBindingPipelineStopwatches()
	stopwatches.QueueTime.Start()
	schedCtx := pipeline.NewSchedulingContext(ca.ctx)
	schedCtx.Write(runtime.BindingPipelineStopwatchesStateKey, stopwatches)

	// Wait for the next available binding pipeline.
	bindingPipeline := <-ca.bindingPipelines
	stopwatches.QueueTime.Stop()
	defer func() {
		// Return the binding pipeline to the queue of available pipelines.
		ca.bindingPipelines <- bindingPipeline
	}()

	status := bindingPipeline.CommitSchedulingDecision(schedCtx, schedDecision)
	ca.logStopwatches(schedDecision, stopwatches, status)
	return status
}

func (ca *DefaultPolarisClusterAgent) logStopwatches(schedDecision *client.ClusterSchedulingDecision, stopwatches *runtime.BindingPipelineStopwatches, status pipeline.Status) {
	fullPodName := fmt.Sprintf("%s.%s", schedDecision.Pod.Namespace, schedDecision.Pod.Name)
	ca.logger.Info(
		"BindingComplete",
		"pod", fullPodName,
		"node", schedDecision.NodeName,
		"status", pipeline.StatusCodeAsString(status),
		"queueTimeMs", stopwatches.QueueTime.Duration().Milliseconds(),
		"nodeLockTimeMs", stopwatches.NodeLockTime.Duration().Milliseconds(),
		"fetchNodeMs", stopwatches.FetchNodeInfo.Duration().Milliseconds(),
		"bindingPipelineMs", stopwatches.BindingPipeline.Duration().Milliseconds(),
		"commitMs", stopwatches.CommitDecision.Duration().Milliseconds(),
	)
}
