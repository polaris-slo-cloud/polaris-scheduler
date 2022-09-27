package broker

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
	// The prefix of all broker endpoints.
	// Thus, the broker endpoints are reachable under "/broker/*"
	BrokerEndpointsPrefix = "broker"

	// The endpoint name for committing a scheduling decision.
	CommitSchedulingDecisionEndpoint = "scheduling-decision"
)

var (
	_ PolarisClusterBroker = (*DefaultPolarisClusterBroker)(nil)
)

// Default, orchestrator-independent implementation of the PolarisClusterBroker.
//
// This service will expose a REST API at /broker/*
// All orchestrator-specific cluster interactions are handled by the clusterClient.
type DefaultPolarisClusterBroker struct {
	ctx           context.Context
	config        *config.ClusterBrokerConfig
	clusterClient client.ClusterClient
	ginEngine     *gin.Engine
	logger        *logr.Logger
}

func NewDefaultPolarisClusterBroker(
	clusterBrokerConfig *config.ClusterBrokerConfig,
	ginEngine *gin.Engine,
	clusterClient client.ClusterClient,
	logger *logr.Logger,
) *DefaultPolarisClusterBroker {
	cb := &DefaultPolarisClusterBroker{
		config:        clusterBrokerConfig,
		ginEngine:     ginEngine,
		clusterClient: clusterClient,
		logger:        logger,
	}

	return cb
}

func (cb *DefaultPolarisClusterBroker) ClusterClient() client.ClusterClient {
	return cb.clusterClient
}

func (cb *DefaultPolarisClusterBroker) Config() *config.ClusterBrokerConfig {
	return cb.config
}

func (cb *DefaultPolarisClusterBroker) Logger() *logr.Logger {
	return cb.logger
}

func (cb *DefaultPolarisClusterBroker) Start(ctx context.Context) error {
	if cb.ctx != nil {
		return fmt.Errorf("this DefaultPolarisClusterBroker is already running")
	}
	cb.ctx = ctx

	apiPath, err := url.JoinPath(BrokerEndpointsPrefix, CommitSchedulingDecisionEndpoint)
	if err != nil {
		panic(err)
	}
	cb.ginEngine.POST(apiPath, cb.handlePostSchedulingDecision)

	return nil
}

func (cb *DefaultPolarisClusterBroker) handlePostSchedulingDecision(c *gin.Context) {
	var schedDecision client.ClusterSchedulingDecision

	if err := c.ShouldBind(&schedDecision); err != nil {
		brokerError := &PolarisClusterBrokerError{Error: err}
		c.JSON(http.StatusBadRequest, brokerError)
		return
	}

	if err := cb.clusterClient.CommitSchedulingDecision(cb.ctx, &schedDecision); err != nil {
		brokerError := &PolarisClusterBrokerError{Error: err}
		c.JSON(http.StatusInternalServerError, brokerError)
		return
	}

	c.JSON(http.StatusCreated, &schedDecision)
}
