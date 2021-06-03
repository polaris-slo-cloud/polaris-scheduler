package kubeutil

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetLabel returns the label with the specified key.
func GetLabel(obj *metav1.ObjectMeta, key string) (string, bool) {
	if obj.Labels != nil {
		value, exists := obj.Labels[key]
		return value, exists
	}
	return "", false
}

// GetNamespace returns the namespace of the object, or "default" if the namespace is not set.
func GetNamespace(obj *metav1.ObjectMeta) string {
	if obj.Namespace != "" {
		return obj.Namespace
	}
	return "default"
}
