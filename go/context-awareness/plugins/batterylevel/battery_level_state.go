package batterylevel

import (
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/pipeline"
)

var (
	_ pipeline.StateData = (*batteryLevelState)(nil)
)

const (
	stateKey = PluginName + ".state"
)

type batteryLevelState struct {
	minBatteryLevelPercent             int
	minBatteryCapacityMilliAmpereHours *int64
}
