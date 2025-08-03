package main

import (
	"context"
	"flag"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/nadundesilva/k8s-node-perf-evaluator/pkg/config"
	"github.com/nadundesilva/k8s-node-perf-evaluator/pkg/evaluator"
	"github.com/nadundesilva/k8s-node-perf-evaluator/pkg/reports"
	"github.com/nadundesilva/k8s-node-perf-evaluator/pkg/reports/writer"
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

	testRunResults := reports.CalculateTestSuiteResults(testRun)
	writerType := os.Getenv("TEST_RUNNER_REPORT_FORMAT")
	if writerType == "" {
		writerType = "text"
	}
	writer, err := writer.ResolveWriter(writerType)
	if err != nil {
		logger.Fatalw("Failed to resolve a writer", "error", err)
	}

	outputWriters := []io.Writer{os.Stdout}
	if outputFile := os.Getenv("TEST_RUNNER_REPORT_FILE"); outputFile != "" {
		err = os.MkdirAll(filepath.Dir(outputFile), 0755)
		if err != nil {
			logger.Fatalw("Failed to create output file parent directory", "error", err)
		}

		var fileWriter *os.File
		fileWriter, err = os.OpenFile(outputFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			logger.Fatalw("Failed to create output file", "error", err)
		}
		outputWriters = append(outputWriters, fileWriter)
	}

	err = writer.Write(testRunResults, io.MultiWriter(outputWriters...))
	if err != nil {
		logger.Fatalw("Failed to print report", "error", err)
	}
}
