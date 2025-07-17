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
	"time"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/pointer"
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

func (r *SnoozeWindowReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := logf.FromContext(ctx)
	logger.Info("Fire it up and ready to serve! ~ Blitzcrank")

	var deployment appsv1.Deployment

	if err := r.Get(ctx, req.NamespacedName, &deployment); err != nil {
		logger.Info("Error Getting Deployments", "error", err)
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
	}

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

	if snoozeWindow.Spec.SnoozeSchedule == "true" {
		deployment.Spec.Replicas = pointer.Int32Ptr(0)
		logger.Info("Updating Deployment", "name", deployment.ObjectMeta.Name)
	}

	if err := r.Update(ctx, &deployment); err != nil {
		return ctrl.Result{}, nil
	}

	// Update status
	if err := r.Status().Update(ctx, snoozeWindow); err != nil {
		logger.Error(err, "Failed to update status")
		return ctrl.Result{}, err
	}

	// Return result with next reconciliation time
	// Simplified Requeue time to check after minutes

	nextReconcile := time.Now().Add(time.Minute)
	return ctrl.Result{RequeueAfter: time.Until(nextReconcile)}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SnoozeWindowReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&schedulingv1alpha1.SnoozeWindow{}).
		Named("snoozewindow").
		Complete(r)
}
