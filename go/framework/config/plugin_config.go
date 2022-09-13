package config

import "fmt"

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

// Reads a string value from the raw plugin config.
func ReadStringFromPluginConfig(rawConfig PluginConfig, key string) (string, error) {
	if val, ok := rawConfig[key]; ok {
		if valStr, ok := val.(string); ok {
			return valStr, nil
		} else {
			return "", fmt.Errorf("value is not of type string")
		}
	} else {
		return "", fmt.Errorf("the key %s does not exist in the config", key)
	}
}

// Reads an int32 value from the raw plugin config.
func ReadInt32FromPluginConfig(rawConfig PluginConfig, key string) (int32, error) {
	if val, ok := rawConfig[key]; ok {
		if valInt32, ok := val.(int32); ok {
			return valInt32, nil
		} else {
			return 0, fmt.Errorf("value is not of type int32")
		}
	} else {
		return 0, fmt.Errorf("the key %s does not exist in the config", key)
	}
}

// Reads a nested object from the raw plugin config.
func ReadNestedObjectFromPluginConfig(rawConfig PluginConfig, key string) (PluginConfig, error) {
	if val, ok := rawConfig[key]; ok {
		if valObj, ok := val.(PluginConfig); ok {
			return valObj, nil
		} else {
			return nil, fmt.Errorf("value is not of type nested PluginConfig")
		}
	} else {
		return nil, fmt.Errorf("the key %s does not exist in the config", key)
	}
}

// Reads a string map from the raw plugin config.
func ReadStringMapFromPluginConfig(rawConfig PluginConfig, key string) (map[string]string, error) {
	if val, ok := rawConfig[key]; ok {
		if valObj, ok := val.(map[string]string); ok {
			return valObj, nil
		} else {
			return nil, fmt.Errorf("value is not of type map[string]string")
		}
	} else {
		return nil, fmt.Errorf("the key %s does not exist in the config", key)
	}
}
