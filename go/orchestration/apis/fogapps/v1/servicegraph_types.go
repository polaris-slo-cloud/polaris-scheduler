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
	autoscaling "k8s.io/api/autoscaling/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// All fields in these CRDs are required, except if marked as `// +optional`
// +kubebuilder:validation:Required

// IMPORTANT: Run `make` and `make manifests` to regenerate code and YAML files after modifying this file.
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ServiceGraphSpec defines the desired state of a ServiceGraph
type ServiceGraphSpec struct {

	// Designates the default service account used for running the services described by the nodes
	// and thus, defines the default permissions that the application’s services have.
	//
	// +optional
	ServiceAccountName *string `json:"serviceAccountName,omitempty"`

	// The set of nodes in this ServiceGraph.
	Nodes []ServiceGraphNode `json:"nodes"`

	// The set of links between the nodes.
	//
	// +optional
	Links []ServiceLink `json:"links,omitempty"`

	// The SLOs defined for the entire application described by this ServiceGraph.
	//
	// +optional
	SLOs []ServiceLevelObjective `json:"slos,omitempty"`

	// The set of RAINBOW services that should be available to the entire application.
	//
	// +optional
	RainbowServices []RainbowService `json:"rainbowServices,omitempty"`

	// ConfigMaps that are available to all components of the application.
	//
	// +optional
	ConfigMaps []ConfigMap `json:"configMaps,omitempty"`

	// Allows configuring DNS for all pods created from this ServiceGraph.
	//
	// +optional
	DNSConfig *DNSConfig `json:"dnsConfig,omitempty"`

	// ToDo: add secrets
}

// ServiceGraphStatus defines the observed state of ServiceGraph
type ServiceGraphStatus struct {

	// The last metadata.Generation value that was observed by the controller.
	//
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Describes the state of the resources created from each ServiceGraphNode, indexed by ServiceGraphNode.Name.
	//
	// Note that it is possible that not all nodes from the ServiceGraph have been added to this map yet.
	//
	// +optional
	NodeStates map[string]*ServiceGraphNodeStatus `json:"nodeStates,omitempty"`

	// Lists the SloMappings that were created from this ServiceGraph.
	//
	// +optional
	SloMappings []autoscaling.CrossVersionObjectReference `json:"sloMappings,omitempty"`

	// The latest available high-level state information on this ServiceGraph.
	//
	// +optional
	Conditions []ServiceGraphCondition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ServiceGraph describes a fog application that should be deployed on RAINBOW.
type ServiceGraph struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ServiceGraphSpec   `json:"spec,omitempty"`
	Status ServiceGraphStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ServiceGraphList contains a list of ServiceGraph
type ServiceGraphList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ServiceGraph `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ServiceGraph{}, &ServiceGraphList{})
}
