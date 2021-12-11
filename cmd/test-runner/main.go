package main

import (
	"context"
	"flag"

	"github.com/nadundesilva/k8s-node-perf-evaluator/pkg/config"
	"github.com/nadundesilva/k8s-node-perf-evaluator/pkg/k8s"
	"go.uber.org/zap"
)

func main() {
	ctx := context.Background()

	zapLogger, _ := zap.NewProduction()
	defer zapLogger.Sync()
	logger := zapLogger.Sugar()
	logger.Info("Starting Node Performance Evaluator")

	configFile := flag.String("config", "config.yaml", "(optional) absolute path to the config file")
	flag.Parse()

	config, err := config.Read(*configFile)
	if err != nil {
		logger.Fatalw("failed to read Config", "error", err)
	}

	k8sClient := k8s.NewFromKubeConfig(config.KubeConfig)
	nodesList, err := k8sClient.ListNodes(ctx)
	if err != nil {
		logger.Errorw("Failed to list the nodes in the cluster", "error", err)
	}

	for _, node := range nodesList.Items {
		logger.Infof("Node: %s", node.GetObjectMeta().GetName())
	}
}
