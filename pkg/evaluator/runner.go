package evaluator

import (
	"context"

	"github.com/nadundesilva/k8s-node-perf-evaluator/pkg/config"
	"github.com/nadundesilva/k8s-node-perf-evaluator/pkg/k8s"
	"go.uber.org/zap"
)

type testRunner struct {
	config    *config.Config
	logger    *zap.SugaredLogger
	k8sClient k8s.Interface
}

func NewTestRunner(config *config.Config, logger *zap.SugaredLogger) *testRunner {
	return &testRunner{
		config:    config,
		logger:    logger,
		k8sClient: k8s.NewFromKubeConfig(config.KubeConfig),
	}
}

func (runner *testRunner) RunTest(ctx context.Context) {
	nodesList, err := runner.k8sClient.ListNodes(ctx, runner.config.NodeSelector)
	if err != nil {
		runner.logger.Fatalw("failed to list the nodes in the cluster", "error", err)
	}

	nodeNames := []string{}
	for _, node := range nodesList.Items {
		nodeNames = append(nodeNames, node.GetObjectMeta().GetName())
	}
	runner.logger.Infow("resolved available nodes to be tested", "nodes", nodeNames)
}
