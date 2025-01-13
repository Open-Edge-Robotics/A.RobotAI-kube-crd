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

// NaviSpec defines the desired state of Navi
type NaviSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of Navi. Edit navi_types.go to remove/update
	RosDomainId  string            `json:"rosDomainId,omitempty"`
	NetWork      string            `json:"network,omitempty"`
	Image        string            `json:"image,omitempty"`
	Args         string            `json:"args,omitempty"`
	ParamPath    string            `json:"paramPath,omitempty"`
	MapPath      string            `json:"mapPath,omitempty"`
	RvizPath     string            `json:"rvizPath,omitempty"`
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	HostNetwork  bool              `json:"hostNetwork,omitempty"`
}

// NaviStatus defines the observed state of Navi
type NaviStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Navi is the Schema for the navis API
type Navi struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NaviSpec   `json:"spec,omitempty"`
	Status NaviStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// NaviList contains a list of Navi
type NaviList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Navi `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Navi{}, &NaviList{})
}
