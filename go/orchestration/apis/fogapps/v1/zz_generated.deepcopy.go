// +build !ignore_autogenerated

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

// Code generated by controller-gen. DO NOT EDIT.

package v1

import (
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clusterv1 "k8s.rainbow-h2020.eu/rainbow/orchestration/apis/cluster/v1"
	slov1 "k8s.rainbow-h2020.eu/rainbow/orchestration/apis/slo/v1"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ApiVersionKind) DeepCopyInto(out *ApiVersionKind) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ApiVersionKind.
func (in *ApiVersionKind) DeepCopy() *ApiVersionKind {
	if in == nil {
		return nil
	}
	out := new(ApiVersionKind)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ArbitraryObject) DeepCopyInto(out *ArbitraryObject) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ArbitraryObject.
func (in *ArbitraryObject) DeepCopy() *ArbitraryObject {
	if in == nil {
		return nil
	}
	out := new(ArbitraryObject)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ConfigMap) DeepCopyInto(out *ConfigMap) {
	*out = *in
	if in.Data != nil {
		in, out := &in.Data, &out.Data
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.BinaryData != nil {
		in, out := &in.BinaryData, &out.BinaryData
		*out = make(map[string][]byte, len(*in))
		for key, val := range *in {
			var outVal []byte
			if val == nil {
				(*out)[key] = nil
			} else {
				in, out := &val, &outVal
				*out = make([]byte, len(*in))
				copy(*out, *in)
			}
			(*out)[key] = outVal
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ConfigMap.
func (in *ConfigMap) DeepCopy() *ConfigMap {
	if in == nil {
		return nil
	}
	out := new(ConfigMap)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CpuInfo) DeepCopyInto(out *CpuInfo) {
	*out = *in
	if in.Architectures != nil {
		in, out := &in.Architectures, &out.Architectures
		*out = make([]CpuArchitecture, len(*in))
		copy(*out, *in)
	}
	if in.MinCores != nil {
		in, out := &in.MinCores, &out.MinCores
		*out = new(int32)
		**out = **in
	}
	if in.MinBashClockMHz != nil {
		in, out := &in.MinBashClockMHz, &out.MinBashClockMHz
		*out = new(int32)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CpuInfo.
func (in *CpuInfo) DeepCopy() *CpuInfo {
	if in == nil {
		return nil
	}
	out := new(CpuInfo)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DNSConfig) DeepCopyInto(out *DNSConfig) {
	*out = *in
	in.PodDNSConfig.DeepCopyInto(&out.PodDNSConfig)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DNSConfig.
func (in *DNSConfig) DeepCopy() *DNSConfig {
	if in == nil {
		return nil
	}
	out := new(DNSConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ExposedPorts) DeepCopyInto(out *ExposedPorts) {
	*out = *in
	if in.Ports != nil {
		in, out := &in.Ports, &out.Ports
		*out = make([]corev1.ServicePort, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ExposedPorts.
func (in *ExposedPorts) DeepCopy() *ExposedPorts {
	if in == nil {
		return nil
	}
	out := new(ExposedPorts)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GeoLocation) DeepCopyInto(out *GeoLocation) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GeoLocation.
func (in *GeoLocation) DeepCopy() *GeoLocation {
	if in == nil {
		return nil
	}
	out := new(GeoLocation)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GpuInfo) DeepCopyInto(out *GpuInfo) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GpuInfo.
func (in *GpuInfo) DeepCopy() *GpuInfo {
	if in == nil {
		return nil
	}
	out := new(GpuInfo)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LinkQosRequirements) DeepCopyInto(out *LinkQosRequirements) {
	*out = *in
	if in.LinkType != nil {
		in, out := &in.LinkType, &out.LinkType
		*out = new(LinkType)
		(*in).DeepCopyInto(*out)
	}
	if in.Throughput != nil {
		in, out := &in.Throughput, &out.Throughput
		*out = new(NetworkThroughputRequirements)
		(*in).DeepCopyInto(*out)
	}
	if in.Latency != nil {
		in, out := &in.Latency, &out.Latency
		*out = new(NetworkLatencyRequirements)
		(*in).DeepCopyInto(*out)
	}
	if in.PacketLoss != nil {
		in, out := &in.PacketLoss, &out.PacketLoss
		*out = new(NetworkPacketLossRequirements)
		**out = **in
	}
	if in.ElasticityStrategy != nil {
		in, out := &in.ElasticityStrategy, &out.ElasticityStrategy
		*out = new(NetworkElasticityStrategyConfig)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LinkQosRequirements.
func (in *LinkQosRequirements) DeepCopy() *LinkQosRequirements {
	if in == nil {
		return nil
	}
	out := new(LinkQosRequirements)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LinkTrustRequirements) DeepCopyInto(out *LinkTrustRequirements) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LinkTrustRequirements.
func (in *LinkTrustRequirements) DeepCopy() *LinkTrustRequirements {
	if in == nil {
		return nil
	}
	out := new(LinkTrustRequirements)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LinkType) DeepCopyInto(out *LinkType) {
	*out = *in
	if in.Protocol != nil {
		in, out := &in.Protocol, &out.Protocol
		*out = new(LinkProtocol)
		**out = **in
	}
	if in.MinQualityClass != nil {
		in, out := &in.MinQualityClass, &out.MinQualityClass
		*out = new(clusterv1.NetworkQualityClass)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LinkType.
func (in *LinkType) DeepCopy() *LinkType {
	if in == nil {
		return nil
	}
	out := new(LinkType)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MonitoringConfig) DeepCopyInto(out *MonitoringConfig) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MonitoringConfig.
func (in *MonitoringConfig) DeepCopy() *MonitoringConfig {
	if in == nil {
		return nil
	}
	out := new(MonitoringConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NetworkElasticityStrategyConfig) DeepCopyInto(out *NetworkElasticityStrategyConfig) {
	*out = *in
	out.ElasticityStrategy = in.ElasticityStrategy
	if in.StabilizationWindow != nil {
		in, out := &in.StabilizationWindow, &out.StabilizationWindow
		*out = new(slov1.StabilizationWindow)
		(*in).DeepCopyInto(*out)
	}
	if in.StaticElasticityStrategyConfig != nil {
		in, out := &in.StaticElasticityStrategyConfig, &out.StaticElasticityStrategyConfig
		*out = new(runtime.RawExtension)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NetworkElasticityStrategyConfig.
func (in *NetworkElasticityStrategyConfig) DeepCopy() *NetworkElasticityStrategyConfig {
	if in == nil {
		return nil
	}
	out := new(NetworkElasticityStrategyConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NetworkLatencyRequirements) DeepCopyInto(out *NetworkLatencyRequirements) {
	*out = *in
	if in.MaxPacketDelayVariance != nil {
		in, out := &in.MaxPacketDelayVariance, &out.MaxPacketDelayVariance
		*out = new(int32)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NetworkLatencyRequirements.
func (in *NetworkLatencyRequirements) DeepCopy() *NetworkLatencyRequirements {
	if in == nil {
		return nil
	}
	out := new(NetworkLatencyRequirements)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NetworkPacketLossRequirements) DeepCopyInto(out *NetworkPacketLossRequirements) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NetworkPacketLossRequirements.
func (in *NetworkPacketLossRequirements) DeepCopy() *NetworkPacketLossRequirements {
	if in == nil {
		return nil
	}
	out := new(NetworkPacketLossRequirements)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NetworkThroughputRequirements) DeepCopyInto(out *NetworkThroughputRequirements) {
	*out = *in
	if in.MaxBandwidthVariance != nil {
		in, out := &in.MaxBandwidthVariance, &out.MaxBandwidthVariance
		*out = new(int64)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NetworkThroughputRequirements.
func (in *NetworkThroughputRequirements) DeepCopy() *NetworkThroughputRequirements {
	if in == nil {
		return nil
	}
	out := new(NetworkThroughputRequirements)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NodeHardware) DeepCopyInto(out *NodeHardware) {
	*out = *in
	if in.NodeType != nil {
		in, out := &in.NodeType, &out.NodeType
		*out = new(string)
		**out = **in
	}
	if in.CpuInfo != nil {
		in, out := &in.CpuInfo, &out.CpuInfo
		*out = new(CpuInfo)
		(*in).DeepCopyInto(*out)
	}
	if in.GpuInfo != nil {
		in, out := &in.GpuInfo, &out.GpuInfo
		*out = new(GpuInfo)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NodeHardware.
func (in *NodeHardware) DeepCopy() *NodeHardware {
	if in == nil {
		return nil
	}
	out := new(NodeHardware)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NodeTrustRequirements) DeepCopyInto(out *NodeTrustRequirements) {
	*out = *in
	if in.MinTpmVersion != nil {
		in, out := &in.MinTpmVersion, &out.MinTpmVersion
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NodeTrustRequirements.
func (in *NodeTrustRequirements) DeepCopy() *NodeTrustRequirements {
	if in == nil {
		return nil
	}
	out := new(NodeTrustRequirements)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RainbowService) DeepCopyInto(out *RainbowService) {
	*out = *in
	out.Type = in.Type
	if in.Config != nil {
		in, out := &in.Config, &out.Config
		*out = new(ArbitraryObject)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RainbowService.
func (in *RainbowService) DeepCopy() *RainbowService {
	if in == nil {
		return nil
	}
	out := new(RainbowService)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ReplicasConfig) DeepCopyInto(out *ReplicasConfig) {
	*out = *in
	if in.InitialCount != nil {
		in, out := &in.InitialCount, &out.InitialCount
		*out = new(int32)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ReplicasConfig.
func (in *ReplicasConfig) DeepCopy() *ReplicasConfig {
	if in == nil {
		return nil
	}
	out := new(ReplicasConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceGraph) DeepCopyInto(out *ServiceGraph) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServiceGraph.
func (in *ServiceGraph) DeepCopy() *ServiceGraph {
	if in == nil {
		return nil
	}
	out := new(ServiceGraph)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ServiceGraph) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceGraphCondition) DeepCopyInto(out *ServiceGraphCondition) {
	*out = *in
	if in.Message != nil {
		in, out := &in.Message, &out.Message
		*out = new(string)
		**out = **in
	}
	in.LastTransitionTime.DeepCopyInto(&out.LastTransitionTime)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServiceGraphCondition.
func (in *ServiceGraphCondition) DeepCopy() *ServiceGraphCondition {
	if in == nil {
		return nil
	}
	out := new(ServiceGraphCondition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceGraphList) DeepCopyInto(out *ServiceGraphList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ServiceGraph, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServiceGraphList.
func (in *ServiceGraphList) DeepCopy() *ServiceGraphList {
	if in == nil {
		return nil
	}
	out := new(ServiceGraphList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ServiceGraphList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceGraphNode) DeepCopyInto(out *ServiceGraphNode) {
	*out = *in
	if in.ServiceAccountName != nil {
		in, out := &in.ServiceAccountName, &out.ServiceAccountName
		*out = new(string)
		**out = **in
	}
	if in.PodLabels != nil {
		in, out := &in.PodLabels, &out.PodLabels
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.InitContainers != nil {
		in, out := &in.InitContainers, &out.InitContainers
		*out = make([]corev1.Container, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Containers != nil {
		in, out := &in.Containers, &out.Containers
		*out = make([]corev1.Container, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.ImagePullSecrets != nil {
		in, out := &in.ImagePullSecrets, &out.ImagePullSecrets
		*out = make([]corev1.LocalObjectReference, len(*in))
		copy(*out, *in)
	}
	if in.Volumes != nil {
		in, out := &in.Volumes, &out.Volumes
		*out = make([]corev1.Volume, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	in.Replicas.DeepCopyInto(&out.Replicas)
	if in.ExposedPorts != nil {
		in, out := &in.ExposedPorts, &out.ExposedPorts
		*out = new(ExposedPorts)
		(*in).DeepCopyInto(*out)
	}
	if in.Affinity != nil {
		in, out := &in.Affinity, &out.Affinity
		*out = new(corev1.Affinity)
		(*in).DeepCopyInto(*out)
	}
	if in.SLOs != nil {
		in, out := &in.SLOs, &out.SLOs
		*out = make([]ServiceLevelObjective, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.RainbowServices != nil {
		in, out := &in.RainbowServices, &out.RainbowServices
		*out = make([]RainbowService, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.TrustRequirements != nil {
		in, out := &in.TrustRequirements, &out.TrustRequirements
		*out = new(NodeTrustRequirements)
		(*in).DeepCopyInto(*out)
	}
	if in.NodeHardware != nil {
		in, out := &in.NodeHardware, &out.NodeHardware
		*out = new(NodeHardware)
		(*in).DeepCopyInto(*out)
	}
	if in.GeoLocation != nil {
		in, out := &in.GeoLocation, &out.GeoLocation
		*out = new(GeoLocation)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServiceGraphNode.
func (in *ServiceGraphNode) DeepCopy() *ServiceGraphNode {
	if in == nil {
		return nil
	}
	out := new(ServiceGraphNode)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceGraphNodeStatus) DeepCopyInto(out *ServiceGraphNodeStatus) {
	*out = *in
	if in.DeploymentResourceType != nil {
		in, out := &in.DeploymentResourceType, &out.DeploymentResourceType
		*out = new(metav1.GroupVersionKind)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServiceGraphNodeStatus.
func (in *ServiceGraphNodeStatus) DeepCopy() *ServiceGraphNodeStatus {
	if in == nil {
		return nil
	}
	out := new(ServiceGraphNodeStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceGraphSpec) DeepCopyInto(out *ServiceGraphSpec) {
	*out = *in
	if in.ServiceAccountName != nil {
		in, out := &in.ServiceAccountName, &out.ServiceAccountName
		*out = new(string)
		**out = **in
	}
	if in.Nodes != nil {
		in, out := &in.Nodes, &out.Nodes
		*out = make([]ServiceGraphNode, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Links != nil {
		in, out := &in.Links, &out.Links
		*out = make([]ServiceLink, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.SLOs != nil {
		in, out := &in.SLOs, &out.SLOs
		*out = make([]ServiceLevelObjective, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.RainbowServices != nil {
		in, out := &in.RainbowServices, &out.RainbowServices
		*out = make([]RainbowService, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.ConfigMaps != nil {
		in, out := &in.ConfigMaps, &out.ConfigMaps
		*out = make([]ConfigMap, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.DNSConfig != nil {
		in, out := &in.DNSConfig, &out.DNSConfig
		*out = new(DNSConfig)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServiceGraphSpec.
func (in *ServiceGraphSpec) DeepCopy() *ServiceGraphSpec {
	if in == nil {
		return nil
	}
	out := new(ServiceGraphSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceGraphStatus) DeepCopyInto(out *ServiceGraphStatus) {
	*out = *in
	if in.NodeStates != nil {
		in, out := &in.NodeStates, &out.NodeStates
		*out = make(map[string]*ServiceGraphNodeStatus, len(*in))
		for key, val := range *in {
			var outVal *ServiceGraphNodeStatus
			if val == nil {
				(*out)[key] = nil
			} else {
				in, out := &val, &outVal
				*out = new(ServiceGraphNodeStatus)
				(*in).DeepCopyInto(*out)
			}
			(*out)[key] = outVal
		}
	}
	if in.SloMappings != nil {
		in, out := &in.SloMappings, &out.SloMappings
		*out = make([]autoscalingv1.CrossVersionObjectReference, len(*in))
		copy(*out, *in)
	}
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]ServiceGraphCondition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServiceGraphStatus.
func (in *ServiceGraphStatus) DeepCopy() *ServiceGraphStatus {
	if in == nil {
		return nil
	}
	out := new(ServiceGraphStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceLevelObjective) DeepCopyInto(out *ServiceLevelObjective) {
	*out = *in
	out.SloType = in.SloType
	in.SloUserConfig.DeepCopyInto(&out.SloUserConfig)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServiceLevelObjective.
func (in *ServiceLevelObjective) DeepCopy() *ServiceLevelObjective {
	if in == nil {
		return nil
	}
	out := new(ServiceLevelObjective)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceLink) DeepCopyInto(out *ServiceLink) {
	*out = *in
	if in.QosRequirements != nil {
		in, out := &in.QosRequirements, &out.QosRequirements
		*out = new(LinkQosRequirements)
		(*in).DeepCopyInto(*out)
	}
	if in.TrustRequirements != nil {
		in, out := &in.TrustRequirements, &out.TrustRequirements
		*out = new(LinkTrustRequirements)
		**out = **in
	}
	if in.SLOs != nil {
		in, out := &in.SLOs, &out.SLOs
		*out = make([]ServiceLevelObjective, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServiceLink.
func (in *ServiceLink) DeepCopy() *ServiceLink {
	if in == nil {
		return nil
	}
	out := new(ServiceLink)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SloUserConfig) DeepCopyInto(out *SloUserConfig) {
	*out = *in
	out.ElasticityStrategy = in.ElasticityStrategy
	in.SloConfig.DeepCopyInto(&out.SloConfig)
	if in.StabilizationWindow != nil {
		in, out := &in.StabilizationWindow, &out.StabilizationWindow
		*out = new(slov1.StabilizationWindow)
		(*in).DeepCopyInto(*out)
	}
	if in.StaticElasticityStrategyConfig != nil {
		in, out := &in.StaticElasticityStrategyConfig, &out.StaticElasticityStrategyConfig
		*out = new(runtime.RawExtension)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SloUserConfig.
func (in *SloUserConfig) DeepCopy() *SloUserConfig {
	if in == nil {
		return nil
	}
	out := new(SloUserConfig)
	in.DeepCopyInto(out)
	return out
}
