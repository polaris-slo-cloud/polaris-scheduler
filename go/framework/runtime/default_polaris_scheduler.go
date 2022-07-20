package runtime

import (
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/config"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/pipeline"
)

var (
	_ pipeline.PolarisScheduler = (*DefaultPolarisScheduler)(nil)
)

// The default implementation of the PolarisScheduler.
type DefaultPolarisScheduler struct {
	config    *config.SchedulerConfig
	podSource pipeline.PodSource
}

// Creates a new instance of the default implementation of the PolarisScheduler.
func NewDefaultPolarisScheduler(conf *config.SchedulerConfig, podSource pipeline.PodSource) *DefaultPolarisScheduler {
	config.SetDefaultsSchedulerConfig(conf)
	scheduler := DefaultPolarisScheduler{
		config:    conf,
		podSource: podSource,
	}
	return &scheduler
}

func (ps *DefaultPolarisScheduler) Config() *config.SchedulerConfig {
	return ps.config
}

func (ps *DefaultPolarisScheduler) Start() {
	for i := 0; i < int(ps.config.ParallelSchedulingPipelines); i++ {
		go ps.executePipeline(i)
	}

	// ToDo: Sorting and sending to internal (sorted queue) .
}

func (ps *DefaultPolarisScheduler) executePipeline(id int) {

}
