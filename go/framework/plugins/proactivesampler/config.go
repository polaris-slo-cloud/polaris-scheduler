package proactivesampler

import "polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/config"

const (
	DefaultCachedSamplesCount = 100
	DefaultMaxSampleAgeMs     = 500
)

// Configuration data for the ProactiveNodesSamplerPlugin.
type ProactiveNodesSamplerPluginConfig struct {

	// The wrapped sampling plugin that is used to obtain the samples.
	WrappedSampler *config.PluginsConfigListEntry `json:"wrappedSampler" yaml:"wrappedSampler"`

	// The number of samples that should be cached.
	CachedSamplesCount int32 `json:"cachedSamplesCount" yaml:"cachedSamplesCount"`

	// The maximum age of a sample in milliseconds.
	MaxSampleAgeMs int32 `json:"maxSampleAgeMs" yaml:"maxSampleAgeMs"`
}
