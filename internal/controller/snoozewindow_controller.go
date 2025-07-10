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

package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	schedulingv1alpha1 "codeacme.org/kube-snooze/api/v1alpha1"
)

const (
	// Annotation keys
	SnoozeEnabledAnnotation = "kube-snooze/enabled"
	SnoozePolicyAnnotation  = "kube-snooze/policy"
	BackupReplicasKey       = "kube-snooze/backup-replicas"
	BackupStateKey          = "kube-snooze/backup-state"

	// State constants
	StateSnoozed = "snoozed"
	StateAwake   = "awake"

	// Condition types
	ConditionReady   = "Ready"
	ConditionSnoozed = "Snoozed"
)

// SnoozeWindowReconciler reconciles a SnoozeWindow object
type SnoozeWindowReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=scheduling.codeacme.org,resources=snoozewindows,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=scheduling.codeacme.org,resources=snoozewindows/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=scheduling.codeacme.org,resources=snoozewindows/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=deployments;statefulsets,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=batch,resources=cronjobs;jobs,verbs=get;list;watch;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=pods;configmaps,verbs=get;list;watch;update;patch;create;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *SnoozeWindowReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := logf.FromContext(ctx)

	// Fetch the SnoozeWindow instance
	snoozeWindow := &schedulingv1alpha1.SnoozeWindow{}
	if err := r.Get(ctx, req.NamespacedName, snoozeWindow); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get SnoozeWindow")
		return ctrl.Result{}, err
	}

	logger.Info("Reconciling SnoozeWindow", "name", snoozeWindow.Name, "namespace", snoozeWindow.Namespace)

	// Determine if we should snooze or wake based on schedule
	shouldSnooze, shouldWake, nextReconcile, err := r.evaluateSchedule(ctx, snoozeWindow)
	if err != nil {
		logger.Error(err, "Failed to evaluate schedule")
		return ctrl.Result{}, err
	}

	// Find resources to manage
	resources, err := r.findResources(ctx, snoozeWindow)
	if err != nil {
		logger.Error(err, "Failed to find resources")
		return ctrl.Result{}, err
	}

	// Update status with managed resources count
	snoozeWindow.Status.ManagedResources = int32(len(resources))

	// Perform snooze or wake actions
	if shouldSnooze && snoozeWindow.Status.CurrentState != StateSnoozed {
		if err := r.snoozeResources(ctx, snoozeWindow, resources); err != nil {
			logger.Error(err, "Failed to snooze resources")
			return ctrl.Result{}, err
		}
		snoozeWindow.Status.CurrentState = StateSnoozed
		snoozeWindow.Status.LastSnoozeTime = &metav1.Time{Time: time.Now()}
	} else if shouldWake && snoozeWindow.Status.CurrentState != StateAwake {
		if err := r.wakeResources(ctx, snoozeWindow, resources); err != nil {
			logger.Error(err, "Failed to wake resources")
			return ctrl.Result{}, err
		}
		snoozeWindow.Status.CurrentState = StateAwake
		snoozeWindow.Status.LastWakeTime = &metav1.Time{Time: time.Now()}
	}

	// Update next schedule times
	snoozeWindow.Status.NextSnoozeTime = &metav1.Time{Time: nextReconcile}
	snoozeWindow.Status.NextWakeTime = &metav1.Time{Time: nextReconcile.Add(time.Hour)} // Simplified

	// Update conditions
	r.updateConditions(snoozeWindow)

	// Update status
	if err := r.Status().Update(ctx, snoozeWindow); err != nil {
		logger.Error(err, "Failed to update status")
		return ctrl.Result{}, err
	}

	// Return result with next reconciliation time
	return ctrl.Result{RequeueAfter: time.Until(nextReconcile)}, nil
}

// evaluateSchedule determines if resources should be snoozed or woken based on the schedule
func (r *SnoozeWindowReconciler) evaluateSchedule(ctx context.Context, snoozeWindow *schedulingv1alpha1.SnoozeWindow) (bool, bool, time.Time, error) {
	now := time.Now()

	// Parse timezone
	loc := time.UTC
	if snoozeWindow.Spec.Timezone != "" {
		var err error
		loc, err = time.LoadLocation(snoozeWindow.Spec.Timezone)
		if err != nil {
			return false, false, now, fmt.Errorf("invalid timezone %s: %w", snoozeWindow.Spec.Timezone, err)
		}
	}
	now = now.In(loc)

	// Check snooze schedule
	shouldSnooze := false
	if snoozeWindow.Spec.SnoozeSchedule != nil {
		if snoozeWindow.Spec.SnoozeSchedule.CronExpression != "" {
			parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
			schedule, err := parser.Parse(snoozeWindow.Spec.SnoozeSchedule.CronExpression)
			if err != nil {
				return false, false, now, fmt.Errorf("invalid cron expression: %w", err)
			}
			shouldSnooze = schedule.Next(now.Add(-time.Minute)).Before(now)
		}
	}

	// Check wake schedule
	shouldWake := false
	if snoozeWindow.Spec.WakeSchedule != nil {
		if snoozeWindow.Spec.WakeSchedule.CronExpression != "" {
			parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
			schedule, err := parser.Parse(snoozeWindow.Spec.WakeSchedule.CronExpression)
			if err != nil {
				return false, false, now, fmt.Errorf("invalid cron expression: %w", err)
			}
			shouldWake = schedule.Next(now.Add(-time.Minute)).Before(now)
		}
	}

	// Calculate next reconciliation time (simplified - check every minute)
	nextReconcile := now.Add(time.Minute)

	return shouldSnooze, shouldWake, nextReconcile, nil
}

// findResources discovers resources that match the snooze window criteria
func (r *SnoozeWindowReconciler) findResources(ctx context.Context, snoozeWindow *schedulingv1alpha1.SnoozeWindow) ([]unstructured.Unstructured, error) {
	var resources []unstructured.Unstructured
	namespace := snoozeWindow.Spec.Namespace
	if namespace == "" {
		namespace = snoozeWindow.Namespace
	}

	// Build label selector
	var selector labels.Selector
	if len(snoozeWindow.Spec.LabelSelector) > 0 {
		selector = labels.SelectorFromSet(snoozeWindow.Spec.LabelSelector)
	}

	// Find resources for each resource type
	for _, resourceType := range snoozeWindow.Spec.ResourceTypes {
		gvk := schema.FromAPIVersionAndKind(resourceType.APIVersion, resourceType.Kind)

		// Create list object
		list := &unstructured.UnstructuredList{}
		list.SetGroupVersionKind(gvk)

		// List resources
		listOpts := &client.ListOptions{
			Namespace: namespace,
		}
		if selector != nil {
			listOpts.LabelSelector = selector
		}

		if err := r.List(ctx, list, listOpts); err != nil {
			return nil, fmt.Errorf("failed to list %s: %w", resourceType.Kind, err)
		}

		// Filter resources that have snooze enabled
		for _, resource := range list.Items {
			if r.isSnoozeEnabled(&resource) {
				resources = append(resources, resource)
			}
		}
	}

	return resources, nil
}

// isSnoozeEnabled checks if a resource has snooze enabled via annotations
func (r *SnoozeWindowReconciler) isSnoozeEnabled(resource *unstructured.Unstructured) bool {
	annotations := resource.GetAnnotations()
	if enabled, exists := annotations[SnoozeEnabledAnnotation]; exists && enabled == "true" {
		return true
	}
	return false
}

// snoozeResources puts resources into snooze mode
func (r *SnoozeWindowReconciler) snoozeResources(ctx context.Context, snoozeWindow *schedulingv1alpha1.SnoozeWindow, resources []unstructured.Unstructured) error {
	logger := logf.FromContext(ctx)

	for _, resource := range resources {
		// Backup current state
		if err := r.backupResourceState(ctx, &resource); err != nil {
			logger.Error(err, "Failed to backup resource state", "resource", resource.GetName())
			continue
		}

		// Apply snooze action
		if err := r.applySnoozeAction(ctx, &resource, &snoozeWindow.Spec.SnoozeAction); err != nil {
			logger.Error(err, "Failed to apply snooze action", "resource", resource.GetName())
			continue
		}

		logger.Info("Resource snoozed", "resource", resource.GetName(), "kind", resource.GetKind())
	}

	return nil
}

// wakeResources restores resources from snooze mode
func (r *SnoozeWindowReconciler) wakeResources(ctx context.Context, snoozeWindow *schedulingv1alpha1.SnoozeWindow, resources []unstructured.Unstructured) error {
	logger := logf.FromContext(ctx)

	for _, resource := range resources {
		// Restore original state
		if err := r.restoreResourceState(ctx, &resource); err != nil {
			logger.Error(err, "Failed to restore resource state", "resource", resource.GetName())
			continue
		}

		logger.Info("Resource woken", "resource", resource.GetName(), "kind", resource.GetKind())
	}

	return nil
}

// backupResourceState saves the original state of a resource
func (r *SnoozeWindowReconciler) backupResourceState(ctx context.Context, resource *unstructured.Unstructured) error {
	annotations := resource.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}

	// Backup replicas for scalable resources
	if replicas, found, err := unstructured.NestedInt64(resource.Object, "spec", "replicas"); err == nil && found {
		annotations[BackupReplicasKey] = fmt.Sprintf("%d", replicas)
	}

	// Backup full state if configured
	if backupConfig := resource.GetAnnotations()["kube-snooze/backup-full-state"]; backupConfig == "true" {
		stateBytes, err := json.Marshal(resource.Object)
		if err == nil {
			annotations[BackupStateKey] = string(stateBytes)
		}
	}

	resource.SetAnnotations(annotations)
	return r.Update(ctx, resource)
}

// restoreResourceState restores the original state of a resource
func (r *SnoozeWindowReconciler) restoreResourceState(ctx context.Context, resource *unstructured.Unstructured) error {
	annotations := resource.GetAnnotations()

	// Restore replicas
	if backupReplicas, exists := annotations[BackupReplicasKey]; exists {
		if replicas, err := json.Marshal(backupReplicas); err == nil {
			unstructured.SetNestedField(resource.Object, replicas, "spec", "replicas")
		}
	}

	// Restore full state if available
	if backupState, exists := annotations[BackupStateKey]; exists {
		var originalState map[string]interface{}
		if err := json.Unmarshal([]byte(backupState), &originalState); err == nil {
			resource.Object = originalState
		}
	}

	// Clean up backup annotations
	delete(annotations, BackupReplicasKey)
	delete(annotations, BackupStateKey)
	resource.SetAnnotations(annotations)

	return r.Update(ctx, resource)
}

// applySnoozeAction applies the specified snooze action to a resource
func (r *SnoozeWindowReconciler) applySnoozeAction(ctx context.Context, resource *unstructured.Unstructured, action *schedulingv1alpha1.SnoozeAction) error {
	if action.ScaleToZero {
		// Scale to zero replicas
		unstructured.SetNestedField(resource.Object, int64(0), "spec", "replicas")
		return r.Update(ctx, resource)
	}

	if action.Delete {
		// Delete the resource
		return r.Delete(ctx, resource)
	}

	if action.Patch != nil {
		// Apply custom patch
		return r.applyPatch(ctx, resource, action.Patch)
	}

	return nil
}

// applyPatch applies a custom patch to a resource
func (r *SnoozeWindowReconciler) applyPatch(ctx context.Context, resource *unstructured.Unstructured, patch *schedulingv1alpha1.Patch) error {
	// For simplicity, we'll use strategic merge patch
	patchBytes := patch.Data.Raw

	return r.Patch(ctx, resource, client.RawPatch(types.StrategicMergePatchType, patchBytes))
}

// updateConditions updates the status conditions
func (r *SnoozeWindowReconciler) updateConditions(snoozeWindow *schedulingv1alpha1.SnoozeWindow) {
	now := metav1.Now()

	// Update Ready condition
	readyCondition := metav1.Condition{
		Type:               ConditionReady,
		Status:             metav1.ConditionTrue,
		Reason:             "Reconciled",
		Message:            fmt.Sprintf("SnoozeWindow is %s", snoozeWindow.Status.CurrentState),
		LastTransitionTime: now,
	}

	// Update Snoozed condition
	snoozedCondition := metav1.Condition{
		Type:               ConditionSnoozed,
		Status:             metav1.ConditionFalse,
		Reason:             "Awake",
		Message:            "Resources are awake",
		LastTransitionTime: now,
	}

	if snoozeWindow.Status.CurrentState == StateSnoozed {
		snoozedCondition.Status = metav1.ConditionTrue
		snoozedCondition.Reason = "Snoozed"
		snoozedCondition.Message = "Resources are snoozed"
	}

	// Update conditions
	metav1.SetMetaDataAnnotation(&snoozeWindow.ObjectMeta, "kube-snooze/last-updated", now.Format(time.RFC3339))
}

// SetupWithManager sets up the controller with the Manager.
func (r *SnoozeWindowReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&schedulingv1alpha1.SnoozeWindow{}).
		Named("snoozewindow").
		Complete(r)
}
