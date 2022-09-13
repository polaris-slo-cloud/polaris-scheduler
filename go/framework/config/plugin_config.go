package config

// The base interface for a plugin configuration.
type PluginConfig map[string]interface{}

// A single entry in the PluginsList.
type PluginListEntry struct {

	// The name of the Plugin.
	Name string `json:"name" yaml:"name"`

	// The weight of the plugin (applies only to Score plugins).
	// Default: 1
	Weight int32 `json:"weight" yaml:"weight"`
}

// Used to configure the plugins used in the scheduling pipeline.
type PluginsList struct {
	Sort *PluginListEntry `json:"sort" yaml:"sort"`

	SampleNodes *PluginListEntry `json:"sampleNodes" yaml:"sampleNodes"`

	PreFilter []*PluginListEntry `json:"preFilter" yaml:"preFilter"`

	Filter []*PluginListEntry `json:"filter" yaml:"filter"`

	PreScore []*PluginListEntry `json:"preScore" yaml:"preScore"`

	Score []*PluginListEntry `json:"score" yaml:"score"`

	Reserve []*PluginListEntry `json:"reserve" yaml:"reserve"`
}

// Stores configuration for a specific plugin in the PluginsList.
type PluginsConfigListEntry struct {

	// The name of the Plugin.
	Name string `json:"name" yaml:"name"`

	// Configuration data for the plugin.
	Config PluginConfig `json:"config" yaml:"config"`
}
