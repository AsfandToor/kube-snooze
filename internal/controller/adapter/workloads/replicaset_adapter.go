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

func (rs *ReplicaSetAdapter) GetName() string {
	return rs.replicaset.Name
}

func (rs *ReplicaSetAdapter) GetNamespace() string {
	return rs.replicaset.Namespace
}

func (rs *ReplicaSetAdapter) GetAnnotations() map[string]string {
	return rs.replicaset.GetAnnotations()
}

func (rs *ReplicaSetAdapter) SetAnnotations(annotations map[string]string) {
	rs.replicaset.SetAnnotations(annotations)
}

func (rs *ReplicaSetAdapter) IsSnoozed() bool {
	annotations := rs.GetAnnotations()
	_, isSnoozed := annotations[BackupReplicasKey]
	return isSnoozed
}

func (rs *ReplicaSetAdapter) Snooze(ctx context.Context, r client.Client) error {
	replicas := strconv.Itoa(int(*rs.replicaset.Spec.Replicas))
	rs.replicaset.Spec.Replicas = ptr.To[int32](0)
	annotations := rs.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}
	annotations[BackupReplicasKey] = replicas
	rs.SetAnnotations(annotations)
	return r.Update(ctx, rs.replicaset)
}

func (rs *ReplicaSetAdapter) Wake(ctx context.Context, r client.Client) error {
	annotations := rs.replicaset.GetAnnotations()
	if storedReplicas, exists := annotations[BackupReplicasKey]; exists {
		desiredReplicas, err := strconv.ParseInt(storedReplicas, 10, 32)
		if err != nil {
			return err
		}

		if desiredReplicas > 0 {
			rs.replicaset.Spec.Replicas = ptr.To(int32(desiredReplicas))
		}

		// Clean up annotation
		delete(annotations, BackupReplicasKey)
		rs.SetAnnotations(annotations)
	}

	return r.Update(ctx, rs.replicaset)
}

func (rs *ReplicaSetAdapter) GetResourceType() string {
	return "replicaset"
}
