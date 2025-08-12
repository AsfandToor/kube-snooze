package workloads

import (
	"context"
	"strconv"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ReplicaSetAdapter struct {
	replicaset *appsv1.ReplicaSet
}

func NewReplicaSetAdapter(replicaset *appsv1.ReplicaSet) *ReplicaSetAdapter {
	return &ReplicaSetAdapter{replicaset: replicaset}
}

func (r *ReplicaSetAdapter) GetName() string {
	return r.replicaset.Name
}

func (r *ReplicaSetAdapter) GetNamespace() string {
	return r.replicaset.Namespace
}

func (r *ReplicaSetAdapter) GetAnnotations() map[string]string {
	return r.replicaset.GetAnnotations()
}

func (r *ReplicaSetAdapter) SetAnnotations(annotations map[string]string) {
	r.replicaset.SetAnnotations(annotations)
}

func (r *ReplicaSetAdapter) IsSnoozed() bool {
	annotations := r.GetAnnotations()
	_, isSnoozed := annotations[BackupReplicasKey]
	return isSnoozed
}

func (r *ReplicaSetAdapter) Snooze(ctx context.Context, client client.Client) error {
	replicas := strconv.Itoa(int(*r.replicaset.Spec.Replicas))
	r.replicaset.Spec.Replicas = ptr.To[int32](0)
	annotations := r.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}
	annotations[BackupReplicasKey] = replicas
	r.SetAnnotations(annotations)
	return client.Update(ctx, r.replicaset)
}

func (r *ReplicaSetAdapter) Wake(ctx context.Context, client client.Client) error {
	annotations := r.replicaset.GetAnnotations()
	if storedReplicas, exists := annotations[BackupReplicasKey]; exists {
		desiredReplicas, err := strconv.ParseInt(storedReplicas, 10, 32)
		if err != nil {
			return err
		}

		if desiredReplicas > 0 {
			r.replicaset.Spec.Replicas = ptr.To[int32](int32(desiredReplicas))
		}

		// Clean up annotation
		delete(annotations, BackupReplicasKey)
		r.SetAnnotations(annotations)
	}

	return client.Update(ctx, r.replicaset)
}

func (r *ReplicaSetAdapter) GetResourceType() string {
	return "replicaset"
}
