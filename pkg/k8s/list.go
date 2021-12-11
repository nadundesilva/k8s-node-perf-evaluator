package k8s

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *client) ListNodes(ctx context.Context, selector Selector) (*corev1.NodeList, error) {
	return c.clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{
		LabelSelector: selector.LabelSelector,
		FieldSelector: selector.FieldSelector,
	})
}

