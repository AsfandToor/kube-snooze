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

// SnoozeWindowSpec defines the desired state of SnoozeWindow.
type SnoozeWindowSpec struct {
	LabelSelector  map[string]string  `json:"label_selector,omitempty"`
	Timezone       string             `json:"timezone"`
	SnoozeSchedule SnoozeScheduleSpec `json:"snooze_schedule,omitempty"`
	WakeSchedule   string             `json:"wake_schedule,omitempty"`
}

type SnoozeScheduleSpec struct {
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	Date      string `json:"date"`
}

type SnoozeWindowStatus struct {
	SleepyInstances int                `json:"sleepy_instances,omitempty"`
	Conditions      []metav1.Condition `json:"conditions,omitempty"`
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
