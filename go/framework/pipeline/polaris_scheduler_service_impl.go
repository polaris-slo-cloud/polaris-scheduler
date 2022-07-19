package pipeline

import (
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/config"
)

var (
	_ PolarisSchedulerService = (*polarisSchedulerServiceImpl)(nil)
)

type polarisSchedulerServiceImpl struct {
	config    *config.SchedulerConfig
	podSource PodSource
}

func newPolarisSchedulerServiceImpl(conf *config.SchedulerConfig, podSource PodSource) *polarisSchedulerServiceImpl {
	config.SetDefaultsSchedulerConfig(conf)
	scheduler := polarisSchedulerServiceImpl{
		config:    conf,
		podSource: podSource,
	}
	return &scheduler
}

func (ps *polarisSchedulerServiceImpl) Config() *config.SchedulerConfig {
	return ps.config
}

func (ps *polarisSchedulerServiceImpl) Start() {
	for i := 0; i < int(ps.config.ParallelSchedulingPipelines); i++ {
		go ps.executePipeline(i)
	}

	// ToDo: Sorting and sending to internal (sorted queue).
}

func (ps *polarisSchedulerServiceImpl) executePipeline(id int) {

}
