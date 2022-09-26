package runtime

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-logr/logr"

	"polaris-slo-cloud.github.io/polaris-scheduler/v2/cluster-broker/config"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/cluster-broker/sampling"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/client"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/remotesampling"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/k8s-connector/kubernetes"
)

const (
	samplingEndpointsPrefix = "samples"
)

var (
	_ sampling.PolarisNodeSampler = (*DefaultPolarisNodeSampler)(nil)
)

// Default implementation of the PolarisNodeSampler.
type DefaultPolarisNodeSampler struct {
	ctx                       context.Context
	config                    *config.ClusterBrokerConfig
	clusterClient             kubernetes.KubernetesClusterClient
	samplingStrategyFactories []sampling.SamplingStrategyFactoryFunc
	samplingStrategies        []sampling.SamplingStrategy
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
	clusterBrokerConfig *config.ClusterBrokerConfig,
	clusterClient kubernetes.KubernetesClusterClient,
	samplingStrategyFactories []sampling.SamplingStrategyFactoryFunc,
	logger *logr.Logger,
) *DefaultPolarisNodeSampler {
	updateInterval, err := time.ParseDuration(fmt.Sprintf("%vms", clusterBrokerConfig.NodesCacheUpdateIntervalMs))
	if err != nil {
		panic(err)
	}

	sampler := &DefaultPolarisNodeSampler{
		config:                    clusterBrokerConfig,
		clusterClient:             clusterClient,
		samplingStrategyFactories: samplingStrategyFactories,
		samplingStrategies:        make([]sampling.SamplingStrategy, 0, len(samplingStrategyFactories)),
		nodesCache:                kubernetes.NewKubernetesNodesCache(clusterClient, updateInterval, int(clusterBrokerConfig.NodesCacheUpdateQueueSize)),
		logger:                    logger,
	}
	return sampler
}

func (sampler *DefaultPolarisNodeSampler) Config() *config.ClusterBrokerConfig {
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

	for _, factoryFunc := range sampler.samplingStrategyFactories {
		if err := sampler.createAndRegisterSamplingStrategy(factoryFunc); err != nil {
			return err
		}
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

func (sampler *DefaultPolarisNodeSampler) createAndRegisterSamplingStrategy(factoryFunc sampling.SamplingStrategyFactoryFunc) error {
	strategy, err := factoryFunc(sampler)
	if err != nil {
		return err
	}
	sampler.samplingStrategies = append(sampler.samplingStrategies, strategy)

	apiPath, err := url.JoinPath(samplingEndpointsPrefix, strategy.Name())
	if err != nil {
		panic(err)
	}

	sampler.ginEngine.POST(apiPath, func(c *gin.Context) {
		sampler.handleSamplingRequest(c, strategy)
	})

	return nil
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
