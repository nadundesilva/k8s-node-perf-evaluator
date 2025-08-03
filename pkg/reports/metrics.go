package reports

import (
	"time"

	"github.com/nadundesilva/k8s-node-perf-evaluator/pkg/evaluator"
)

type TestSuiteResult struct {
	Name        string
	TestResults []*TestResult
}

type TestResult struct {
	NodeName           string
	AverageLatency     time.Duration
	FailedRequestCount int
	FailedPercentage   float64
}

func CalculateTestSuiteResults(testSuites []*evaluator.TestSuite) []*TestSuiteResult {
	testSuiteResults := []*TestSuiteResult{}
	for _, testSuite := range testSuites {
		testSuiteResults = append(testSuiteResults, calculateTestSuiteResult(testSuite))
	}
	return testSuiteResults
}

func calculateTestSuiteResult(testSuite *evaluator.TestSuite) *TestSuiteResult {
	testResults := []*TestResult{}
	for _, test := range testSuite.Tests {
		testResults = append(testResults, calculateTestResult(test))
	}
	return &TestSuiteResult{
		Name:        testSuite.Name,
		TestResults: testResults,
	}
}

func calculateTestResult(test *evaluator.Test) *TestResult {
	if test.TotalRequestsCount == 0 {
		return &TestResult{
			NodeName:           test.NodeName,
			AverageLatency:     0,
			FailedRequestCount: test.TotalFailedRequestsCount,
			FailedPercentage:   0,
		}
	}
	return &TestResult{
		NodeName:           test.NodeName,
		AverageLatency:     time.Duration(test.TotalLatency.Nanoseconds() / int64(test.TotalRequestsCount)),
		FailedRequestCount: test.TotalFailedRequestsCount,
		FailedPercentage:   float64(test.TotalFailedRequestsCount) / float64(test.TotalRequestsCount) * 100,
	}
}
