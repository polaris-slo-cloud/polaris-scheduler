package config

// The base interface for a plugin configuration.
type PluginConfig interface{}

// A single entry in the PluginsList.
type PluginListEntry struct {

	// The name of the Plugin.
	Name string

	// The weight of the plugin (applies only to Score plugins).
	// Default: 1
	Weight uint32
}

// Used to configure the plugins used in the scheduling pipeline.
type PluginsList struct {
	Sort *PluginListEntry

	SampleNodes *PluginListEntry

	PreFilter []*PluginListEntry

	Filter []*PluginListEntry

	PreScore []*PluginListEntry

	Score []*PluginListEntry

	Reserve []*PluginListEntry
}

// Stores configuration for a specific plugin in the PluginsList.
type PluginsConfigListEntry struct {

	// The name of the Plugin.
	Name string

	// Configuration data for the plugin.
	config PluginConfig
}
