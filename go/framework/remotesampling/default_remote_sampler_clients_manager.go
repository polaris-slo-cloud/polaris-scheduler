package remotesampling

import (
	"context"
	"math"
	"sync"

	"github.com/go-logr/logr"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/config"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/util"
)

var (
	_ RemoteSamplerClientsManager = (*DefaultRemoteSamplerClientsManager)(nil)
)

type samplingContext struct {
	ctx          context.Context
	cancelFn     context.CancelFunc
	request      *RemoteNodesSamplerRequest
	waitGroup    *sync.WaitGroup
	results      map[string]*RemoteNodesSamplerResult
	resultsMutex *sync.Mutex
}

type queuedSamplingRequest struct {
	ctx            *samplingContext
	clusterSampler RemoteSamplerClient
}

// Default implementation of RemoteSamplerClientsManager.
type DefaultRemoteSamplerClientsManager struct {
	remoteSamplers        map[string]RemoteSamplerClient
	remoteSamplersList    []RemoteSamplerClient
	maxConcurrentRequests int
	samplingReqQueue      chan *queuedSamplingRequest
	random                util.Random
	logger                *logr.Logger
}

func newSamplingContext(ctx context.Context, request *RemoteNodesSamplerRequest, clustersCount int) *samplingContext {
	var waitGroup sync.WaitGroup
	waitGroup.Add(clustersCount)

	reqCtx, cancelFn := context.WithCancel(ctx)

	samplingCtx := &samplingContext{
		ctx:          reqCtx,
		cancelFn:     cancelFn,
		request:      request,
		waitGroup:    &waitGroup,
		results:      make(map[string]*RemoteNodesSamplerResult, clustersCount),
		resultsMutex: &sync.Mutex{},
	}

	return samplingCtx
}

func NewDefaultRemoteSamplerClientsManager(
	remoteClusters map[string]*config.RemoteClusterConfig,
	samplingStrategy string,
	maxConcurrentRequests int,
	logger *logr.Logger,
) *DefaultRemoteSamplerClientsManager {
	remoteSamplers := make(map[string]RemoteSamplerClient, len(remoteClusters))
	remoteSamplersList := make([]RemoteSamplerClient, 0, len(remoteClusters))
	for clusterName, clusterConfig := range remoteClusters {
		samplerClient := NewDefaultRemoteSamplerClient(clusterName, clusterConfig.BaseURI, samplingStrategy, logger)
		remoteSamplers[clusterName] = samplerClient
		remoteSamplersList = append(remoteSamplersList, samplerClient)
	}

	scm := &DefaultRemoteSamplerClientsManager{
		remoteSamplers:        remoteSamplers,
		remoteSamplersList:    remoteSamplersList,
		maxConcurrentRequests: maxConcurrentRequests,
		random:                util.NewDefaultRandom(),
		logger:                logger,
	}

	return scm
}

func (scm *DefaultRemoteSamplerClientsManager) SampleNodesFromClusters(
	ctx context.Context,
	request *RemoteNodesSamplerRequest,
	percentageOfClustersToSample float64,
) (map[string]*RemoteNodesSamplerResult, error) {
	scm.ensureSamplingRoutinesStarted()

	remoteSamplers := scm.compileClusterSamplersList(percentageOfClustersToSample)
	samplingCtx := newSamplingContext(ctx, request, len(remoteSamplers))
	defer samplingCtx.cancelFn()

	for _, clusterSampler := range remoteSamplers {
		queuedReq := &queuedSamplingRequest{
			ctx:            samplingCtx,
			clusterSampler: clusterSampler,
		}
		scm.samplingReqQueue <- queuedReq
	}

	samplingCtx.waitGroup.Wait()
	return samplingCtx.results, nil
}

func (scm *DefaultRemoteSamplerClientsManager) compileClusterSamplersList(percentageOfClustersToSample float64) []RemoteSamplerClient {
	totalClustersCount := len(scm.remoteSamplersList)
	reqSamplersCount := int(math.Ceil(percentageOfClustersToSample * float64(totalClustersCount)))
	if reqSamplersCount == 0 {
		reqSamplersCount = 1
	}
	if reqSamplersCount == totalClustersCount {
		return scm.remoteSamplersList
	}

	samplers := make([]RemoteSamplerClient, reqSamplersCount)
	chosenIndices := make(map[int]bool, reqSamplersCount)

	for i := 0; i < reqSamplersCount; i++ {
		var randIndex int
		for {
			randIndex = scm.random.Int(totalClustersCount)
			if _, exists := chosenIndices[randIndex]; !exists {
				break
			}
		}
		chosenIndices[randIndex] = true
		samplers[i] = scm.remoteSamplersList[randIndex]
	}

	return samplers
}

func (scm *DefaultRemoteSamplerClientsManager) sampleSingleCluster(samplingCtx *samplingContext, clusterSampler RemoteSamplerClient) {
	result := &RemoteNodesSamplerResult{}

	if response, err := clusterSampler.SampleNodes(samplingCtx.ctx, samplingCtx.request); err == nil {
		result.Response = response
	} else {
		result.Error = err
	}

	samplingCtx.resultsMutex.Lock()
	defer samplingCtx.resultsMutex.Unlock()
	defer samplingCtx.waitGroup.Done()
	samplingCtx.results[clusterSampler.ClusterName()] = result
}

// Starts the sampling goroutines if they have not been started yet.
func (scm *DefaultRemoteSamplerClientsManager) ensureSamplingRoutinesStarted() {
	if scm.samplingReqQueue != nil {
		return
	}

	// SampleNodesFromAllClusters() is blocking anyway, so a buffer of max(maxConcurrentRequests, remoteSamplersCount)
	// should be fine, even if it is called from multiple goroutines.
	bufferSize := int(math.Max(float64(scm.maxConcurrentRequests), float64(len(scm.remoteSamplers))))
	scm.samplingReqQueue = make(chan *queuedSamplingRequest, bufferSize)

	for i := 0; i < scm.maxConcurrentRequests; i++ {
		go scm.runRemoteSamplingLoop(i)
	}
}

func (scm *DefaultRemoteSamplerClientsManager) runRemoteSamplingLoop(id int) {
	for req := range scm.samplingReqQueue {
		scm.sampleSingleCluster(req.ctx, req.clusterSampler)
	}
}
