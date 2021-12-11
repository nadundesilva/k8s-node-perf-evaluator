package k8s

import (
	"context"

	"github.com/nadundesilva/k8s-node-perf-evaluator/pkg/config"
	corev1 "k8s.io/api/core/v1"
)

type Interface interface {
	ListNodes(ctx context.Context, selector config.Selector) (*corev1.NodeList, error)
}
