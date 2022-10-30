package podsubmission

import (
	"net/http"

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
	incomingPods chan *core.Pod
}

func NewPodSubmissionApi(schedConfig *config.SchedulerConfig) *PodSubmissionApi {
	ps := &PodSubmissionApi{
		incomingPods: make(chan *core.Pod, schedConfig.IncomingPodsBufferSize),
	}
	return ps
}

func (ps *PodSubmissionApi) IncomingPods() chan *core.Pod {
	return ps.incomingPods
}

// Registers the submit pod endpoint with the specified gin engine under the specified path (relative to the engine's root path).
func (ps *PodSubmissionApi) RegisterSubmitPodEndpoint(relativePath string, ginEngine *gin.Engine) error {
	ginEngine.POST(relativePath, ps.handleSubmitPodRequest)
	return nil
}

func (ps *PodSubmissionApi) handleSubmitPodRequest(c *gin.Context) {
	var pod core.Pod

	if err := c.Bind(&pod); err != nil {
		podSubmissionError := &PodSubmissionApiError{Error: client.NewPolarisErrorDto(err)}
		c.JSON(http.StatusBadRequest, podSubmissionError)
		return
	}

	ps.incomingPods <- &pod
	c.JSON(http.StatusCreated, &pod)
}
