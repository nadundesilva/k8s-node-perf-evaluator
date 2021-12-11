package main

import (
	"context"
	"flag"

	"github.com/nadundesilva/k8s-node-perf-evaluator/pkg/config"
	"github.com/nadundesilva/k8s-node-perf-evaluator/pkg/evaluator"
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

	testRunner := evaluator.NewTestRunner(config, logger)
	testRunner.RunTest(ctx)
}
