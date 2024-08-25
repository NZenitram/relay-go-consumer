package main

import (
	"fmt"
	"log"
	"net/http"
)

// startHTTPServer starts an HTTP server on port 8080
func StartHTTPServer() {
	http.HandleFunc("/health", healthCheckHandler)
	http.HandleFunc("/metrics", metricsHandler)

	log.Println("Starting HTTP server on port 8989")
	if err := http.ListenAndServe(":8989", nil); err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}
}

// healthCheckHandler handles health check requests
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "OK")
}

// metricsHandler handles metrics requests
func metricsHandler(w http.ResponseWriter, r *http.Request) {
	// Implement your metrics logic here
	fmt.Fprintln(w, "Metrics not implemented")
}
