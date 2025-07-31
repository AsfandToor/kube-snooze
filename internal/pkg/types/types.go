package types

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

type SnoozableResource interface {
	GetName() string
	GetNamespace() string
	GetAnnotations() map[string]string
	SetAnnotations(annotations map[string]string)
	IsSnoozed() bool
	Snooze(ctx context.Context, r client.Client) error
	Wake(ctx context.Context, r client.Client) error
	GetResourceType() string
}
