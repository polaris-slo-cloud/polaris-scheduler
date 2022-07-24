package runtime

import (
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/pipeline"
)

var (
	_ pipeline.DecisionPipeline = (*DefaultDecisionPipeline)(nil)
)

// Default implementation of the Polaris DecisionPipeline
type DefaultDecisionPipeline struct {
	id        int
	plugins   *pipeline.DecisionPipelinePlugins
	scheduler pipeline.PolarisScheduler
}

// Creates a new instance of the DefaultDecisionPipeline.
func NewDefaultDecisionPipeline(id int, plugins *pipeline.DecisionPipelinePlugins, scheduler pipeline.PolarisScheduler) *DefaultDecisionPipeline {
	decisionPipeline := DefaultDecisionPipeline{
		id:        id,
		plugins:   plugins,
		scheduler: scheduler,
	}
	return &decisionPipeline
}

func (dp *DefaultDecisionPipeline) SchedulePod(podInfo *pipeline.SampledPodInfo) (*pipeline.SchedulingDecision, pipeline.Status) {
	panic("unimplemented")
}
