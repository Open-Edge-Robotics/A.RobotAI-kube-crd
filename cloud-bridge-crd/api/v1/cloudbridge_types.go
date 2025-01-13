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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CloudBridgeSpec defines the desired state of CloudBridge
type CloudBridgeSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of CloudBridge. Edit cloudbridge_types.go to remove/update
	Network      string   `json:"network,omitempty"`
	RobotName    string   `json:"robotName,omitempty"`
	RosDomainId  string   `json:"rosDomainId,omitempty"`
	Image        string   `json:"image,omitempty"`
	Port         int32    `json:"port,omitempty"`
	ConfigDir    string   `json:"configDir,omitempty"`
	Command      []string `json:"command,omitempty"`
	Args         []string `json:"args,omitempty"`
	CloudIP      string   `json:"cloudIP,omitempty"`
	NodeSelector string   `json:"nodeSelector,omitempty"`
}

// CloudBridgeStatus defines the observed state of CloudBridge
type CloudBridgeStatus struct {
	Nodes      []string           `json:"nodes"`
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// CloudBridge is the Schema for the cloudbridges API
type CloudBridge struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CloudBridgeSpec   `json:"spec,omitempty"`
	Status CloudBridgeStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CloudBridgeList contains a list of CloudBridge
type CloudBridgeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CloudBridge `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CloudBridge{}, &CloudBridgeList{})
}
