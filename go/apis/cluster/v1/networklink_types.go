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

// CONTROLLER INFO: To create a controller for this CRD, run the kubebuilder command
// kubebuilder create api --group cluster --version v1 --kind NetworkLink
// again and answer "no" for the resource and "yes" for the controller.

// IMPORTANT: Run `make` and `make manifests` to regenerate code and YAML files after modifying this file.
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// NetworkLinkSpec defines the desired state of NetworkLink
type NetworkLinkSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of NetworkLink. Edit networklink_types.go to remove/update
	Foo string `json:"foo,omitempty"`
}

// NetworkLinkStatus defines the observed state of NetworkLink
type NetworkLinkStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// NetworkLink is the Schema for the networklinks API
type NetworkLink struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NetworkLinkSpec   `json:"spec,omitempty"`
	Status NetworkLinkStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// NetworkLinkList contains a list of NetworkLink
type NetworkLinkList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NetworkLink `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NetworkLink{}, &NetworkLinkList{})
}
