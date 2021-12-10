package main

import (
	"fmt"
	"math"
	"net/http"
)

func main() {
	serviceMux := http.NewServeMux()

	serviceMux.Handle("/ping", http.HandlerFunc(handlePing))
	serviceMux.Handle("/cpu-intensive-task", http.HandlerFunc(handleCpuIntensiveTask))

	http.ListenAndServe(":8080", serviceMux)
}

func handlePing(w http.ResponseWriter, r *http.Request) {
	// Do nothing
}

func handleCpuIntensiveTask(w http.ResponseWriter, r *http.Request) {
	var result float64 = 0
	for i := 0; i < 1000; i++ {
		result += math.Tan(float64(i)) * math.Atan(float64(i))
	}
	fmt.Fprintf(w, "Result: %.2f", result)
}
