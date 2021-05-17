package controllerutil

import (
	"context"
	"fmt"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientPkg "sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	_changesList *ResourceChangesList
	_addition    *ResourceAddition
	_update      *ResourceUpdate
	_deletion    *ResourceDeletion
	_            ResourceChange = _changesList
	_            ResourceChange = _addition
	_            ResourceChange = _update
	_            ResourceChange = _deletion
)

// ResourceChange models a single resource change or a set of changes.
type ResourceChange interface {
	// Apply applies the resource change to the cluster.
	Apply(ctx context.Context, client clientPkg.Client) error
}

// SingleResourceChange is the supertype for all ResourceChanges that affect a single resource.
type SingleResourceChange struct {
	// The resource to be added, changed, or deleted.
	Resource clientPkg.Object
}

// ResourceAddition adds a new resource.
type ResourceAddition struct {
	SingleResourceChange
}

// NewResourceAddition creates a new ResourceAddition.
func NewResourceAddition(resource clientPkg.Object) *ResourceAddition {
	return &ResourceAddition{
		SingleResourceChange: SingleResourceChange{
			Resource: resource,
		},
	}
}

func (me *ResourceAddition) Apply(ctx context.Context, client clientPkg.Client) error {
	if err := client.Create(ctx, me.Resource); err != nil {
		return fmt.Errorf("unable to create resource %s. Cause: %w", me.Resource, err)
	}
	return nil
}

// ResourceUpdate adds a new resource.
type ResourceUpdate struct {
	SingleResourceChange
}

// NewResourceUpdate creates a new ResourceUpdate.
func NewResourceUpdate(resource clientPkg.Object) *ResourceUpdate {
	return &ResourceUpdate{
		SingleResourceChange: SingleResourceChange{
			Resource: resource,
		},
	}
}

func (me *ResourceUpdate) Apply(ctx context.Context, client clientPkg.Client) error {
	if err := client.Update(ctx, me.Resource); err != nil {
		return fmt.Errorf("unable to update resource %s. Cause: %w", me.Resource, err)
	}
	return nil
}

// ResourceDeletion adds a new resource.
type ResourceDeletion struct {
	SingleResourceChange
}

// NewResourceDeletion creates a new ResourceDeletion.
func NewResourceDeletion(resource clientPkg.Object) *ResourceDeletion {
	return &ResourceDeletion{
		SingleResourceChange: SingleResourceChange{
			Resource: resource,
		},
	}
}

func (me *ResourceDeletion) Apply(ctx context.Context, client clientPkg.Client) error {
	if err := client.Delete(ctx, me.Resource, clientPkg.PropagationPolicy(meta.DeletePropagationBackground)); err != nil {
		return fmt.Errorf("unable to delete resource %s. Cause: %w", me.Resource, err)
	}
	return nil
}

// ResourceChangesList collects the changes that need to be made by a controller to
type ResourceChangesList struct {
	// The changes to be applied sequentially.
	Changes []ResourceChange
}

// NewResourceChangesList creates a new, empty ResourceChangesList.
func NewResourceChangesList() *ResourceChangesList {
	return &ResourceChangesList{
		Changes: make([]ResourceChange, 0, 10),
	}
}

func (me *ResourceChangesList) Apply(ctx context.Context, client clientPkg.Client) error {
	for _, change := range me.Changes {
		if err := change.Apply(ctx, client); err != nil {
			return err
		}
	}
	return nil
}

// Size returns the number of changes in the list.
func (me *ResourceChangesList) Size() int {
	return len(me.Changes)
}

// AddChange adds the specified change to the list of changes.
func (me *ResourceChangesList) AddChange(change ResourceChange) {
	me.Changes = append(me.Changes, change)
}
