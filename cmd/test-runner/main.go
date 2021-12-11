package main

import (
	"context"
	"flag"
	"path/filepath"

	"github.com/nadundesilva/k8s-node-perf-evaluator/pkg/k8s"
	"go.uber.org/zap"
	"k8s.io/client-go/util/homedir"
)

func main() {
	zapLogger, _ := zap.NewProduction()
	defer zapLogger.Sync()
	logger := zapLogger.Sugar()
	logger.Info("Starting Node Performance Evaluator")
	ctx := context.Background()

	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	k8sClient := k8s.NewFromKubeConfig(*kubeconfig)
	nodesList, err := k8sClient.ListNodes(ctx)
	if err != nil {
		logger.Errorw("Failed to list the nodes in the cluster", "error", err)
	}

	for _, node := range nodesList.Items {
		logger.Infof("Node: %s", node.GetObjectMeta().GetName())
	}
}
