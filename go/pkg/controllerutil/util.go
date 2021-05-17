package controllerutil

import (
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SetOwnerReferenceFn is a callback used to set the owner of a resource object.
//
// The callback usuall needs to be provided by a controller and is used by a factory
// for the respective type of owned objects.
type SetOwnerReferenceFn func(ownedObj meta.Object) error
