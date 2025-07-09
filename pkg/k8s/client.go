package k8s

import (
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type client struct {
	clientset *kubernetes.Clientset
}

var _ Interface = (*client)(nil)

func NewFromKubeConfig(kubeConfigPath string) (Interface, error) {
	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create k8s config: %w", err)
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create k8s client set: %w", err)
	}

	return &client{
		clientset: clientset,
	}, nil
}
