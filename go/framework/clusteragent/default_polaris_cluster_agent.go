package clusteragent

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/go-logr/logr"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/client"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/config"
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
	ctx           context.Context
	config        *config.ClusterAgentConfig
	clusterClient client.ClusterClient
	ginEngine     *gin.Engine
	logger        *logr.Logger
}

func NewDefaultPolarisClusterAgent(
	clusterAgentConfig *config.ClusterAgentConfig,
	ginEngine *gin.Engine,
	clusterClient client.ClusterClient,
	logger *logr.Logger,
) *DefaultPolarisClusterAgent {
	cb := &DefaultPolarisClusterAgent{
		config:        clusterAgentConfig,
		ginEngine:     ginEngine,
		clusterClient: clusterClient,
		logger:        logger,
	}

	return cb
}

func (cb *DefaultPolarisClusterAgent) ClusterClient() client.ClusterClient {
	return cb.clusterClient
}

func (cb *DefaultPolarisClusterAgent) Config() *config.ClusterAgentConfig {
	return cb.config
}

func (cb *DefaultPolarisClusterAgent) Logger() *logr.Logger {
	return cb.logger
}

func (cb *DefaultPolarisClusterAgent) Start(ctx context.Context) error {
	if cb.ctx != nil {
		return fmt.Errorf("this DefaultPolarisClusterAgent is already running")
	}
	cb.ctx = ctx

	apiPath, err := url.JoinPath(ClusterAgentEndpointsPrefix, CommitSchedulingDecisionEndpoint)
	if err != nil {
		panic(err)
	}
	cb.ginEngine.POST(apiPath, cb.handlePostSchedulingDecision)

	return nil
}

func (cb *DefaultPolarisClusterAgent) handlePostSchedulingDecision(c *gin.Context) {
	var schedDecision client.ClusterSchedulingDecision

	if err := c.Bind(&schedDecision); err != nil {
		agentError := &PolarisClusterAgentError{Error: client.NewPolarisErrorDto(err)}
		c.JSON(http.StatusBadRequest, agentError)
		return
	}

	if err := cb.clusterClient.CommitSchedulingDecision(cb.ctx, &schedDecision); err != nil {
		cb.logger.Info("SchedulingDecisionCommitFailed", "reason", err)
		agentError := &PolarisClusterAgentError{Error: client.NewPolarisErrorDto(err)}
		c.JSON(http.StatusInternalServerError, agentError)
		return
	}

	c.JSON(http.StatusCreated, &schedDecision)
}
