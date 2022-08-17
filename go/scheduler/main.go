package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"polaris-slo-cloud.github.io/polaris-scheduler/v2/scheduler/cmd"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	schedulerCmd := cmd.NewPolarisSchedulerCmd()
	if err := schedulerCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
