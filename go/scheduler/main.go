package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/pipeline"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/scheduler/cmd"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/scheduler/plugins/prioritysort"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/scheduler/plugins/randomsampler"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	ctx := setupSignalHandlingContext()

	pluginsRegistry := pipeline.NewPluginsRegistry(map[string]pipeline.PluginFactoryFunc{
		prioritysort.PluginName:  prioritysort.NewPrioritySortPlugin,
		randomsampler.PluginName: randomsampler.NewRandomNodesSamplerPlugin,
	})

	schedulerCmd := cmd.NewPolarisSchedulerCmd(ctx, pluginsRegistry)
	if err := schedulerCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	<-ctx.Done()
}

func setupSignalHandlingContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())

	signalHandler := make(chan os.Signal, 2)
	signal.Notify(signalHandler, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-signalHandler
		cancel()
		<-signalHandler
		os.Exit(1) // Exit immediately if we receive a second signal.
	}()

	return ctx
}
