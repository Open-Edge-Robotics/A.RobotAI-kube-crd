/*
Copyright 2023.

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

// NavigationSpec defines the desired state of Navigation
type NavigationSpec struct {
	Network      string      `json:"network,omitempty"`
	Image        string      `json:"image,omitempty"`
	RobotName    string      `json:"robotName,omitempty"`
	RosDomainId  string      `json:"rosDomainId,omitempty"`
	InitialPose  InitialPose `json:"initialPose,omitempty"`
	Command      []string    `json:"command,omitempty"`
	Args         []string    `json:"args,omitempty"`
	NodeSelector string      `json:"nodeSelector,omitempty"`
}

// NavigationStatus defines the observed state of Navigation
type NavigationStatus struct {
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
	Nodes      []string           `json:"nodes"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Navigation is the Schema for the navigations API
type Navigation struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NavigationSpec   `json:"spec,omitempty"`
	Status NavigationStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// NavigationList contains a list of Navigation
type NavigationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Navigation `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Navigation{}, &NavigationList{})
}
