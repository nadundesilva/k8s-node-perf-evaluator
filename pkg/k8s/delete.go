package k8s

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var DELETE_PROPAGATION = metav1.DeletePropagationBackground

func (c *client) DeleteNamespace(ctx context.Context, name string) (error) {
	return c.clientset.CoreV1().Namespaces().Delete(ctx, name, metav1.DeleteOptions{
		PropagationPolicy: &DELETE_PROPAGATION,
	})
}