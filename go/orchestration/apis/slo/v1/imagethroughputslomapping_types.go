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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Important: Run `make` and `make manifests` to regenerate code and YAML files after modifying this file.

// +kubebuilder:validation:Required

// ImageThroughputSloConfig contains the configuration for a ImageThroughputSloMappingSpec.
type ImageThroughputSloConfig struct {

	// The desired number of images that should be processed per minute.
	//
	// +kubebuilder:validation:Minimum=1
	TargetImagesPerMinute int32 `json:"targetImagesPerMinute"`

	// The minimum CPU usage percentage that must be achieved before scaling out on a
	// too low targetImagesPerMinute rate.
	//
	// +optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default=70
	MinCpuUsage int32 `json:"minCpuUsage,omitempty"`
}

// ImageThroughputSloMappingSpec represents an SloMapping for the Image Throughput SLO.
type ImageThroughputSloMappingSpec struct {
	SloMapping SloMapping `json:",inline"`

	SloConfig ImageThroughputSloConfig `json:"sloConfig"`
}

// ImageThroughputSloMappingStatus defines the observed state of ImageThroughputSloMapping
type ImageThroughputSloMappingStatus struct {
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ImageThroughputSloMapping is the Schema for the imagethroughputslomappings API
type ImageThroughputSloMapping struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ImageThroughputSloMappingSpec   `json:"spec,omitempty"`
	Status ImageThroughputSloMappingStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ImageThroughputSloMappingList contains a list of ImageThroughputSloMapping
type ImageThroughputSloMappingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ImageThroughputSloMapping `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ImageThroughputSloMapping{}, &ImageThroughputSloMappingList{})
}
