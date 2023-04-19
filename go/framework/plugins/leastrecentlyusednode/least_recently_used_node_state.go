package leastrecentlyusednode

import (
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/pipeline"
)

var (
	_ pipeline.StateData = (*leastRecentlyUsedNodeState)(nil)
)

const (
	stateKey = PluginName + ".state"
)

type leastRecentlyUsedNodeState struct {
	// The Unix timestamp when the scoring phase started.
	// This ensures that all node's timestamps are assessed relative to the same point in time.
	scoringStart int64
}
