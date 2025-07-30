package types

import (
	"context"

	"codeacme.org/kube-snooze/internal/controller"
)

type SnoozableResource interface {
	GetName() string
	GetNamespace() string
	GetAnnotations() map[string]string
	SetAnnotations(annotations map[string]string)
	IsSnoozed() bool
	Snooze(ctx context.Context, r *controller.SnoozeWindowReconciler) error
	Wake(ctx context.Context, r *controller.SnoozeWindowReconciler) error
	GetResourceType() string
}
