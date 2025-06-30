package reports

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/nadundesilva/k8s-node-perf-evaluator/pkg/evaluator"
)

func Print(testSuites []*evaluator.TestSuite, output io.Writer) error {
	for _, testSuite := range testSuites {
		err := printTitle(testSuite.Name, output)
		if err != nil {
			return fmt.Errorf("failed to print title of console report %s: %w", testSuite.Name, err)
		}

		w := tabwriter.NewWriter(output, 1, 1, 3, ' ', 0)
		_, err = fmt.Fprintln(w, "NODE\tAVERAGE LATENCY\tFAILED REQUESTS\t")
		if err != nil {
			return fmt.Errorf("failed to write header of console report %s: %w", testSuite.Name, err)
		}
		for _, test := range testSuite.Tests {
			metrics := calculateMetrics(test)
			_, err = fmt.Fprintf(w, "%s\t%s\t%.2f%% (%d)\t\n", test.NodeName, metrics.averageLatency, metrics.failedPercentage, metrics.failedRequestCount)
			if err != nil {
				return fmt.Errorf("failed to write row of console report %s: %w", testSuite.Name, err)
			}
		}
		err = w.Flush()
		if err != nil {
			return fmt.Errorf("failed to flush console report %s: %w", testSuite.Name, err)
		}
	}
	return nil
}

func printTitle(title string, output io.Writer) error {
	verticalLine := strings.Repeat("=", len(title)+2)
	_, err := fmt.Fprintf(output, "\n%s\n %s \n%s\n\n", verticalLine, title, verticalLine)
	return err
}
