package leastrecentlyusednode

import (
	"fmt"
	"time"

	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/config"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/pipeline"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/pluginsutil"
)

var (
	_ pipeline.Plugin                        = (*LeastRecentlyUsedNodePlugin)(nil)
	_ pipeline.PreScorePlugin                = (*LeastRecentlyUsedNodePlugin)(nil)
	_ pipeline.ScorePlugin                   = (*LeastRecentlyUsedNodePlugin)(nil)
	_ pipeline.ScoreExtensions               = (*LeastRecentlyUsedNodePlugin)(nil)
	_ pipeline.SchedulingPluginFactoryFunc   = NewLeastRecentlyUsedNodeSchedulingPlugin
	_ pipeline.ClusterAgentPluginFactoryFunc = NewLeastRecentlyUsedNodeClusterAgentPlugin
)

const (
	PluginName = "LeastRecentlyUsedNode"
)

// This plugin assigns higher scores to nodes, which have not received a new pod for a longer time,
// i.e., nodes with an older LastPodAddedTimestamp get higher scores.
type LeastRecentlyUsedNodePlugin struct {
}

func NewLeastRecentlyUsedNodeSchedulingPlugin(configMap config.PluginConfig, scheduler pipeline.PolarisScheduler) (pipeline.Plugin, error) {
	return newLeastRecentlyUsedNodePlugin(), nil
}

func NewLeastRecentlyUsedNodeClusterAgentPlugin(configMap config.PluginConfig, clusterAgentServices pipeline.ClusterAgentServices) (pipeline.Plugin, error) {
	return newLeastRecentlyUsedNodePlugin(), nil
}

func newLeastRecentlyUsedNodePlugin() *LeastRecentlyUsedNodePlugin {
	lru := &LeastRecentlyUsedNodePlugin{}
	return lru
}

func (lru *LeastRecentlyUsedNodePlugin) Name() string {
	return PluginName
}

func (lru *LeastRecentlyUsedNodePlugin) PreScore(ctx pipeline.SchedulingContext, podInfo *pipeline.PodInfo, eligibleNodes []*pipeline.NodeInfo) pipeline.Status {
	state := &leastRecentlyUsedNodeState{
		scoringStart: time.Now().Unix(),
	}
	ctx.Write(stateKey, state)

	return pipeline.NewSuccessStatus()
}

func (lru *LeastRecentlyUsedNodePlugin) ScoreExtensions() pipeline.ScoreExtensions {
	return lru
}

func (lru *LeastRecentlyUsedNodePlugin) Score(ctx pipeline.SchedulingContext, podInfo *pipeline.PodInfo, nodeInfo *pipeline.NodeInfo) (int64, pipeline.Status) {
	state, err := lru.readState(ctx)
	if err != nil {
		return 0, pipeline.NewInternalErrorStatus(err)
	}

	score := state.scoringStart - nodeInfo.Node.LastPodAddedTimestamp
	return score, pipeline.NewSuccessStatus()
}

func (*LeastRecentlyUsedNodePlugin) NormalizeScores(ctx pipeline.SchedulingContext, podInfo *pipeline.PodInfo, scores []pipeline.NodeScore) pipeline.Status {
	pluginsutil.NormalizeScoresGeneric(scores)
	return pipeline.NewSuccessStatus()
}

func (lru *LeastRecentlyUsedNodePlugin) readState(ctx pipeline.SchedulingContext) (*leastRecentlyUsedNodeState, error) {
	state, ok := ctx.Read(stateKey)
	if !ok {
		return nil, fmt.Errorf("%s not found", stateKey)
	}
	resState, ok := state.(*leastRecentlyUsedNodeState)
	if !ok {
		return nil, fmt.Errorf("invalid object stored as %s", stateKey)
	}
	return resState, nil
}
