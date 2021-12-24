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

type TestService struct {
	Uuid     string
	NodeName string
	BaseUrl  string
}

type TestSuite struct {
	Name  string
	Tests []*Test
}

type Test struct {
	NodeName                 string
	TotalRequestsCount       int
	TotalFailedRequestsCount int
	TotalLatency             time.Duration
}

type status string

const (
	STATUS_SUCCESS status = "success"

	LOAD_TEST_WORKER_COUNT = 10
	ITERATION_COUNT        = 10
)

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

func (runner *testRunner) RunTest(ctx context.Context) ([]*TestSuite, error) {
	nodesList, err := runner.k8sClient.ListNodes(ctx, k8s.Selector{
		LabelSelector: runner.config.NodeSelector.LabelSelector,
		FieldSelector: runner.config.NodeSelector.FieldSelector,
	})
	if err != nil {
		runner.logger.Fatalw("failed to list the nodes in the cluster", "error", err)
	}

	testSuites := []*TestSuite{}
	runSuite := func(run func(ctx context.Context, testServices []*TestService) *TestSuite) error {
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
			return err
		}
		testSuites = append(testSuites, run(ctx, testServices))
		return nil
	}

	err = runSuite(func(ctx context.Context, testServices []*TestService) *TestSuite {
		return runner.runPingTest(ctx, testServices)
	})
	if err != nil {
		return testSuites, err
	}

	err = runSuite(func(ctx context.Context, testServices []*TestService) *TestSuite {
		return runner.runLoadTest(ctx, "CPU Intensive Load Test", "cpu-intensive-task", testServices)
	})
	if err != nil {
		return testSuites, err
	}

	return testSuites, nil
}

func (runner *testRunner) prepareTestServices(ctx context.Context, nodesList *corev1.NodeList) ([]*TestService, error) {
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

	testServices := []*TestService{}
	for _, node := range nodesList.Items {
		nodeName := node.GetObjectMeta().GetName()
		testService := &TestService{
			Uuid:     uuid.New().String(),
			NodeName: nodeName,
		}

		deployment, err := runner.k8sClient.CreateDeployment(ctx, runner.makeDeployment(*testService))
		if err != nil {
			runner.logger.Fatalw("failed to create deployment for node", "node", nodeName, "error", err)
		}

		service, err := runner.k8sClient.CreateService(ctx, runner.makeService(*testService))
		if err != nil {
			runner.logger.Fatalw("failed to create service for node", "node", nodeName, "error", err)
		}

		ingress, err := runner.k8sClient.CreateIngress(ctx, runner.makeIngress(*testService))
		if err != nil {
			runner.logger.Fatalw("failed to create ingress for node", "node", nodeName, "error", err)
		}
		testService.BaseUrl = runner.config.Ingress.ProtocolScheme + "://" + ingress.Spec.Rules[0].Host + ingress.Spec.Rules[0].HTTP.Paths[0].Path

		testServices = append(testServices, testService)
		runner.logger.Infow("created test service", "namespace", namespace.GetName(), "node", nodeName,
			"deployment", deployment.GetName(), "service", service.GetName(), "ingress", ingress.GetName())
	}
	return testServices, nil
}

func (runner *testRunner) runPingTest(ctx context.Context, testSvcs []*TestService) *TestSuite {
	name := "Ping Test"
	runner.logger.Infow("starting "+name, "services", len(testSvcs))
	testSuite := &TestSuite{
		Name:  name,
		Tests: []*Test{},
	}
	for _, testSvc := range testSvcs {
		test := &Test{
			NodeName: testSvc.NodeName,
		}
		url := makeUrl(testSvc.BaseUrl, "ping")

		for i := 0; i < ITERATION_COUNT; i++ {
			runner.runTestRequest(ctx, &url, test)
		}
		testSuite.Tests = append(testSuite.Tests, test)
	}
	runner.logger.Infow("completed " + name)
	return testSuite
}

func (runner *testRunner) runLoadTest(ctx context.Context, name string, reqPath string, testSvcs []*TestService) *TestSuite {
	runner.logger.Infow("starting CPU intensive load test", "services", len(testSvcs))
	testSuite := &TestSuite{
		Name:  name,
		Tests: []*Test{},
	}
	for _, testSvc := range testSvcs {
		url := makeUrl(testSvc.BaseUrl, reqPath)

		workerChannels := []chan int{}
		workerResultsChannels := []chan Test{}
		for i := 0; i < LOAD_TEST_WORKER_COUNT; i++ {
			workerChannel := make(chan int)
			workerResultsChannel := make(chan Test)
			go func() {
				workerTest := Test{
					NodeName: testSvc.NodeName,
				}
				workerChannel <- -1 // Signal ready to start test

				reqCount := <-workerChannel
				for i := 0; i < reqCount; i++ {
					runner.runTestRequest(ctx, &url, &workerTest)
				}
				workerChannel <- -1 // Signal test completed
				workerResultsChannel <- workerTest
			}()
			workerChannels = append(workerChannels, workerChannel)
			workerResultsChannels = append(workerResultsChannels, workerResultsChannel)
		}

		// Wait for workers to be ready
		for _, workerChannel := range workerChannels {
			<-workerChannel
		}

		// Send start reqCount
		for _, workerChannel := range workerChannels {
			workerChannel <- ITERATION_COUNT
		}

		// Wait for workers to complete
		for _, workerChannel := range workerChannels {
			<-workerChannel
		}

		// Merge results
		finalTest := &Test{
			NodeName:                 testSvc.NodeName,
			TotalRequestsCount:       0,
			TotalFailedRequestsCount: 0,
			TotalLatency:             0,
		}
		for _, workerChannel := range workerResultsChannels {
			test := <-workerChannel
			finalTest.TotalRequestsCount += test.TotalRequestsCount
			finalTest.TotalFailedRequestsCount += test.TotalFailedRequestsCount
			finalTest.TotalLatency += test.TotalLatency
		}
		testSuite.Tests = append(testSuite.Tests, finalTest)
	}
	runner.logger.Infow("completed CPU intensive load test")
	return testSuite
}

func (runner *testRunner) runTestRequest(ctx context.Context, url *string, test *Test) {
	reqStartTime := time.Now()
	resp, err := runner.httpClient.Get(*url)
	test.TotalLatency += time.Since(reqStartTime)
	if err != nil || resp.StatusCode != 200 {
		runner.logger.Fatalw("Error 1", "error", err)
		test.TotalFailedRequestsCount += 1
	} else {
		response := &testServiceResponse{}
		err = json.NewDecoder(resp.Body).Decode(response)
		if err != nil || response.Status != STATUS_SUCCESS {
			runner.logger.Fatalw("Error 2", "error", err)
			test.TotalFailedRequestsCount += 1
		}
	}
	test.TotalRequestsCount += 1
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
