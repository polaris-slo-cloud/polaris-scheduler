package resourcesfit

import (
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/pipeline"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/util"
)

var (
	_ pipeline.StateData = (*resourcesFitState)(nil)
)

const (
	stateKey = PluginName + ".state"
)

type resourcesFitState struct {
	// The resources required by the pod.
	reqResources *util.Resources
}
