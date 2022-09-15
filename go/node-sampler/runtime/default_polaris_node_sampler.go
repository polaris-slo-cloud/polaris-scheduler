package runtime

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-logr/logr"

	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/client"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/remotesampling"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/k8s-connector/kubernetes"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/node-sampler/config"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/node-sampler/sampling"
)

const (
	samplingEndpointsPrefix = "samples"
)

var (
	_ PolarisNodeSampler = (*DefaultPolarisNodeSampler)(nil)
)

// Default implementation of the PolarisNodeSampler.
type DefaultPolarisNodeSampler struct {
	ctx                context.Context
	config             *config.NodeSamplerConfig
	clusterClient      client.ClusterClient
	samplingStrategies []sampling.SamplingStrategy
	nodesCache         client.NodesCache
	ginEngine          *gin.Engine
	logger             *logr.Logger
}

type defaultPolarisNodeSamplerStatus struct {
	Status     string         `json:"status" yaml:"status"`
	Routes     gin.RoutesInfo `json:"routes" yaml:"routes"`
	NodesCount int            `json:"nodesCount" yaml:"nodesCount"`
}

func NewDefaultPolarisNodeSampler(
	nodeSamplerConfig *config.NodeSamplerConfig,
	clusterClient client.ClusterClient,
	samplingStrategies []sampling.SamplingStrategy,
	logger *logr.Logger,
) *DefaultPolarisNodeSampler {
	updateInterval, err := time.ParseDuration(fmt.Sprintf("%vms", nodeSamplerConfig.NodesCacheUpdateIntervalMs))
	if err != nil {
		panic(err)
	}

	sampler := &DefaultPolarisNodeSampler{
		config:             nodeSamplerConfig,
		clusterClient:      clusterClient,
		samplingStrategies: samplingStrategies,
		nodesCache:         kubernetes.NewKubernetesNodesCache(clusterClient, updateInterval, int(nodeSamplerConfig.NodesCacheUpdateQueueSize)),
		logger:             logger,
	}
	return sampler
}

func (sampler *DefaultPolarisNodeSampler) Config() *config.NodeSamplerConfig {
	return sampler.config
}

func (sampler *DefaultPolarisNodeSampler) ClusterClient() client.ClusterClient {
	return sampler.clusterClient
}

func (sampler *DefaultPolarisNodeSampler) NodesCache() client.NodesCache {
	return sampler.nodesCache
}

func (sampler *DefaultPolarisNodeSampler) SamplingStrategies() []sampling.SamplingStrategy {
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

	sampler.ginEngine = gin.Default()
	sampler.ginEngine.SetTrustedProxies(nil)

	for _, strategy := range sampler.samplingStrategies {
		sampler.registerSamplingStrategy(strategy)
	}

	sampler.ginEngine.GET("/status", func(c *gin.Context) {
		sampler.handleStatusRequest(c)
	})

	go func() {
		if err := sampler.ginEngine.Run(sampler.config.ListenOn...); err != nil {
			sampler.logger.Error(err, "Error executing HTTP server.")
		}
	}()
	return nil
}

func (sampler *DefaultPolarisNodeSampler) getNodesCount() int {
	reader := sampler.nodesCache.Nodes().ReadLock()
	defer reader.Unlock()
	return reader.Len()
}

func (sampler *DefaultPolarisNodeSampler) registerSamplingStrategy(strategy sampling.SamplingStrategy) {
	apiPath, err := url.JoinPath(samplingEndpointsPrefix, strategy.Name())
	if err != nil {
		panic(err)
	}

	sampler.ginEngine.POST(apiPath, func(c *gin.Context) {
		sampler.handleSamplingRequest(c, strategy)
	})
}

func (sampler *DefaultPolarisNodeSampler) handleSamplingRequest(c *gin.Context, strategy sampling.SamplingStrategy) {
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
