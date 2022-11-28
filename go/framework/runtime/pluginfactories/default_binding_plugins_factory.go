package pluginfactories

import (
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/pipeline"
)

var (
	_ pipeline.BindingPluginsFactory = (*DefaultBindingPluginsFactory)(nil)
)

type DefaultBindingPluginsFactory struct {
	pluginsRegistry *pipeline.PluginsRegistry[pipeline.ClusterAgentServices]
}

func NewDefaultBindingPluginsFactory(registry *pipeline.PluginsRegistry[pipeline.ClusterAgentServices]) *DefaultBindingPluginsFactory {
	factory := &DefaultBindingPluginsFactory{
		pluginsRegistry: registry,
	}
	return factory
}

func (f *DefaultBindingPluginsFactory) NewBindingPipelinePlugins(clusterAgentServices pipeline.ClusterAgentServices) (*pipeline.BindingPipelinePlugins, error) {
	clusterAgentConfig := clusterAgentServices.Config()
	pluginsConfig := clusterAgentConfig.PluginsConfig
	pluginsList := clusterAgentConfig.BindingPlugins

	bindingPipelinePlugins := pipeline.BindingPipelinePlugins{}
	pluginInstances := make(map[string]pipeline.Plugin, 0)

	if plugins, err := setUpPlugins[pipeline.CheckConflictsPlugin](pluginsList.CheckConflicts, f.pluginsRegistry, clusterAgentServices, pluginInstances, pipeline.CheckConflictsStage, pluginsConfig); err == nil {
		bindingPipelinePlugins.CheckConflicts = plugins
	} else {
		return nil, err
	}

	return &bindingPipelinePlugins, nil
}
