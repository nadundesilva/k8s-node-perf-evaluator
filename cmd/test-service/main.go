package main

import (
	"fmt"
	"log"
	"math"
	"net/http"
)

const SERVER_PORT = 8080

func main() {
	serviceMux := http.NewServeMux()

	serviceMux.Handle("/ping", http.HandlerFunc(handlePing))
	serviceMux.Handle("/cpu-intensive-task", http.HandlerFunc(handleCpuIntensiveTask))

	listenAddress := fmt.Sprintf(":%d", SERVER_PORT)
	log.Printf("Starting listening on %s", listenAddress)
	http.ListenAndServe(listenAddress, serviceMux)
}

func handlePing(w http.ResponseWriter, r *http.Request) {
	// Do nothing
}

func handleCpuIntensiveTask(w http.ResponseWriter, r *http.Request) {
	var result float64 = 0
	for i := 0; i < int(math.Pow(10, 8)); i++ {
		result += math.Tan(float64(i)) * math.Atan(float64(i))
	}
	fmt.Fprintf(w, "Result: %.2f", result)
}
