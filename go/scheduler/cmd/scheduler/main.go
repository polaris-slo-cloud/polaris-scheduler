/*
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"math/rand"
	"os"
	"time"

	"k8s.io/component-base/logs"
	"k8s.io/kubernetes/cmd/kube-scheduler/app"

	"rainbow-h2020.eu/gomod/rainbow-scheduler/pkg/schedulerplugins/atomicdeployment"
	"rainbow-h2020.eu/gomod/rainbow-scheduler/pkg/schedulerplugins/latency"
	"rainbow-h2020.eu/gomod/rainbow-scheduler/pkg/schedulerplugins/nodecost"
	"rainbow-h2020.eu/gomod/rainbow-scheduler/pkg/schedulerplugins/podspernode"
	"rainbow-h2020.eu/gomod/rainbow-scheduler/pkg/schedulerplugins/prioritymqsort"
	"rainbow-h2020.eu/gomod/rainbow-scheduler/pkg/schedulerplugins/reserve"
	"rainbow-h2020.eu/gomod/rainbow-scheduler/pkg/schedulerplugins/servicegraph"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	// When executed, the command returned by NewSchedulerCommand(), uses
	// scheduler.WithFrameworkOutOfTreeRegistry(outOfTreeRegistry) to append the specified plugins to
	// the default plugins (see Kubernetes source: cmd/kube-scheduler/app/server.go).
	command := app.NewSchedulerCommand(
		app.WithPlugin(prioritymqsort.PluginName, prioritymqsort.New),
		app.WithPlugin(servicegraph.PluginName, servicegraph.New),
		app.WithPlugin(latency.PluginName, latency.New),
		app.WithPlugin(podspernode.PluginName, podspernode.New),
		app.WithPlugin(nodecost.PluginName, nodecost.New),
		app.WithPlugin(reserve.PluginName, reserve.New),
		app.WithPlugin(atomicdeployment.PluginName, atomicdeployment.New),
	)

	logs.InitLogs()
	defer logs.FlushLogs()

	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}
