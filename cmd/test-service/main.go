package main

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"time"
)

var servicePort = os.Getenv("SERVICE_PORT")

func main() {
	serviceMux := http.NewServeMux()

	serviceMux.Handle("/ping", http.HandlerFunc(handlePing))
	serviceMux.Handle("/cpu-intensive-task", http.HandlerFunc(handleCPUIntensiveTask))

	if servicePort == "" {
		servicePort = "8080"
	}
	listenAddress := fmt.Sprintf(":%s", servicePort)
	server := &http.Server{
		Addr:              listenAddress,
		Handler:           serviceMux,
		ReadHeaderTimeout: time.Second * 5,
	}

	log.Printf("Starting listening on %s", listenAddress)
	err := server.ListenAndServe()
	if err != nil {
		log.Fatalf("Failed to listen to test service: %v", err)
	}
}

func handlePing(w http.ResponseWriter, _ *http.Request) {
	_, err := fmt.Fprintf(w, "{\"status\":\"success\"}")
	if err != nil {
		log.Printf("Failed to write response to ping: %v", err)
	}
}

func handleCPUIntensiveTask(w http.ResponseWriter, _ *http.Request) {
	var result float64
	for i := 0; i < int(math.Pow(10, 5)); i++ {
		result += math.Tan(float64(i)) * math.Atan(float64(i))
	}
	_, err := fmt.Fprintf(w, "{\"status\":\"success\",\"result\":\"%.2f\"}", result)
	if err != nil {
		log.Printf("Failed to write response to CPU intensive task: %v", err)
	}
}
