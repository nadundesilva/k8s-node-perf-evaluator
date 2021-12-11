package k8s

import (
	"context"

	corev1 "k8s.io/api/core/v1"
)

type Interface interface {
	ListNodes(ctx context.Context) (*corev1.NodeList, error)
}
