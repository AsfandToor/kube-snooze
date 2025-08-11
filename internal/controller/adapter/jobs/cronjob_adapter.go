package jobs

import (
	"context"

	batchv1 "k8s.io/api/batch/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type CronJobAdapter struct {
	cronjob *batchv1.CronJob
}

func NewCronJobAdapter(cronjob *batchv1.CronJob) *CronJobAdapter {
	return &CronJobAdapter{cronjob: cronjob}
}

func (c *CronJobAdapter) GetName() string {
	return c.cronjob.Name
}

func (c *CronJobAdapter) GetNamespace() string {
	return c.cronjob.Namespace
}

func (c *CronJobAdapter) GetAnnotations() string {
	return c.cronjob.Namespace
}

func (c *CronJobAdapter) SetAnnotations(annotations map[string]string) {
	c.cronjob.SetAnnotations(annotations)
}

func (c *CronJobAdapter) IsSnoozed() bool {
	return *c.cronjob.Spec.Suspend
}

func (c *CronJobAdapter) Snooze(ctx context.Context, r client.Client) error {
	isSuspended := true
	c.cronjob.Spec.Suspend = &isSuspended
	return r.Update(ctx, c.cronjob)
}

func (c *CronJobAdapter) Wake(ctx context.Context, r client.Client) error {
	isSuspended := false
	c.cronjob.Spec.Suspend = &isSuspended
	return r.Update(ctx, c.cronjob)
}

func (c *CronJobAdapter) GetResourceType() string {
	return "cronjob"
}
