package reports

import (
	"time"

	"github.com/nadundesilva/k8s-node-perf-evaluator/pkg/evaluator"
)

type metrics struct {
	averageLatency     time.Duration
	failedRequestCount int
	failedPercentage   float64
}

func calculateMetrics(test *evaluator.Test) *metrics {
	return &metrics{
		averageLatency:     time.Duration(test.TotalLatency.Nanoseconds() / int64(test.TotalRequestsCount)),
		failedRequestCount: test.TotalFailedRequestsCount,
		failedPercentage:   float64(test.TotalFailedRequestsCount/test.TotalRequestsCount) * 100,
	}
}
