package k8s

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var getOptions = metav1.GetOptions{}

func (c *client) GetNamespace(ctx context.Context, name string) (*corev1.Namespace, error) {
	namespace, err := c.clientset.CoreV1().Namespaces().Get(ctx, name, getOptions)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		} else {
			return nil, err
		}
	}
	return namespace, nil
}
