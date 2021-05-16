/*
Copyright 2021 Rainbow Project.

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

package fogapps

import (
	"context"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	fogapps "k8s.rainbow-h2020.eu/rainbow/apis/fogapps/v1"
)

// ServiceGraphReconciler reconciles a ServiceGraph object
type ServiceGraphReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// Permissions on ServiceGraphs:
//+kubebuilder:rbac:groups=fogapps.k8s.rainbow-h2020.eu,resources=servicegraphs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=fogapps.k8s.rainbow-h2020.eu,resources=servicegraphs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=fogapps.k8s.rainbow-h2020.eu,resources=servicegraphs/finalizers,verbs=update

// Permissions on Deployments and StatefulSets:
//+kubebuilder:rbac:groups=apps/v1,resources=deployments;statefulsets,verbs=get;list;watch;create;update;patch;delete

// Reconcile is triggered whenever a ServiceGraph is added, changed, or removed.
//
// Reconcile applies changes to the deployments in the cluster to ensure that they reflect the new state of the ServiceGraph object.
func (me *ServiceGraphReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := me.Log.WithValues("Reconcile ServiceGraph", req.NamespacedName)

	var serviceGraph fogapps.ServiceGraph
	if err := me.Get(ctx, req.NamespacedName, &serviceGraph); err != nil {
		log.Error(err, "Unable to fetch ServiceGraph")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// ToDo

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ServiceGraphReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&fogapps.ServiceGraph{}).
		Complete(r)
}
