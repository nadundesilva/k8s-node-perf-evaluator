package writer

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/nadundesilva/k8s-node-perf-evaluator/pkg/reports"
)

type textWriter struct{}

var _ Writer = &textWriter{}

func (w *textWriter) Write(testSuiteResults []*reports.TestSuiteResult, output io.Writer) error {
	for _, testSuiteResult := range testSuiteResults {
		err := w.writeTitle(testSuiteResult.Name, output)
		if err != nil {
			return fmt.Errorf("failed to print title of console report %s: %w", testSuiteResult.Name, err)
		}

		w := tabwriter.NewWriter(output, 1, 1, 3, ' ', 0)
		_, err = fmt.Fprintln(w, "NODE\tAVERAGE LATENCY\tFAILED REQUESTS\t")
		if err != nil {
			return fmt.Errorf("failed to write header of console report %s: %w", testSuiteResult.Name, err)
		}
		for _, testResult := range testSuiteResult.TestResults {
			_, err = fmt.Fprintf(w, "%s\t%s\t%.2f%% (%d)\t\n", testResult.NodeName, testResult.AverageLatency, testResult.FailedPercentage, testResult.FailedRequestCount)
			if err != nil {
				return fmt.Errorf("failed to write row of console report %s: %w", testSuiteResult.Name, err)
			}
		}
		err = w.Flush()
		if err != nil {
			return fmt.Errorf("failed to flush console report %s: %w", testSuiteResult.Name, err)
		}
	}
	return nil
}

func (w *textWriter) writeTitle(title string, output io.Writer) error {
	verticalLine := strings.Repeat("=", len(title)+2)
	_, err := fmt.Fprintf(output, "\n%s\n %s \n%s\n\n", verticalLine, title, verticalLine)
	return err
}
