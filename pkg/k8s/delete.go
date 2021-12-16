package k8s

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var deletePropagation = metav1.DeletePropagationForeground
var deleteOptions = metav1.DeleteOptions{
	PropagationPolicy: &deletePropagation,
}

func (c *client) DeleteNamespace(ctx context.Context, name string) (error) {
	return c.clientset.CoreV1().Namespaces().Delete(ctx, name, deleteOptions)
}
