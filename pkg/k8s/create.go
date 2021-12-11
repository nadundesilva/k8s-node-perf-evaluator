package k8s

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var createOptions = metav1.CreateOptions{}

func (c *client) CreateNamespace(ctx context.Context, namespace *corev1.Namespace) (*corev1.Namespace, error) {
	return c.clientset.CoreV1().Namespaces().Create(ctx, namespace, createOptions)
}

func (c *client) CreateDeployment(ctx context.Context, deployment *appsv1.Deployment) (*appsv1.Deployment, error) {
	return c.clientset.AppsV1().Deployments(deployment.GetNamespace()).Create(ctx, deployment, createOptions)
}

func (c *client) CreateService(ctx context.Context, service *corev1.Service) (*corev1.Service, error) {
	return c.clientset.CoreV1().Services(service.GetNamespace()).Create(ctx, service, createOptions)
}

func (c *client) CreateIngress(ctx context.Context, ingress *networkingv1.Ingress) (*networkingv1.Ingress, error) {
	return c.clientset.NetworkingV1().Ingresses(ingress.GetNamespace()).Create(ctx, ingress, createOptions)
}
