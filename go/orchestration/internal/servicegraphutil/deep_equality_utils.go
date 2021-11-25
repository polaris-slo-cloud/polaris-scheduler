package servicegraphutil

import (
	"reflect"

	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
)

// Returns true if both Deployment specs are equal with respect to the fields set by the ServiceGraph controller.
// This only compares fields, which are of interest to us to avoid false inequalities caused by
// empty fields being set to default values.
func DeepEqualDeploymentSpecs(a, b *apps.DeploymentSpec) bool {
	equal := reflect.DeepEqual(a.Replicas, b.Replicas)
	equal = equal && reflect.DeepEqual(a.Selector, b.Selector)
	equal = equal && DeepEqualPodTemplate(&a.Template, &b.Template)
	return equal
}

// Returns true if both StatefulSet specs are equal with respect to the fields set by the ServiceGraph controller.
func DeepEqualStatefulSetSpecs(a, b *apps.StatefulSetSpec) bool {
	equal := reflect.DeepEqual(a.Replicas, b.Replicas)
	equal = equal && reflect.DeepEqual(a.Selector, b.Selector)
	equal = equal && DeepEqualPodTemplate(&a.Template, &b.Template)
	return equal
}

// Returns true if both PodTemplates are equal with respect to the fields set by the ServiceGraph controller.
func DeepEqualPodTemplate(a, b *core.PodTemplateSpec) bool {
	equal := reflect.DeepEqual(a.ObjectMeta.Labels, b.ObjectMeta.Labels)
	equal = equal && reflect.DeepEqual(a.Spec.Affinity, b.Spec.Affinity)
	equal = equal && reflect.DeepEqual(a.Spec.DNSConfig, b.Spec.DNSConfig)
	equal = equal && reflect.DeepEqual(a.Spec.DNSPolicy, b.Spec.DNSPolicy)
	equal = equal && reflect.DeepEqual(a.Spec.HostAliases, b.Spec.HostAliases)
	equal = equal && a.Spec.HostIPC == b.Spec.HostIPC
	equal = equal && a.Spec.HostNetwork == b.Spec.HostNetwork
	equal = equal && a.Spec.HostPID == b.Spec.HostPID
	equal = equal && a.Spec.Hostname == b.Spec.Hostname
	equal = equal && reflect.DeepEqual(a.Spec.ImagePullSecrets, b.Spec.ImagePullSecrets)
	equal = equal && a.Spec.SchedulerName == b.Spec.SchedulerName
	equal = equal && reflect.DeepEqual(a.Spec.Volumes, b.Spec.Volumes)
	equal = equal && reflect.DeepEqual(a.Spec.NodeSelector, b.Spec.NodeSelector)
	equal = equal && a.Spec.ServiceAccountName == b.Spec.ServiceAccountName
	equal = equal && DeepEqualContainerArrays(a.Spec.InitContainers, b.Spec.InitContainers)
	equal = equal && DeepEqualContainerArrays(a.Spec.Containers, b.Spec.Containers)
	return equal
}

// Returns true if both Container arrays are equal with respect to the fields set by the ServiceGraph controller.
func DeepEqualContainerArrays(a, b []core.Container) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if !DeepEqualContainers(&a[i], &b[i]) {
			return false
		}
	}
	return true
}

// Returns true if both Containers are equal with respect to the fields set by the ServiceGraph controller.
func DeepEqualContainers(a, b *core.Container) bool {
	equal := a.Name == b.Name
	equal = equal && a.Image == b.Image
	equal = equal && reflect.DeepEqual(a.Args, b.Args)
	equal = equal && reflect.DeepEqual(a.Command, b.Command)
	equal = equal && reflect.DeepEqual(a.Env, b.Env)
	equal = equal && reflect.DeepEqual(a.EnvFrom, b.EnvFrom)
	equal = equal && reflect.DeepEqual(a.VolumeDevices, b.VolumeDevices)
	equal = equal && reflect.DeepEqual(a.VolumeMounts, b.VolumeMounts)
	equal = equal && DeepEqualResourceLists(a.Resources.Limits, b.Resources.Limits)
	equal = equal && DeepEqualResourceLists(a.Resources.Requests, b.Resources.Requests)
	return equal
}

// Returns true if both ResourceLists are equal.
func DeepEqualResourceLists(a, b core.ResourceList) bool {
	if len(a) != len(b) {
		return false
	}
	for key, valueA := range a {
		valueB, found := b[key]
		if !found || !valueA.Equal(valueB) {
			return false
		}
	}
	return true
}
