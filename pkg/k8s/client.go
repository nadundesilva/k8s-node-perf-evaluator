package k8s

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type client struct {
	clientset *kubernetes.Clientset
}

var _ Interface = (*client)(nil)

func NewFromKubeConfig(kubeConfigPath string) *client {
	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		panic(err.Error())
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	return &client{
		clientset: clientset,
	}
}

func (c *client) ListNodes(ctx context.Context) (*corev1.NodeList, error) {
	return c.clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
}
