package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/nadundesilva/k8s-node-perf-evaluator/pkg/config"
	"github.com/nadundesilva/k8s-node-perf-evaluator/pkg/evaluator"
	"github.com/nadundesilva/k8s-node-perf-evaluator/pkg/reports"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	ctx := context.Background()

	zapConf := zap.NewDevelopmentConfig()
	zapConf.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	zapLogger, err := zapConf.Build()
	if err != nil {
		log.Printf("Failed to create logger: %v", err)
		return
	}
	defer func() {
		err = zapLogger.Sync()
		if err != nil {
			log.Printf("Failed to sync logger: %v", err)
		}
	}()
	logger := zapLogger.Sugar()
	logger.Info("Starting Node Performance Evaluator")

	configFile := flag.String("config", "config.yaml", "(optional) absolute path to the config file")
	flag.Parse()

	config, err := config.Read(*configFile)
	if err != nil {
		logger.Fatalw("failed to read Config", "error", err)
	}

	testRunner, err := evaluator.NewTestRunner(config, logger)
	if err != nil {
		logger.Fatalw("failed to create test runner", "error", err)
	}
	testRun, err := testRunner.RunTest(ctx)
	if err != nil {
		logger.Fatalw("Failed to run test", "error", err)
	}
	err = reports.Print(testRun, os.Stdout)
	if err != nil {
		logger.Fatalw("Failed to print report", "error", err)
	}
}
