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
	"fmt"

	"github.com/go-logr/logr"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	fogapps "k8s.rainbow-h2020.eu/rainbow/orchestration/apis/fogapps/v1"
)

var (
	ownerKey         = ".metadata.controller"
	fogAppsGVString  = fogapps.GroupVersion.String()
	serviceGraphKind = "ServiceGraph"
)

// ServiceGraphReconciler reconciles a ServiceGraph object
type ServiceGraphReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// serviceGraphChildObjects collects all child objects that are created from a ServiceGraph.
type serviceGraphChildObjects struct {
	Deployments  []apps.Deployment
	StatefulSets []apps.StatefulSet
	Services     []core.Service
	Ingresses    []networking.Ingress
}

// Permissions on ServiceGraphs:
//+kubebuilder:rbac:groups=fogapps.k8s.rainbow-h2020.eu,resources=servicegraphs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=fogapps.k8s.rainbow-h2020.eu,resources=servicegraphs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=fogapps.k8s.rainbow-h2020.eu,resources=servicegraphs/finalizers,verbs=update

// Permissions on Deployments and StatefulSets:
//+kubebuilder:rbac:groups=apps/v1,resources=deployments;statefulsets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps/v1,resources=deployments/status;statefulsets/status,verbs=get

// Permissions on Services:
//+kubebuilder:rbac:groups=v1,resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=v1,resources=services/status,verbs=get

// Permissions on Ingresses:
//+kubebuilder:rbac:groups=networking.k8s.io/v1,resources=ingresses,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.k8s.io/v1,resources=ingresses/status,verbs=get

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

	children, err := me.fetchChildObjects(ctx, req)
	if err != nil {
		return ctrl.Result{}, err
	}
	log.Info("Successfully fetched all child objects.")

	changes, err := ProcessServiceGraph(
		&serviceGraph,
		children,
		log,
		func(ownedObj meta.Object) error {
			return ctrl.SetControllerReference(&serviceGraph, ownedObj, me.Scheme)
		},
	)
	if err != nil {
		return ctrl.Result{}, err
	}

	log.Info("Applying changes.", "count", changes.Size())
	if err := changes.Apply(ctx, me.Client); err != nil {
		return ctrl.Result{}, err
	}
	log.Info("Successfully applied all changes.")

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (me *ServiceGraphReconciler) SetupWithManager(mgr ctrl.Manager) error {
	var indexerFn client.IndexerFunc = func(rawObj client.Object) []string {
		owner := meta.GetControllerOf(rawObj)
		if owner == nil {
			return nil
		}

		if owner.APIVersion == fogAppsGVString && owner.Kind == serviceGraphKind {
			return []string{owner.Name}
		}
		return nil
	}

	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &apps.Deployment{}, ownerKey, indexerFn); err != nil {
		return err
	}
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &apps.StatefulSet{}, ownerKey, indexerFn); err != nil {
		return err
	}
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &core.Service{}, ownerKey, indexerFn); err != nil {
		return err
	}
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &networking.Ingress{}, ownerKey, indexerFn); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&fogapps.ServiceGraph{}).
		Owns(&apps.Deployment{}).
		Owns(&apps.StatefulSet{}).
		Complete(me)
}

// fetchChildObjects loads all objects that have been created from the respective ServiceGraph
func (me *ServiceGraphReconciler) fetchChildObjects(ctx context.Context, req ctrl.Request) (*serviceGraphChildObjects, error) {
	children := serviceGraphChildObjects{}

	var deployments apps.DeploymentList
	if err := me.List(ctx, &deployments, client.InNamespace(req.Namespace), client.MatchingFields{ownerKey: req.Name}); err != nil {
		return nil, fmt.Errorf("unable to load child Deployments. Cause: %w", err)
	}
	children.Deployments = deployments.Items

	var statefulSets apps.StatefulSetList
	if err := me.List(ctx, &statefulSets, client.InNamespace(req.Namespace), client.MatchingFields{ownerKey: req.Name}); err != nil {
		return nil, fmt.Errorf("unable to load child StatefulSets. Cause: %w", err)
	}
	children.StatefulSets = statefulSets.Items

	var services core.ServiceList
	if err := me.List(ctx, &services, client.InNamespace(req.Namespace), client.MatchingFields{ownerKey: req.Name}); err != nil {
		return nil, fmt.Errorf("unable to load child Services. Cause: %w", err)
	}
	children.Services = services.Items

	var ingresses networking.IngressList
	if err := me.List(ctx, &ingresses, client.InNamespace(req.Namespace), client.MatchingFields{ownerKey: req.Name}); err != nil {
		return nil, fmt.Errorf("unable to load child Ingresses. Cause: %w", err)
	}
	children.Ingresses = ingresses.Items

	return &children, nil
}
