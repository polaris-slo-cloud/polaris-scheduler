package sampling

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/go-logr/logr"

	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/client"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/config"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/remotesampling"
)

const (
	SamplingEndpointsPrefix = "samples"
)

var (
	_ PolarisNodeSampler = (*DefaultPolarisNodeSampler)(nil)
)

// Default implementation of the PolarisNodeSampler.
//
// This implementation relies on Gin for providing a REST interface.
// Since the Gin engine can be shared among multiple services, it must be started
// by the owner of this object after calling Start().
type DefaultPolarisNodeSampler struct {
	ctx                       context.Context
	config                    *config.ClusterAgentConfig
	clusterClient             client.ClusterClient
	samplingStrategyFactories []SamplingStrategyFactoryFunc
	samplingStrategies        []SamplingStrategy
	nodesCache                client.NodesCache
	ginEngine                 *gin.Engine
	logger                    *logr.Logger
}

type defaultPolarisNodeSamplerStatus struct {
	Status     string         `json:"status" yaml:"status"`
	Routes     gin.RoutesInfo `json:"routes" yaml:"routes"`
	NodesCount int            `json:"nodesCount" yaml:"nodesCount"`
}

func NewDefaultPolarisNodeSampler(
	clusterAgentConfig *config.ClusterAgentConfig,
	ginEngine *gin.Engine,
	clusterClient client.ClusterClient,
	nodesCache client.NodesCache,
	samplingStrategyFactories []SamplingStrategyFactoryFunc,
	logger *logr.Logger,
) *DefaultPolarisNodeSampler {
	sampler := &DefaultPolarisNodeSampler{
		config:                    clusterAgentConfig,
		ginEngine:                 ginEngine,
		clusterClient:             clusterClient,
		samplingStrategyFactories: samplingStrategyFactories,
		samplingStrategies:        make([]SamplingStrategy, 0, len(samplingStrategyFactories)),
		nodesCache:                nodesCache,
		logger:                    logger,
	}
	return sampler
}

func (sampler *DefaultPolarisNodeSampler) Config() *config.ClusterAgentConfig {
	return sampler.config
}

func (sampler *DefaultPolarisNodeSampler) ClusterClient() client.ClusterClient {
	return sampler.clusterClient
}

func (sampler *DefaultPolarisNodeSampler) NodesCache() client.NodesCache {
	return sampler.nodesCache
}

func (sampler *DefaultPolarisNodeSampler) SamplingStrategies() []SamplingStrategy {
	return sampler.samplingStrategies
}

func (sampler *DefaultPolarisNodeSampler) Logger() *logr.Logger {
	return sampler.logger
}

func (sampler *DefaultPolarisNodeSampler) Start(ctx context.Context) error {
	if sampler.ctx != nil {
		return fmt.Errorf("this DefaultPolarisNodeSampler is already running")
	}
	sampler.ctx = ctx

	if err := sampler.nodesCache.StartWatch(ctx); err != nil {
		return err
	}
	sampler.logger.Info("Successfully populated the nodes cache", "nodes", sampler.getNodesCount())

	for _, factoryFunc := range sampler.samplingStrategyFactories {
		if err := sampler.createAndRegisterSamplingStrategy(factoryFunc); err != nil {
			return err
		}
	}

	sampler.ginEngine.GET("/samples/status", func(c *gin.Context) {
		sampler.handleStatusRequest(c)
	})

	return nil
}

func (sampler *DefaultPolarisNodeSampler) getNodesCount() int {
	reader := sampler.nodesCache.Nodes().ReadLock()
	defer reader.Unlock()
	return reader.Len()
}

func (sampler *DefaultPolarisNodeSampler) createAndRegisterSamplingStrategy(factoryFunc SamplingStrategyFactoryFunc) error {
	strategy, err := factoryFunc(sampler)
	if err != nil {
		return err
	}
	sampler.samplingStrategies = append(sampler.samplingStrategies, strategy)

	apiPath, err := url.JoinPath(SamplingEndpointsPrefix, strategy.Name())
	if err != nil {
		panic(err)
	}

	sampler.ginEngine.POST(apiPath, func(c *gin.Context) {
		sampler.handleSamplingRequest(c, strategy)
	})

	return nil
}

func (sampler *DefaultPolarisNodeSampler) handleSamplingRequest(c *gin.Context, strategy SamplingStrategy) {
	var samplingReq remotesampling.RemoteNodesSamplerRequest

	if err := c.ShouldBind(&samplingReq); err != nil {
		samplingError := &remotesampling.RemoteNodesSamplerError{Error: err}
		c.JSON(http.StatusBadRequest, samplingError)
		return
	}

	if samplingResp, err := strategy.SampleNodes(&samplingReq); err == nil {
		c.JSON(http.StatusOK, samplingResp)
	} else {
		samplingError := &remotesampling.RemoteNodesSamplerError{Error: err}
		c.JSON(http.StatusInternalServerError, samplingError)
	}
}

func (sampler *DefaultPolarisNodeSampler) handleStatusRequest(c *gin.Context) {
	status := &defaultPolarisNodeSamplerStatus{
		Status:     "ok",
		Routes:     sampler.ginEngine.Routes(),
		NodesCount: sampler.getNodesCount(),
	}
	c.JSON(http.StatusOK, status)
}
