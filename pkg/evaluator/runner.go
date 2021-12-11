package evaluator

import (
	"context"

	"github.com/nadundesilva/k8s-node-perf-evaluator/pkg/config"
	"github.com/nadundesilva/k8s-node-perf-evaluator/pkg/k8s"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
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
	nodesList, err := runner.k8sClient.ListNodes(ctx, k8s.Selector{
		LabelSelector: runner.config.NodeSelector.LabelSelector,
		FieldSelector: runner.config.NodeSelector.FieldSelector,
	})
	if err != nil {
		runner.logger.Fatalw("failed to list the nodes in the cluster", "error", err)
	}

	nodeNames := []string{}
	for _, node := range nodesList.Items {
		nodeNames = append(nodeNames, node.GetObjectMeta().GetName())
	}
	runner.logger.Infow("resolved available nodes to be tested", "nodes", nodeNames)

	runner.prepareTestServices(ctx, nodesList)
	defer func() {
		err := runner.cleanupTestServices(ctx)
		if err != nil {
			runner.logger.Warnw("failed to cleanup test services", "namespace", runner.config.Namespace, "error", err)
		}
		runner.logger.Info("cleaned up all resource", "namespace", runner.config.Namespace)
	}()
}

func (runner *testRunner) prepareTestServices(ctx context.Context, nodesList *corev1.NodeList) {
	namespace, err := runner.k8sClient.CreateNamespace(ctx, runner.makeNamespace(runner.config.Namespace))
	if err != nil {
		runner.logger.Fatalw("failed to create test services namespace", "namespace", runner.config.Namespace)
	}
	runner.logger.Infow("created test services namespace", "namespace", namespace.GetName())

	for _, node := range nodesList.Items {
		nodeName := node.GetObjectMeta().GetName()
		deployment, err := runner.k8sClient.CreateDeployment(ctx, runner.makeDeployment(nodeName))
		if err != nil {
			runner.logger.Fatalw("failed to create deployment for node", "node", nodeName)
		}

		runner.logger.Infow("created test service", "namespace", namespace.GetName(), "node", nodeName,
			"deployment", deployment.GetName())
	}
}

func (runner *testRunner) cleanupTestServices(ctx context.Context) error {
	return runner.k8sClient.DeleteNamespace(ctx, runner.config.Namespace)
}
