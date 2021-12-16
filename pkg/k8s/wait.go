package k8s

import (
	"context"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)


func (c *client) WaitForNamespaceDeletion(ctx context.Context, name string) error {
	w, err := c.clientset.CoreV1().Namespaces().Watch(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	for true {
		select {
		case event := <-w.ResultChan():
			if event.Type == watch.Deleted {
				return nil
			}
		case <-time.After(30 * time.Second):
			return fmt.Errorf("timed out waiting for namespace deletion: %v", name)
		}
	}
	return nil
}
