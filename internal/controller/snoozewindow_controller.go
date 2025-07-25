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
	"strconv"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	schedulingv1alpha1 "codeacme.org/kube-snooze/api/v1alpha1"
	"codeacme.org/kube-snooze/internal/utils"
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
	logger.Info("Firing up SnoozeScheduler")

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

	isSnoozeActive, duration, err := utils.IsTimeOngoing(snoozeWindow.Spec.SnoozeSchedule.StartTime, snoozeWindow.Spec.SnoozeSchedule.EndTime, snoozeWindow.Spec.SnoozeSchedule.Date)
	if err != nil {
		logger.Error(err, "parsing snooze schedule")
	}

	var deployments appsv1.DeploymentList

	labelSelectors := client.MatchingLabels{
		"kube-snooze/enabled": "true",
	}

	if err := r.List(ctx, &deployments, client.InNamespace(req.Namespace), labelSelectors); err != nil {
		logger.Error(err, "failed to list deployments")
		return ctrl.Result{}, err
	}

	for _, deploy := range deployments.Items {
		logger.Info("TargetingDeployment", "name")
		if isSnoozeActive {
			replicas := strconv.Itoa(int(*deploy.Spec.Replicas))
			deploy.Spec.Replicas = pointer.Int32Ptr(0)

			// Save deployment replicas in the respective annotations.
			deploy.SetAnnotations(map[string]string{
				"kube-snooze/replicas": replicas,
			})

			if err := r.Update(ctx, &deploy); err != nil {
				logger.Error(err, "DeploymentsUpdateFailed")
				nextReconcile := time.Now().Add(time.Minute)
				return ctrl.Result{RequeueAfter: time.Until(nextReconcile)}, err
			}
		} else {
			// Make this dynamic and persist in annotations
			annotations := deploy.GetAnnotations()

			if _, ok := annotations["kube-snooze/replicas"]; ok {
				desiredReplicas, err := strconv.ParseInt(annotations["kube-snooze/replicas"], 10, 32)
				if err != nil {
					logger.Error(err, "ParsingStoredReplicas")
				}

				if desiredReplicas > 0 {
					deploy.Spec.Replicas = pointer.Int32Ptr(int32(desiredReplicas))
				}
			}
		}

		if err := r.Update(ctx, &deploy); err != nil {
			logger.Error(err, "ErrorUpdatingDeployment")
			return ctrl.Result{}, nil
		}
	}

	// Add this block to above part
	// snoozeWindow.Status.SleepyInstances = len(deployments.Items)
	//
	// // Update status
	// if err := r.Status().Update(ctx, snoozeWindow); err != nil {
	// 	logger.Error(err, "Failed to update status")
	// 	return ctrl.Result{}, err
	// }

	return ctrl.Result{RequeueAfter: duration}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SnoozeWindowReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&schedulingv1alpha1.SnoozeWindow{}).
		Named("snoozewindow").
		Complete(r)
}
