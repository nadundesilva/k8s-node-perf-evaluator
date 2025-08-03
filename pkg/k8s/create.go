package k8s

import (
	"context"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

var createOptions = metav1.CreateOptions{}

func (c *client) CreateNamespace(ctx context.Context, namespace *corev1.Namespace) (*corev1.Namespace, error) {
	return c.clientset.CoreV1().Namespaces().Create(ctx, namespace, createOptions)
}

func (c *client) CreateDeployment(ctx context.Context, deployment *appsv1.Deployment) (*appsv1.Deployment, error) {
	d, err := c.clientset.AppsV1().Deployments(deployment.GetNamespace()).Create(ctx, deployment, createOptions)
	if err != nil {
		return nil, err
	}
	err = wait.PollUntilContextTimeout(ctx, time.Second, time.Minute, true, func(ctx context.Context) (bool, error) {
		namespace := deployment.GetNamespace()
		deploymentName := deployment.GetName()

		var deployment *appsv1.Deployment
		deployment, err = c.clientset.AppsV1().Deployments(namespace).Get(ctx, deploymentName, metav1.GetOptions{})
		if err != nil {
			return false, fmt.Errorf("failed to get deployment %s/%s: %w", namespace, deploymentName, err)
		}

		if deployment.Status.Replicas == deployment.Status.AvailableReplicas &&
			deployment.Status.UpdatedReplicas == deployment.Status.Replicas &&
			deployment.Status.ObservedGeneration >= deployment.Generation {
			return true, nil
		}

		return false, nil
	})
	if err != nil {
		return nil, err
	}
	return d, nil
}

func (c *client) CreateService(ctx context.Context, service *corev1.Service) (*corev1.Service, error) {
	return c.clientset.CoreV1().Services(service.GetNamespace()).Create(ctx, service, createOptions)
}

func (c *client) CreateIngress(ctx context.Context, ingress *networkingv1.Ingress) (*networkingv1.Ingress, error) {
	ing, err := c.clientset.NetworkingV1().Ingresses(ingress.GetNamespace()).Create(ctx, ingress, createOptions)
	if err != nil {
		return nil, err
	}
	err = wait.PollUntilContextTimeout(ctx, time.Second, time.Minute, true, func(ctx context.Context) (bool, error) {
		namespace := ingress.GetNamespace()
		ingressName := ingress.GetName()

		var ingress *networkingv1.Ingress
		ingress, err = c.clientset.NetworkingV1().Ingresses(namespace).Get(ctx, ingressName, metav1.GetOptions{})
		if err != nil {
			return false, fmt.Errorf("failed to get ingress %s/%s: %w", namespace, ingressName, err)
		}

		if len(ingress.Status.LoadBalancer.Ingress) > 0 {
			return true, nil
		}

		return false, nil
	})
	if err != nil {
		return nil, err
	}
	return ing, nil
}
