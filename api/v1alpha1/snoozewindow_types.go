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
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SnoozeWindowSpec defines the desired state of SnoozeWindow.
type SnoozeWindowSpec struct {
	// Namespace specifies the Kubernetes namespace to apply snooze policies to
	Namespace string `json:"namespace,omitempty"`
	// LabelSelector defines key-value pairs to match resources for snoozing
	LabelSelector map[string]string `json:"labelSelector,omitempty"`

	// SnoozeSchedule defines when resources should be put into snooze mode
	SnoozeSchedule *ScheduleConfig `json:"snoozeSchedule"`
	// WakeSchedule defines when resources should be restored from snooze mode
	WakeSchedule *ScheduleConfig `json:"wakeSchedule"`
	// Timezone specifies the timezone for schedule calculations (e.g., "UTC", "America/New_York")
	Timezone string `json:"timezone,omitempty"`

	// ResourceTypes specifies which Kubernetes resource types to manage
	ResourceTypes []ResourceType `json:"resourceTypes"`

	// SnoozeAction defines what action to take when snoozing resources
	SnoozeAction SnoozeAction `json:"snoozeAction"`

	// BackupConfig defines how to store original resource state before snoozing
	BackupConfig *BackupConfig `json:"backupConfig,omitempty"`
}

type ScheduleConfig struct {
	// CronExpression defines a cron schedule (e.g., "0 22 * * 1-5" for weekdays at 10 PM)
	CronExpression string `json:"cronExpression,omitempty"`
	// RFC3339Time defines a specific point in time using RFC3339 format
	RFC3339Time string `json:"rfc3339Time,omitempty"`
	// Weekdays specifies which days of the week to apply the schedule (0=Sunday, 1=Monday, etc.)
	Weekdays []int `json:"weekdays,omitempty"`
	// Weekends determines if the schedule should apply on weekends
	Weekends bool `json:"weekends,omitempty"`
}

type ResourceType struct {
	// Kind specifies the Kubernetes resource kind (e.g., "Deployment", "StatefulSet")
	Kind string `json:"kind"` // Deployment, StatefulSet, CronJob, etc.
	// APIVersion specifies the Kubernetes API version (e.g., "apps/v1")
	APIVersion string `json:"apiVersion"`
	// ScaleToZero determines if this resource type should be scaled to zero replicas
	ScaleToZero bool `json:"scaleToZero,omitempty"`
	// Delete determines if this resource type should be deleted during snooze
	Delete bool `json:"delete,omitempty"`
	// Patch defines custom patches to apply to this resource type during snooze
	Patch *Patch `json:"patch,omitempty"`
}

type SnoozeAction struct {
	// ScaleToZero scales resources to zero replicas during snooze
	ScaleToZero bool `json:"scaleToZero,omitempty"`
	// Delete removes resources entirely during snooze
	Delete bool `json:"delete,omitempty"`
	// Patch applies custom modifications to resources during snooze
	Patch *Patch `json:"patch,omitempty"`
}

type Patch struct {
	// Type specifies the patch strategy ("strategic", "merge", or "json")
	Type string `json:"type"`
	// Data contains the patch data to apply to resources
	Data apiextensionsv1.JSON `json:"data"`
}

type BackupConfig struct {
	// StoreInAnnotations saves original state in resource annotations
	StoreInAnnotations bool `json:"storeInAnnotations,omitempty"`
	// StoreInConfigMap saves original state in a ConfigMap
	StoreInConfigMap bool `json:"storeInConfigMap,omitempty"`
	// ConfigMapName specifies the name of the ConfigMap to store backup data
	ConfigMapName string `json:"configMapName,omitempty"`
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
