package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/util"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/node-sampler/cmd"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/node-sampler/sampling"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	ctx := util.SetupSignalHandlingContext()

	samplingStrategies := []sampling.SamplingStrategyFactoryFunc{
		sampling.NewRandomSamplingStrategy,
		sampling.NewRoundRobinSamplingStrategy,
	}

	nodeSamplerCmd := cmd.NewPolarisNodeSamplerCmd(ctx, samplingStrategies)
	if err := nodeSamplerCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	<-ctx.Done()
}
