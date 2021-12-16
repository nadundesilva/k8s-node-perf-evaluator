package evaluator

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nadundesilva/k8s-node-perf-evaluator/pkg/config"
	"github.com/nadundesilva/k8s-node-perf-evaluator/pkg/k8s"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
)

type testRunner struct {
	config     *config.Config
	logger     *zap.SugaredLogger
	k8sClient  k8s.Interface
	httpClient *http.Client
}

type testService struct {
	Uuid        string
	NodeName    string
	BaseUrl     string
	TestResults *testResults
}

type testResults struct {
	PingTest             *testRun
	CpuIntensiveTaskTest *testRun
}

type testRun struct {
	TotalRequestsCount       int
	TotalFailedRequestsCount int
	TotalLatency             time.Duration
}

type status string

const statusSuccess status = "success"

type testServiceResponse struct {
	Status status
}

func NewTestRunner(config *config.Config, logger *zap.SugaredLogger) *testRunner {
	return &testRunner{
		config:    config,
		logger:    logger,
		k8sClient: k8s.NewFromKubeConfig(config.KubeConfig),
		httpClient: &http.Client{
			Timeout: time.Minute,
		},
	}
}

func (runner *testRunner) RunTest(ctx context.Context) (*map[string]testService, error) {
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

	testServices, err := runner.prepareTestServices(ctx, nodesList)
	defer func() {
		err := runner.cleanupTestServices(ctx)
		if err != nil {
			runner.logger.Warnw("failed to cleanup test services", "namespace", runner.config.Namespace, "error", err)
		}
		runner.logger.Info("cleaned up all resource", "namespace", runner.config.Namespace)
	}()
	if err != nil {
		return nil, err
	}
	runner.runPingTest(ctx, testServices)
	return testServices, nil
}

func (runner *testRunner) prepareTestServices(ctx context.Context, nodesList *corev1.NodeList) (*map[string]testService, error) {
	namespace, err := runner.k8sClient.GetNamespace(ctx, runner.config.Namespace)
	if err != nil {
		runner.logger.Fatalw("failed to check if the test services namespace existed", "namespace", runner.config.Namespace, "error", err)
	}
	if namespace != nil {
		err = runner.k8sClient.DeleteNamespace(ctx, namespace.GetName())
		if err != nil {
			return nil, err
		}
		runner.logger.Infow("waiting for namespace deletion to complete", "namespace", namespace.GetName())
		err := runner.k8sClient.WaitForNamespaceDeletion(ctx, namespace.GetName())
		if err != nil {
			return nil, err
		}
		runner.logger.Infow("deleted existing existing test services namespace", "namespace", namespace.GetName())
	}

	namespace, err = runner.k8sClient.CreateNamespace(ctx, runner.makeNamespace(runner.config.Namespace))
	if err != nil {
		runner.logger.Fatalw("failed to create test services namespace", "namespace", runner.config.Namespace, "error", err)
	}
	runner.logger.Infow("created test services namespace", "namespace", namespace.GetName())

	testServices := map[string]testService{}
	for _, node := range nodesList.Items {
		nodeName := node.GetObjectMeta().GetName()
		testService := testService{
			Uuid:        uuid.New().String(),
			NodeName: nodeName,
			TestResults: &testResults{},
		}

		deployment, err := runner.k8sClient.CreateDeployment(ctx, runner.makeDeployment(testService))
		if err != nil {
			runner.logger.Fatalw("failed to create deployment for node", "node", nodeName, "error", err)
		}

		service, err := runner.k8sClient.CreateService(ctx, runner.makeService(testService))
		if err != nil {
			runner.logger.Fatalw("failed to create service for node", "node", nodeName, "error", err)
		}

		ingress, err := runner.k8sClient.CreateIngress(ctx, runner.makeIngress(testService))
		if err != nil {
			runner.logger.Fatalw("failed to create ingress for node", "node", nodeName, "error", err)
		}
		testService.BaseUrl = ingress.Spec.Rules[0].Host + ingress.Spec.Rules[0].HTTP.Paths[0].Path

		testServices[nodeName] = testService
		runner.logger.Infow("created test service", "namespace", namespace.GetName(), "node", nodeName,
			"deployment", deployment.GetName(), "service", service.GetName(), "ingress", ingress.GetName())
	}
	return &testServices, nil
}

func (runner *testRunner) runPingTest(ctx context.Context, testSvcs *map[string]testService) {
	runner.logger.Infow("starting ping test", "services", len(*testSvcs))
	for _, testSvc := range *testSvcs {
		testRun := &testRun{}
		url := makeUrl(testSvc.BaseUrl, "ping")

		for i := 0; i < 1000; i++ {
			runner.runTestRequest(ctx, &url, testRun)
		}
		testSvc.TestResults.PingTest = testRun
	}
	runner.logger.Infow("completed ping test")
}

func (runner *testRunner) runTestRequest(ctx context.Context, url *string, testRun *testRun) {
	reqStartTime := time.Now()
	resp, err := runner.httpClient.Get(*url)
	testRun.TotalLatency += time.Since(reqStartTime)
	if err != nil || resp.StatusCode != 200 {
		testRun.TotalFailedRequestsCount += 1
	} else {
		response := &testServiceResponse{}
		err = json.NewDecoder(resp.Body).Decode(response)
		if err != nil || response.Status != statusSuccess {
			testRun.TotalFailedRequestsCount += 1
		}
	}
	testRun.TotalRequestsCount += 1
}

func (runner *testRunner) cleanupTestServices(ctx context.Context) error {
	return runner.k8sClient.DeleteNamespace(ctx, runner.config.Namespace)
}

func makeUrl(baseUrl, path string) string {
	url := baseUrl
	if !strings.HasSuffix(baseUrl, "/") {
		url += "/"
	}
	return url + path
}
