package kubeutil

import (
	"fmt"

	hashstructure "github.com/mitchellh/hashstructure/v2"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Computes the hash of the objSpec and stores it in an annotation of obj.
//
// We store this hash to avoid problems with deep equality comparisons on fields, which are assigned
// a default value by Kubernetes and would thus always result in being unequal to our generated specs.
// See https://github.com/kubernetes-sigs/kubebuilder/issues/592#issuecomment-484474923
//
// If the hash results in collisions for common cases, an alternative would be
// https://github.com/banzaicloud/k8s-objectmatcher
func SetSpecHash(obj meta.Object, objSpec interface{}) {
	hash, err := hashstructure.Hash(objSpec, hashstructure.FormatV2, nil)
	if err == nil {
		hashStr := fmt.Sprintf("%v", hash)
		SetAnnotation(obj, AnnotationSpecHash, hashStr)
	}
}

// Returns true if the spec hash annotation values of the two objects are equal
// or false, if they are not equal or both are not set.
func CheckSpecHashesAreEqual(a, b meta.Object) bool {
	hashA, foundA := GetAnnotation(a, AnnotationSpecHash)
	hashB, foundB := GetAnnotation(b, AnnotationSpecHash)

	// If both hashes are set, we compare them.
	if foundA && foundB {
		return hashA == hashB
	}

	// If both are not set or only one of them is set, they are not equal.
	return false
}
