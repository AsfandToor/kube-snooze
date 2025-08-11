package workloads

import (
	"context"
	"strconv"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/utils/pointer"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type StatefulSetAdapter struct {
	statefulset *appsv1.StatefulSet
}

func NewStatefulSetAdapter(statefulset *appsv1.StatefulSet) *StatefulSetAdapter {
	return &StatefulSetAdapter{statefulset: statefulset}
}

func (s *StatefulSetAdapter) GetName() string {
	return s.statefulset.Name
}

func (s *StatefulSetAdapter) GetNamespace() string {
	return s.statefulset.Namespace
}

func (s *StatefulSetAdapter) GetAnnotations() map[string]string {
	return s.statefulset.Annotations
}

func (s *StatefulSetAdapter) SetAnnotations(annotations map[string]string) {
	s.statefulset.SetAnnotations(annotations)
}

func (s *StatefulSetAdapter) IsSnoozed() bool {
	annotations := s.statefulset.GetAnnotations()
	_, isSnoozed := annotations[BackupReplicasKey]
	return isSnoozed
}

func (s *StatefulSetAdapter) Snooze(ctx context.Context, r client.Client) error {
	replicas := strconv.Itoa(int(*s.statefulset.Spec.Replicas))
	s.statefulset.Spec.Replicas = ptr.To[int32](0)
	annotations := s.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}
	annotations[BackupReplicasKey] = replicas
	s.SetAnnotations(annotations)
	return r.Update(ctx, s.statefulset)
}

func (s *StatefulSetAdapter) Wake(ctx context.Context, r client.Client) error {
	annotations := s.GetAnnotations()
	if storedReplicas, exists := annotations[BackupReplicasKey]; exists {
		desiredReplicas, err := strconv.ParseInt(storedReplicas, 10, 32)
		if err != nil {
			return err
		}

		if desiredReplicas > 0 {
			s.statefulset.Spec.Replicas = pointer.Int32(int32(desiredReplicas))
		}

		delete(annotations, BackupReplicasKey)
		s.SetAnnotations(annotations)
	}

	return r.Update(ctx, s.statefulset)
}

func (s *StatefulSetAdapter) GetResourceType() string {
	return "statefulset"
}
