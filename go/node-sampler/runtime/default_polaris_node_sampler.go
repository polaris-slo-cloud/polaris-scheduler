package runtime

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-logr/logr"

	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/client"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/k8s-connector/kubernetes"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/node-sampler/config"
)

var (
	_ PolarisNodeSampler = (*DefaultPolarisNodeSampler)(nil)
)

// Default implementation of the PolarisNodeSampler.
type DefaultPolarisNodeSampler struct {
	ctx           context.Context
	config        *config.NodeSamplerConfig
	clusterClient client.ClusterClient
	nodesCache    client.NodesCache
	logger        *logr.Logger
}

func NewDefaultPolarisNodeSampler(
	nodeSamplerConfig *config.NodeSamplerConfig,
	clusterClient client.ClusterClient,
	logger *logr.Logger,
) *DefaultPolarisNodeSampler {
	updateInterval, err := time.ParseDuration(fmt.Sprintf("%vms", nodeSamplerConfig.NodesCacheUpdateIntervalMs))
	if err != nil {
		panic(err)
	}

	sampler := &DefaultPolarisNodeSampler{
		config:        nodeSamplerConfig,
		clusterClient: clusterClient,
		nodesCache:    kubernetes.NewKubernetesNodesCache(clusterClient, updateInterval, int(nodeSamplerConfig.NodesCacheUpdateQueueSize)),
		logger:        logger,
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

	r := gin.Default()
	r.SetTrustedProxies(nil)

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	go func() {
		if err := r.Run(sampler.config.ListenOn...); err != nil {
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
