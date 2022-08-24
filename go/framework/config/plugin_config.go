package config

// The base interface for a plugin configuration.
type PluginConfig map[string]interface{}

// A single entry in the PluginsList.
type PluginListEntry struct {

	// The name of the Plugin.
	Name string `yaml:"name"`

	// The weight of the plugin (applies only to Score plugins).
	// Default: 1
	Weight int32 `yaml:"weight"`
}

// Used to configure the plugins used in the scheduling pipeline.
type PluginsList struct {
	Sort *PluginListEntry `yaml:"sort"`

	SampleNodes *PluginListEntry `yaml:"sampleNodes"`

	PreFilter []*PluginListEntry `yaml:"preFilter"`

	Filter []*PluginListEntry `yaml:"filter"`

	PreScore []*PluginListEntry `yaml:"preScore"`

	Score []*PluginListEntry `yaml:"score"`

	Reserve []*PluginListEntry `yaml:"reserve"`
}

// Stores configuration for a specific plugin in the PluginsList.
type PluginsConfigListEntry struct {

	// The name of the Plugin.
	Name string `yaml:"name"`

	// Configuration data for the plugin.
	Config PluginConfig `yaml:"config"`
}
