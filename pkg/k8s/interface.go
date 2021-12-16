package k8s

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
)

type Interface interface {
	CreateNamespace(ctx context.Context, namespace *corev1.Namespace) (*corev1.Namespace, error)
	CreateDeployment(ctx context.Context, deployment *appsv1.Deployment) (*appsv1.Deployment, error)
	CreateService(ctx context.Context, service *corev1.Service) (*corev1.Service, error)
	CreateIngress(ctx context.Context, ingress *networkingv1.Ingress) (*networkingv1.Ingress, error)

	ListNodes(ctx context.Context, selector Selector) (*corev1.NodeList, error)

	GetNamespace(ctx context.Context, name string) (*corev1.Namespace, error)

	DeleteNamespace(ctx context.Context, name string) (error)

	WaitForNamespaceDeletion(ctx context.Context, name string) error
}
