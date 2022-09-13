package remotesampling

import (
	"context"
	"math"
	"sync"

	"github.com/go-logr/logr"
)

var (
	_ RemoteSamplerClientsManager = (*DefaultRemoteSamplerClientsManager)(nil)
)

type samplingContext struct {
	ctx          context.Context
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
	maxConcurrentRequests int
	samplingReqQueue      chan *queuedSamplingRequest
	logger                *logr.Logger
}

func newSamplingContext(ctx context.Context, request *RemoteNodesSamplerRequest, clustersCount int) *samplingContext {
	var waitGroup sync.WaitGroup
	waitGroup.Add(clustersCount)

	reqCtx, cancelFn := context.WithCancel(ctx)
	defer cancelFn()

	samplingCtx := &samplingContext{
		ctx:          reqCtx,
		request:      request,
		waitGroup:    &waitGroup,
		results:      make(map[string]*RemoteNodesSamplerResult, clustersCount),
		resultsMutex: &sync.Mutex{},
	}

	return samplingCtx
}

func NewDefaultRemoteSamplerClientsManager(remoteSamplerURIs map[string]string, samplingStrategy string, maxConcurrentRequests int, logger *logr.Logger) *DefaultRemoteSamplerClientsManager {
	remoteSamplers := make(map[string]RemoteSamplerClient, len(remoteSamplerURIs))
	for clusterName, uri := range remoteSamplerURIs {
		remoteSamplers[clusterName] = NewDefaultRemoteSamplerClient(clusterName, uri, samplingStrategy, logger)
	}

	scm := &DefaultRemoteSamplerClientsManager{
		remoteSamplers:        remoteSamplers,
		maxConcurrentRequests: maxConcurrentRequests,
		logger:                logger,
	}

	return scm
}

func (scm *DefaultRemoteSamplerClientsManager) SampleNodesFromAllClusters(ctx context.Context, request *RemoteNodesSamplerRequest) (map[string]*RemoteNodesSamplerResult, error) {
	scm.ensureSamplingRoutinesStarted()
	samplingCtx := newSamplingContext(ctx, request, len(scm.remoteSamplers))

	for _, clusterSampler := range scm.remoteSamplers {
		queuedReq := &queuedSamplingRequest{
			ctx:            samplingCtx,
			clusterSampler: clusterSampler,
		}
		scm.samplingReqQueue <- queuedReq
	}

	samplingCtx.waitGroup.Wait()
	return samplingCtx.results, nil
}

func (scm *DefaultRemoteSamplerClientsManager) sampleSingleCluster(samplingCtx *samplingContext, clusterSampler RemoteSamplerClient) {
	result := &RemoteNodesSamplerResult{}

	if response, err := clusterSampler.SampleNodes(samplingCtx.ctx, samplingCtx.request); err != nil {
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
