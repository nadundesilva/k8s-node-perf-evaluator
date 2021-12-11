package main

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
)

var servicePort = os.Getenv("SERVICE_PORT")

func main() {
	serviceMux := http.NewServeMux()

	serviceMux.Handle("/ping", http.HandlerFunc(handlePing))
	serviceMux.Handle("/cpu-intensive-task", http.HandlerFunc(handleCpuIntensiveTask))

	if servicePort == "" {
		servicePort = "8080"
	}
	listenAddress := fmt.Sprintf(":%s", servicePort)
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
