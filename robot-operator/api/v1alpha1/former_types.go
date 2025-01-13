/*
Copyright 2024 seo.youngchae.

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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// FormerSpec defines the desired state of Former
type FormerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	RobotName    string            `json:"robotName,omitempty"`
	RosDomainId  string            `json:"rosDomainId,omitempty"`
	NetWork      string            `json:"network,omitempty"`
	Image        string            `json:"image,omitempty"`
	Args         string            `json:"args,omitempty"`
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	HostNetwork  bool              `json:"hostNetwork,omitempty"`
}

// FormerStatus defines the observed state of Former
type FormerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Former is the Schema for the formers API
type Former struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FormerSpec   `json:"spec,omitempty"`
	Status FormerStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// FormerList contains a list of Former
type FormerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Former `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Former{}, &FormerList{})
}
