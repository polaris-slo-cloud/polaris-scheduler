package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"polaris-slo-cloud.github.io/polaris-scheduler/v2/contextawareness/plugins/batterylevel"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/contextawareness/plugins/geolocation"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/pipeline"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/plugins/prioritysort"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/plugins/remotesampler"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/plugins/resourcesfit"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/util"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/scheduler/cmd"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	ctx := util.SetupSignalHandlingContext()

	pluginsRegistry := pipeline.NewPluginsRegistry(map[string]pipeline.PluginFactoryFunc[pipeline.PolarisScheduler]{
		prioritysort.PluginName:  prioritysort.NewPrioritySortPlugin,
		remotesampler.PluginName: remotesampler.NewRemoteNodesSamplerPlugin,
		resourcesfit.PluginName:  resourcesfit.NewResourcesFitSchedulingPlugin,
		geolocation.PluginName:   resourcesfit.NewResourcesFitSchedulingPlugin,
		batterylevel.PluginName:  batterylevel.NewBatteryLevelSchedulingPlugin,
	})

	schedulerCmd := cmd.NewPolarisSchedulerCmd(ctx, pluginsRegistry)
	if err := schedulerCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	<-ctx.Done()
}
