package service

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ServiceAdapter struct {
	service *corev1.Service
}

func NewServiceAdapter(service *corev1.Service) *ServiceAdapter {
	return &ServiceAdapter{service: service}
}

func (s *ServiceAdapter) GetName() string {
	return s.service.Name
}

func (s *ServiceAdapter) GetNamespace() string {
	return s.service.Namespace
}

func (s *ServiceAdapter) GetAnnotations() map[string]string {
	return s.service.Annotations
}

func (s *ServiceAdapter) SetAnnotations(annotations map[string]string) {
	s.service.SetAnnotations(annotations)
}

func (s *ServiceAdapter) IsSnoozed() bool {
	// Enter the Service Snoze logic here
	return false
}

func (s *ServiceAdapter) Snooze(ctx context.Context, r client.Client) error {
	return nil
}

func (s *ServiceAdapter) Wake(ctx context.Context, r client.Client) error {
	return nil
}

func (s *ServiceAdapter) GetResourceType() string {
	return "service"
}
