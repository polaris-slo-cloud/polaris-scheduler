package configmanager

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
)

var (
	_configMgrImpl *configManagerImpl

	_ ConfigManager = _configMgrImpl
)

type configManagerImpl struct {
	restConfig *rest.Config
	scheme     *runtime.Scheme
}

func newConfigManagerImpl(restConfig *rest.Config, scheme *runtime.Scheme) *configManagerImpl {
	return &configManagerImpl{
		restConfig: restConfig,
		scheme:     scheme,
	}
}

func (me *configManagerImpl) RestConfig() *rest.Config {
	return me.restConfig
}

func (me *configManagerImpl) Scheme() *runtime.Scheme {
	return me.scheme
}
