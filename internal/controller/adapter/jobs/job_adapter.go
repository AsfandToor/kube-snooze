package jobs

import (
	"context"

	batchv1 "k8s.io/api/batch/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type JobAdapter struct {
	job *batchv1.Job
}

func NewJobAdapter(job *batchv1.Job) *JobAdapter {
	return &JobAdapter{job: job}
}

func (j *JobAdapter) GetName() string {
	return j.job.Name
}

func (j *JobAdapter) GetNamespace() string {
	return j.job.Namespace
}

func (j *JobAdapter) GetAnnotations() map[string]string {
	return j.job.Annotations
}

func (j *JobAdapter) SetAnnotations(annotations map[string]string) {
	j.job.SetAnnotations(annotations)
}

func (j *JobAdapter) IsSnoozed() bool {
	return j.job.Spec.Suspend != nil && *j.job.Spec.Suspend
}

func (j *JobAdapter) Snooze(ctx context.Context, r client.Client) error {
	isSuspended := true
	j.job.Spec.Suspend = &isSuspended
	return r.Update(ctx, j.job)
}

func (j *JobAdapter) Wake(ctx context.Context, r client.Client) error {
	isSuspended := false
	j.job.Spec.Suspend = &isSuspended
	return r.Update(ctx, j.job)
}

func (j *JobAdapter) GetResourceType() string {
	return "job"
}
