package writer

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/nadundesilva/k8s-node-perf-evaluator/pkg/reports"
)

type jsonWriter struct{}

var _ Writer = &jsonWriter{}

func (w *jsonWriter) Write(results []*reports.TestSuiteResult, output io.Writer) error {
	b, err := json.MarshalIndent(results, "", "\t")
	if err != nil {
		return fmt.Errorf("failed to convert test results to json: %+w", err)
	}
	_, err = output.Write(b)
	if err != nil {
		return fmt.Errorf("failed to write test results to output: %+w", err)
	}
	return err
}
