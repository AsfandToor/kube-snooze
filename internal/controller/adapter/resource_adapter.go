package adapter

import (
	"context"

	"codeacme.org/kube-snooze/internal/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

type ResourceManager struct {
	resources []types.SnoozableResource
}

func NewResourceManager() *ResourceManager {
	return &ResourceManager{
		resources: make([]types.SnoozableResource, 0),
	}
}

func (rm *ResourceManager) AddResource(resource types.SnoozableResource) {
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
