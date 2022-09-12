package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/pipeline"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/plugins/prioritysort"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/plugins/resourcesfit"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/util"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/scheduler/cmd"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/scheduler/plugins/randomsampler"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	ctx := util.SetupSignalHandlingContext()

	pluginsRegistry := pipeline.NewPluginsRegistry(map[string]pipeline.PluginFactoryFunc{
		prioritysort.PluginName:  prioritysort.NewPrioritySortPlugin,
		randomsampler.PluginName: randomsampler.NewRandomNodesSamplerPlugin,
		resourcesfit.PluginName:  resourcesfit.NewResourcesFitPlugin,
	})

	schedulerCmd := cmd.NewPolarisSchedulerCmd(ctx, pluginsRegistry)
	if err := schedulerCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	<-ctx.Done()
}
