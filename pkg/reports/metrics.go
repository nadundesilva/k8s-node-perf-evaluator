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
	if test.TotalRequestsCount == 0 {
		return &metrics{
			averageLatency:     0,
			failedRequestCount: test.TotalFailedRequestsCount,
			failedPercentage:   0,
		}
	}
	return &metrics{
		averageLatency:     time.Duration(test.TotalLatency.Nanoseconds() / int64(test.TotalRequestsCount)),
		failedRequestCount: test.TotalFailedRequestsCount,
		failedPercentage:   float64(test.TotalFailedRequestsCount) / float64(test.TotalRequestsCount) * 100,
	}
}
