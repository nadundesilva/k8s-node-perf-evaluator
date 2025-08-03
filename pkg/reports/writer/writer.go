package writer

import (
	"fmt"
	"io"

	"github.com/nadundesilva/k8s-node-perf-evaluator/pkg/reports"
)

type Writer interface {
	Write(results []*reports.TestSuiteResult, output io.Writer) error
}

func ResolveWriter(writerType string) (Writer, error) {
	switch writerType {
	case "text":
		return &textWriter{}, nil
	case "json":
		return &jsonWriter{}, nil
	default:
		return nil, fmt.Errorf("unknown writer type: %s", writerType)
	}
}
