package runtime

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/go-logr/logr"

	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/client"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/config"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/pipeline"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/remotesampling"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/runtime/pluginfactories"
)

const (
	SamplingEndpointsPrefix = "samples"
)

var (
	_ pipeline.PolarisNodeSampler = (*DefaultPolarisNodeSampler)(nil)
)

// Default implementation of the PolarisNodeSampler.
//
// This implementation relies on Gin for providing a REST interface.
// Since the Gin engine can be shared among multiple services, it must be started
// by the owner of this object after calling Start().
type DefaultPolarisNodeSampler struct {
	// The Context that was used to start the sampler and which can be used to stop it.
	ctx context.Context

	// The configuration of the ClusterAgent.
	config *config.ClusterAgentConfig

	// The client used to access the cluster.
	clusterClient client.ClusterClient

	// The factory for creating the sampling plugins.
	pluginsFactory pipeline.SamplingPluginsFactory

	// The instances of the SamplingStrategyPlugins, which are shared among all sampling pipelines.
	samplingStrategies []pipeline.SamplingStrategyPlugin

	// The currently available pipelines.
	// To avoid copying the gin.Context into a sampling queue, we have decided to create a
	// queue of idle sampling pipelines, which can be obtained by a gin handler function.
	samplingPipelines chan pipeline.SamplingPipeline

	// The global cache of nodes in the cluster.
	nodesCache client.NodesCache

	// The Gin engine used to run the REST API.
	ginEngine *gin.Engine

	// The logger.
	logger *logr.Logger
}

type defaultPolarisNodeSamplerStatus struct {
	Status     string   `json:"status" yaml:"status"`
	Routes     []string `json:"routes" yaml:"routes"`
	NodesCount int      `json:"nodesCount" yaml:"nodesCount"`
}

func NewDefaultPolarisNodeSampler(
	clusterAgentConfig *config.ClusterAgentConfig,
	ginEngine *gin.Engine,
	clusterClient client.ClusterClient,
	nodesCache client.NodesCache,
	pluginsRegistry *pipeline.PluginsRegistry[pipeline.ClusterAgentServices],
	logger *logr.Logger,
) *DefaultPolarisNodeSampler {
	sampler := &DefaultPolarisNodeSampler{
		config:            clusterAgentConfig,
		ginEngine:         ginEngine,
		clusterClient:     clusterClient,
		pluginsFactory:    pluginfactories.NewDefaultSamplingPluginsFactory(pluginsRegistry),
		samplingPipelines: make(chan pipeline.SamplingPipeline, clusterAgentConfig.ParallelSamplingPipelines),
		nodesCache:        nodesCache,
		logger:            logger,
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

func (sampler *DefaultPolarisNodeSampler) SamplingStrategies() []pipeline.SamplingStrategyPlugin {
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

	if err := sampler.createSamplingPipelines(); err != nil {
		return err
	}

	if samplingStrategies, err := sampler.pluginsFactory.NewSamplingStrategiesPlugins(sampler); err == nil {
		sampler.samplingStrategies = samplingStrategies
	} else {
		return err
	}
	for _, strategy := range sampler.samplingStrategies {
		if err := sampler.registerSamplingStrategy(strategy); err != nil {
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

func (sampler *DefaultPolarisNodeSampler) createSamplingPipelines() error {
	for i := 0; i < int(sampler.config.ParallelSamplingPipelines); i++ {
		samplingPlugins, err := sampler.pluginsFactory.NewSamplingPipelinePlugins(sampler)
		if err != nil {
			return err
		}
		samplingPipeline := NewDefaultSamplingPipeline(i, samplingPlugins, sampler)
		sampler.samplingPipelines <- samplingPipeline
	}
	return nil
}

func (sampler *DefaultPolarisNodeSampler) registerSamplingStrategy(strategy pipeline.SamplingStrategyPlugin) error {
	apiPath, err := url.JoinPath(SamplingEndpointsPrefix, strategy.StrategyName())
	if err != nil {
		panic(err)
	}

	sampler.ginEngine.POST(apiPath, func(c *gin.Context) {
		sampler.handleSamplingRequest(c, strategy)
	})

	return nil
}

func (sampler *DefaultPolarisNodeSampler) handleSamplingRequest(c *gin.Context, samplingStrategy pipeline.SamplingStrategyPlugin) {
	var samplingReq remotesampling.RemoteNodesSamplerRequest

	if err := c.Bind(&samplingReq); err != nil {
		samplingError := &remotesampling.RemoteNodesSamplerError{Error: client.NewPolarisErrorDto(err)}
		c.JSON(http.StatusBadRequest, samplingError)
		return
	}

	// Dequeue the next available sampling pipeline.
	samplingPipeline := <-sampler.samplingPipelines

	if samplingResp, err := sampler.sampleNodes(&samplingReq, samplingPipeline, samplingStrategy); err == nil {
		c.JSON(http.StatusOK, samplingResp)
	} else {
		samplingError := &remotesampling.RemoteNodesSamplerError{Error: client.NewPolarisErrorDto(err)}
		c.JSON(http.StatusInternalServerError, samplingError)
	}

	// Return the sampling pipeline to the queue.
	sampler.samplingPipelines <- samplingPipeline
}

func (sampler *DefaultPolarisNodeSampler) sampleNodes(
	req *remotesampling.RemoteNodesSamplerRequest,
	samplingPipeline pipeline.SamplingPipeline,
	samplingStrategy pipeline.SamplingStrategyPlugin,
) (*remotesampling.RemoteNodesSamplerResponse, error) {
	schedulingCtx := pipeline.NewSchedulingContext(sampler.ctx)

	sampledNodes, status := samplingPipeline.SampleNodes(schedulingCtx, samplingStrategy, req.PodInfo, int(req.NodesToSampleBp))
	if !pipeline.IsSuccessStatus(status) {
		return nil, status.Error()
	}

	resp := &remotesampling.RemoteNodesSamplerResponse{
		Nodes: sampledNodes,
	}
	return resp, nil
}

func (sampler *DefaultPolarisNodeSampler) handleStatusRequest(c *gin.Context) {
	routes := sampler.ginEngine.Routes()
	routePaths := make([]string, len(routes))
	for i, route := range routes {
		routePaths[i] = route.Path
	}

	status := &defaultPolarisNodeSamplerStatus{
		Status:     "ok",
		Routes:     routePaths,
		NodesCount: sampler.getNodesCount(),
	}
	c.JSON(http.StatusOK, status)
}
