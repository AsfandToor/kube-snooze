/*
Copyright 2025.

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

// SnoozeWindowSpec defines the desired state of SnoozeWindow.
type SnoozeWindowSpec struct {
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
	TimeZone  string `json:"timeZone"`
	Selector  string `json:"selector,omitempty"`
}

// SnoozeWindowStatus defines the observed state of SnoozeWindow.
type SnoozeWindowStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// SnoozeWindow is the Schema for the snoozewindows API.
type SnoozeWindow struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SnoozeWindowSpec   `json:"spec,omitempty"`
	Status SnoozeWindowStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SnoozeWindowList contains a list of SnoozeWindow.
type SnoozeWindowList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SnoozeWindow `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SnoozeWindow{}, &SnoozeWindowList{})
}
