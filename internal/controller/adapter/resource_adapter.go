package adapter

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
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

type ResourceManager struct {
	resources []SnoozableResource
}

func (rm *ResourceManager) AddResource(resource SnoozableResource) {
	rm.resources = append(rm.resources, resource)
}

func (rm *ResourceManager) SnoozeAll(ctx context.Context, r client.Client) error {
	logger := logf.FromContext(ctx)

	for _, resource := range rm.resources {
		if resource.IsSnoozed() {
			logger.Info("Resource already snoozed, skipping",
				"type", resource.GetResourceType(),
				"name", resource.GetName())
			continue
		}

		logger.Info("Snoozing resource",
			"type", resource.GetResourceType(),
			"name", resource.GetName())

		if err := resource.Snooze(ctx, r); err != nil {
			logger.Error(err, "Failed to snooze resource",
				"type", resource.GetResourceType(),
				"name", resource.GetName())
			return err
		}
	}

	return nil
}

func (rm *ResourceManager) WakeAll(ctx context.Context, r client.Client) error {
	logger := logf.FromContext(ctx)

	for _, resource := range rm.resources {
		if !resource.IsSnoozed() {
			continue
		}

		logger.Info("Waking resource",
			"type", resource.GetResourceType(),
			"name", resource.GetName())

		if err := resource.Wake(ctx, r); err != nil {
			logger.Error(err, "Failed to wake resource",
				"type", resource.GetResourceType(),
				"name", resource.GetName())
			return err
		}
	}

	return nil
}
