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

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/component-base/logs"
	"k8s.io/kubernetes/cmd/kube-scheduler/app"
	ctrl "sigs.k8s.io/controller-runtime"

	clusterv1 "k8s.rainbow-h2020.eu/rainbow/orchestration/apis/cluster/v1"
	fogappsv1 "k8s.rainbow-h2020.eu/rainbow/orchestration/apis/fogapps/v1"
	slov1 "k8s.rainbow-h2020.eu/rainbow/orchestration/apis/slo/v1"
	"k8s.rainbow-h2020.eu/rainbow/orchestration/pkg/services/configmanager"
	"k8s.rainbow-h2020.eu/rainbow/orchestration/pkg/services/regionmanager"
	"k8s.rainbow-h2020.eu/rainbow/orchestration/pkg/services/servicegraphmanager"

	"k8s.rainbow-h2020.eu/rainbow/scheduler/pkg/schedulerplugins/atomicdeployment"
	"k8s.rainbow-h2020.eu/rainbow/scheduler/pkg/schedulerplugins/latency"
	"k8s.rainbow-h2020.eu/rainbow/scheduler/pkg/schedulerplugins/nodecost"
	"k8s.rainbow-h2020.eu/rainbow/scheduler/pkg/schedulerplugins/podspernode"
	"k8s.rainbow-h2020.eu/rainbow/scheduler/pkg/schedulerplugins/prioritymqsort"
	"k8s.rainbow-h2020.eu/rainbow/scheduler/pkg/schedulerplugins/reserve"
	"k8s.rainbow-h2020.eu/rainbow/scheduler/pkg/schedulerplugins/servicegraph"
)

var (
	scheme = runtime.NewScheme()
)

func initScheme() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(clusterv1.AddToScheme(scheme))
	utilruntime.Must(fogappsv1.AddToScheme(scheme))
	utilruntime.Must(slov1.AddToScheme(scheme))
}

func main() {
	rand.Seed(time.Now().UnixNano())
	initScheme()

	// Initialize the ConfigManager, RegionManager, and ServiceGraphManager.
	configmanager.InitConfigManager(ctrl.GetConfigOrDie(), scheme)
	regionmanager.InitRegionManager()
	servicegraphmanager.InitServiceGraphManager()

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
