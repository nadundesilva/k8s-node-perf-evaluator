package reports

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/nadundesilva/k8s-node-perf-evaluator/pkg/evaluator"
)

func Print(testSuites []*evaluator.TestSuite, output io.Writer) {
	for _, testSuite := range testSuites {
		printTitle(testSuite.Name, output)

		w := tabwriter.NewWriter(output, 1, 1, 3, ' ', 0)
		fmt.Fprintln(w, "NODE\tAVERAGE LATENCY\tFAILED REQUESTS\t")
		for _, test := range testSuite.Tests {
			metrics := calculateMetrics(test)
			fmt.Fprintf(w, "%s\t%s\t%.2f%% (%d)\t\n", test.NodeName, metrics.averageLatency, metrics.failedPercentage, metrics.failedRequestCount)
		}
		w.Flush()
	}
}

func printTitle(title string, output io.Writer) {
	verticalLine := strings.Repeat("=", len(title)+2)
	fmt.Fprintf(output, "\n%s\n %s \n%s\n\n", verticalLine, title, verticalLine)
}
