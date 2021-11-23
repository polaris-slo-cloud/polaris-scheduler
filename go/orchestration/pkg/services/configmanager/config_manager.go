package configmanager

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
)

var (
	configMgrInstance ConfigManager
)

// ConfigManager provides easy access to the controller's configuration.
type ConfigManager interface {

	// Gets the Config object that can be used to create a REST client.
	RestConfig() *rest.Config

	// Gets the Scheme with all known CRDs for transforming from/to Go objects.
	Scheme() *runtime.Scheme
}

// Gets the ConfigManager's singleton instance or panics, if it hasn't been initialized.
func GetConfigManager() ConfigManager {
	if configMgrInstance == nil {
		panic("ConfigManager has not been initialized. Did you call configmanager.InitConfigManager()?")
	}
	return configMgrInstance
}

// Initializes the ConfigManager with the specified configuration and scheme.
func InitConfigManager(restConfig *rest.Config, scheme *runtime.Scheme) ConfigManager {
	configMgrInstance = newConfigManagerImpl(restConfig, scheme)
	return configMgrInstance
}
