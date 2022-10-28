package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"polaris-slo-cloud.github.io/polaris-scheduler/v2/cluster-agent/cmd"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/contextawareness/plugins/geolocation"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/pipeline"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/plugins/randomsampling"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/plugins/resourcesfit"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/plugins/roundrobinsampling"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/util"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	ctx := util.SetupSignalHandlingContext()

	pluginsRegistry := pipeline.NewPluginsRegistry(map[string]pipeline.PluginFactoryFunc[pipeline.PolarisNodeSampler]{
		randomsampling.PluginName:     randomsampling.NewRandomSamplingStrategy,
		roundrobinsampling.PluginName: roundrobinsampling.NewRoundRobinSamplingStrategy,
		resourcesfit.PluginName:       resourcesfit.NewResourcesFitSamplingPlugin,
		geolocation.PluginName:        resourcesfit.NewResourcesFitSamplingPlugin,
	})

	nodeSamplerCmd := cmd.NewPolarisClusterAgentCmd(ctx, pluginsRegistry)
	if err := nodeSamplerCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	<-ctx.Done()
}
