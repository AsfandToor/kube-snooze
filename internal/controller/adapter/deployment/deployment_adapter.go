package deployment

import (
	"context"
	"strconv"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	BackupReplicasKey = "kube-snooze/replicas"
)

type DeploymentAdapter struct {
	deployment *appsv1.Deployment
}

func NewDeploymentAdapter(deployment *appsv1.Deployment) *DeploymentAdapter {
	return &DeploymentAdapter{deployment: deployment}
}

func (d *DeploymentAdapter) GetName() string {
	return d.deployment.Name
}

func (d *DeploymentAdapter) GetNamespace() string {
	return d.deployment.Namespace
}

func (d *DeploymentAdapter) GetAnnotations() map[string]string {
	return d.deployment.GetAnnotations()
}

func (d *DeploymentAdapter) SetAnnotations(annotations map[string]string) {
	d.deployment.SetAnnotations(annotations)
}

func (d *DeploymentAdapter) IsSnoozed() bool {
	annotations := d.deployment.GetAnnotations()
	_, isSnoozed := annotations["kube-snooze/replicas"]
	return isSnoozed
}

func (d *DeploymentAdapter) Snooze(ctx context.Context, r client.Client) error {
	replicas := strconv.Itoa(int(*d.deployment.Spec.Replicas))
	d.deployment.Spec.Replicas = pointer.Int32Ptr(0)
	annotations := d.deployment.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}
	annotations[BackupReplicasKey] = replicas
	d.deployment.SetAnnotations(annotations)
	return r.Update(ctx, d.deployment)
}

func (d *DeploymentAdapter) Wake(ctx context.Context, r client.Client) error {
	return nil
}

func (d *DeploymentAdapter) GetResourceType() string {
	return "deployment"
}
