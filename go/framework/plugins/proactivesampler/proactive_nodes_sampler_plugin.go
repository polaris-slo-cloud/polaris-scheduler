package proactivesampler

import (
	"fmt"

	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/config"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/pipeline"
)

var (
	_ pipeline.SampleNodesPlugin           = (*ProactiveNodesSamplerPlugin)(nil)
	_ pipeline.SchedulingPluginFactoryFunc = NewProactiveNodesSamplerPlugin
)

const (
	PluginName = "ProactiveNodesSampler"
)

// This SampleNodesPlugin proactively samples nodes using a wrapped sampler plugin.
type ProactiveNodesSamplerPlugin struct {
	wrappedSampler     pipeline.SampleNodesPlugin
	nodesToSampleBp    uint32
	cachedSamplesCount int32
	maxSampleAge       int32
}

func NewProactiveNodesSamplerPlugin(pluginConfig config.PluginConfig, scheduler pipeline.PolarisScheduler) (pipeline.Plugin, error) {
	schedulerConfig := scheduler.Config()
	parsedPluginConfig, err := parseAndValidateConfig(pluginConfig, schedulerConfig)
	if err != nil {
		return nil, err
	}
	panic("ToDo: instantiate wrapped plugin (maybe add a method for this to the scheduler)")

	plugin := &ProactiveNodesSamplerPlugin{
		nodesToSampleBp:    schedulerConfig.NodesToSampleBp,
		cachedSamplesCount: parsedPluginConfig.CachedSamplesCount,
		maxSampleAge:       parsedPluginConfig.MaxSampleAgeMs,
	}

	return plugin, nil
}

// Name implements pipeline.SampleNodesPlugin
func (ps *ProactiveNodesSamplerPlugin) Name() string {
	return PluginName
}

// SampleNodes implements pipeline.SampleNodesPlugin
func (ps *ProactiveNodesSamplerPlugin) SampleNodes(ctx pipeline.SchedulingContext, podInfo *pipeline.PodInfo) ([]*pipeline.NodeInfo, pipeline.Status) {
	panic("unimplemented")
	// ToDo
}

func parseAndValidateConfig(rawConfig config.PluginConfig, schedulerConfig *config.SchedulerConfig) (*ProactiveNodesSamplerPluginConfig, error) {
	pluginConfig := &ProactiveNodesSamplerPluginConfig{}

	if wrappedSamplerConfigRaw, err := config.ReadNestedObjectFromPluginConfig(rawConfig, "wrappedSampler"); err == nil {
		if wrappedPluginName, err := config.ReadStringFromPluginConfig(wrappedSamplerConfigRaw, "name"); err != nil {
			if wrappedPluginConfig, err := config.ReadNestedObjectFromPluginConfig(wrappedSamplerConfigRaw, "config"); err != nil {
				pluginConfig.WrappedSampler = &config.PluginsConfigListEntry{
					Name:   wrappedPluginName,
					Config: wrappedPluginConfig,
				}
			}
		}
	}
	if pluginConfig.WrappedSampler == nil {
		return nil, fmt.Errorf("no wrapped sampling plugin configured")
	}

	if _, ok := rawConfig["cachedSamplesCount"]; ok {
		if cachedSamplesCount, err := config.ReadInt32FromPluginConfig(rawConfig, "cachedSamplesCount"); err == nil {
			pluginConfig.CachedSamplesCount = cachedSamplesCount
		} else {
			return nil, err
		}
	} else {
		// This value is optional, so we provide a default, if it is not set.
		pluginConfig.CachedSamplesCount = DefaultCachedSamplesCount
	}

	if _, ok := rawConfig["maxSampleAgeMs"]; ok {
		if maxSampleAgeMs, err := config.ReadInt32FromPluginConfig(rawConfig, "maxSampleAgeMs"); err == nil {
			pluginConfig.MaxSampleAgeMs = maxSampleAgeMs
		} else {
			return nil, err
		}
	} else {
		// This value is optional, so we provide a default, if it is not set.
		pluginConfig.MaxSampleAgeMs = DefaultMaxSampleAgeMs
	}

	return pluginConfig, nil
}
