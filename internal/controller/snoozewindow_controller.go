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
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	schedulingv1alpha1 "codeacme.org/kube-snooze/api/v1alpha1"
	"codeacme.org/kube-snooze/internal/controller/adapter"
	"codeacme.org/kube-snooze/internal/controller/adapter/deployment"
	"codeacme.org/kube-snooze/internal/utils"
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

	isSnoozeActive, hasWindowPassed, duration, err := utils.IsTimeOngoing(snoozeWindow.Spec.SnoozeSchedule.StartTime, snoozeWindow.Spec.SnoozeSchedule.EndTime, snoozeWindow.Spec.SnoozeSchedule.Date)
	if err != nil {
		logger.Error(err, "parsing snooze schedule")
	}

	resourceManager, err := r.buildResourceManager(ctx, snoozeWindow.Namespace)
	if err != nil {
		logger.Error(err, "failed to build resource manager")
		return ctrl.Result{}, err
	}

	if isSnoozeActive {
		if err := resourceManager.SnoozeAll(ctx, r.Client); err != nil {
			logger.Error(err, "failed to snooze resources")
			return ctrl.Result{}, err
		}

		logger.Info("RequeingScheduler", "interval", duration)
		return ctrl.Result{RequeueAfter: duration}, nil
	} else {
		if hasWindowPassed {
			if err := resourceManager.WakeAll(ctx, r.Client); err != nil {
				logger.Error(err, "failed to wake resources")
				return ctrl.Result{}, err
			}
		}

		// Change to Rerun at the time of the Snooze Start time
		logger.Info("RequeingScheduler", "interval", "10 seconds")
		return ctrl.Result{RequeueAfter: time.Second * 10}, nil
	}
}

func (r *SnoozeWindowReconciler) buildResourceManager(ctx context.Context, namespace string) (*adapter.ResourceManager, error) {
	resourceManager := adapter.NewResourceManager()
	labelSelectors := client.MatchingLabels{
		"kube-snooze/enabled": "true",
	}

	var deployments appsv1.DeploymentList
	if err := r.List(ctx, &deployments, client.InNamespace(namespace), labelSelectors); err != nil {
		return nil, err
	}
	for _, deploy := range deployments.Items {
		resourceManager.AddResource(deployment.NewDeploymentAdapter(&deploy))
	}

	return resourceManager, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SnoozeWindowReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&schedulingv1alpha1.SnoozeWindow{}).
		Named("snoozewindow").
		Complete(r)
}
