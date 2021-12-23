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
	err := http.ListenAndServe(listenAddress, serviceMux)
	if err != nil {
		log.Fatalf("Failed to listen to test service: %v", err)
	}
}

func handlePing(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "{\"status\":\"success\"}")
}

func handleCpuIntensiveTask(w http.ResponseWriter, r *http.Request) {
	var result float64 = 0
	for i := 0; i < int(math.Pow(10, 8)); i++ {
		result += math.Tan(float64(i)) * math.Atan(float64(i))
	}
	fmt.Fprintf(w, "{\"status\":\"success\",\"result\":\"%.2f\"}", result)
}
