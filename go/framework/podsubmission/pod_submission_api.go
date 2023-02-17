package podsubmission

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	core "k8s.io/api/core/v1"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/client"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/config"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/pipeline"
)

var (
	_ pipeline.PodSource = (*PodSubmissionApi)(nil)
)

// A PodSource that exposes and external API for submitting pods in an orchestrator-independent manner.
type PodSubmissionApi struct {
	incomingPods chan *pipeline.IncomingPod
}

func NewPodSubmissionApi(schedConfig *config.SchedulerConfig) *PodSubmissionApi {
	ps := &PodSubmissionApi{
		incomingPods: make(chan *pipeline.IncomingPod, schedConfig.IncomingPodsBufferSize),
	}
	return ps
}

func (ps *PodSubmissionApi) IncomingPods() chan *pipeline.IncomingPod {
	return ps.incomingPods
}

// Registers the submit pod endpoint with the specified gin engine under the specified path (relative to the engine's root path).
func (ps *PodSubmissionApi) RegisterSubmitPodEndpoint(relativePath string, ginEngine *gin.Engine) error {
	ginEngine.POST(relativePath, ps.handleSubmitPodRequest)
	return nil
}

// Registers the status endpoint (relative to the engine's root path).
func (ps *PodSubmissionApi) RegisterStatusEndpoint(relativePath string, ginEngine *gin.Engine) error {
	ginEngine.GET(relativePath, ps.handleStatusRequest)
	return nil
}

func (ps *PodSubmissionApi) handleSubmitPodRequest(c *gin.Context) {
	var pod core.Pod

	if err := c.Bind(&pod); err != nil {
		podSubmissionError := &PodSubmissionApiError{Error: client.NewPolarisErrorDto(err)}
		c.JSON(http.StatusBadRequest, podSubmissionError)
		return
	}

	incomingPod := &pipeline.IncomingPod{
		Pod:        &pod,
		ReceivedAt: time.Now(),
	}

	ps.incomingPods <- incomingPod
	c.JSON(http.StatusCreated, &pod)
}

func (ps *PodSubmissionApi) handleStatusRequest(c *gin.Context) {
	status := map[string]string{
		"application": "polaris-scheduler",
		"status":      "ok",
	}
	c.JSON(http.StatusOK, status)
}
