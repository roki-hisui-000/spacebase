package gateway

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"
)

// AsyncGateway is an implementation of Gateway that publishes events to a message broker.
type AsyncGateway struct {
	brokerURL string
}

// NewAsyncGateway creates a new AsyncGateway based on environment variables.
func NewAsyncGateway() *AsyncGateway {
	url := os.Getenv("BROKER_URL")
	if url == "" {
		url = "localhost:9092" // default broker address
	}
	return &AsyncGateway{brokerURL: url}
}

// ServeHTTP receives HTTP requests and publishes an event asynchronously.
func (g *AsyncGateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("[HTTP Async Request] %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)

	// TODO: Parse JSON request body, map to event structure.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_ = ctx

	// TODO: Publish event to broker at g.brokerURL with payload from r.Body.
	// For now, respond Accepted.
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("Event accepted for async processing"))
}
