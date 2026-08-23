package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand/v2"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var r *rand.Rand

// OrderRequest represents the request body for the order persistence API
type OrderRequest struct {
	UserID    string `json:"userId"`
	Price     int    `json:"price"`
	Status    string `json:"status"`
	RequestID string `json:"requestId"`
}

func main() {
	log.Println("Starting Spacebase Dashboard Simulator...")

	// 1. Define command line flags with environment variable fallbacks
	defaultURL := getEnv("API_URL", "http://middleware:8080/api/dashboard/orders")
	defaultInterval := 1000
	if intervalStr, ok := os.LookupEnv("INTERVAL_MS"); ok {
		if val, err := strconv.Atoi(intervalStr); err == nil {
			defaultInterval = val
		}
	}
	defaultCount := 0 // 0 means infinite
	if countStr, ok := os.LookupEnv("COUNT"); ok {
		if val, err := strconv.Atoi(countStr); err == nil {
			defaultCount = val
		}
	}

	apiURL := flag.String("url", defaultURL, "Target API URL for order persistence")
	intervalMS := flag.Int("interval", defaultInterval, "Interval between requests in milliseconds")
	count := flag.Int("count", defaultCount, "Number of requests to send before exiting (0 for infinite)")
	enabledStr := flag.String("enabled", getEnv("ENABLED", "true"), "Set to 'false' to disable simulator and sleep indefinitely")

	flag.Parse()

	// 2. Check if disabled
	if *enabledStr == "false" {
		log.Println("Simulator is disabled via -enabled=false (or ENABLED=false). Going to sleep indefinitely.")
		select {} // Sleep forever
	}

	log.Printf("Configuration:")
	log.Printf("  - API_URL: %s", *apiURL)
	log.Printf("  - INTERVAL_MS: %d ms", *intervalMS)
	if *count > 0 {
		log.Printf("  - LIMIT: %d requests", *count)
	} else {
		log.Printf("  - LIMIT: Infinite requests")
	}

	// Seed the random number generator using Go v2 rand PCG source
	r = rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), 0))

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// 3. Health check / Wait for middleware API to become ready
	waitForAPI(client, *apiURL)

	log.Println("API is ready. Starting to send simulated orders...")

	// 4. Main simulation loop
	ticker := time.NewTicker(time.Duration(*intervalMS) * time.Millisecond)
	defer ticker.Stop()

	sentCount := 0
	for {
		select {
		case <-ticker.C:
			order := generateRandomOrder()
			sendOrder(client, *apiURL, order)
			sentCount++

			if *count > 0 && sentCount >= *count {
				log.Printf("Successfully sent all requested %d orders. Exiting simulator.", *count)
				return
			}
		}
	}
}

// getEnv reads an environment variable or returns a default value
func getEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}

// waitForAPI attempts to ping the health endpoint until it succeeds, blocking until it's ready.
func waitForAPI(client *http.Client, apiURL string) {
	// apiURLが "http://middleware:8080/api/dashboard/orders" などの場合、
	// "/api/" より前のベース部分を取り出し、"/health" を結合してヘルスチェックURLを自動構成します。
	healthURL := "http://middleware:8080/health"
	if idx := strings.Index(apiURL, "/api/"); idx != -1 {
		healthURL = apiURL[:idx] + "/health"
	}

	for {
		log.Printf("Checking connectivity to Health Endpoint: %s ...", healthURL)
		resp, err := client.Get(healthURL)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				log.Println("Successfully connected to API (Health Check OK).")
				return
			}
			log.Printf("API returned non-200 status for health check: %d", resp.StatusCode)
		} else {
			log.Printf("API is not ready yet (%v). Retrying in 5 seconds...", err)
		}

		time.Sleep(5 * time.Second)
	}
}

// generateRandomOrder creates a pseudo-random OrderRequest
func generateRandomOrder() OrderRequest {
	// Status weighted choice:
	// completed: 70%, order: 15%, reject: 10%, error: 5%
	status := "completed"
	roll := r.IntN(100)
	if roll < 5 {
		status = "error"
	} else if roll < 15 {
		status = "reject"
	} else if roll < 30 {
		status = "order"
	}

	// Price: 100 to 2000 in steps of 50
	steps := r.IntN(39) // 0 to 38
	price := 100 + (steps * 50)

	// Random userID from user_1 to user_100
	userID := fmt.Sprintf("user_%d", r.IntN(100)+1)

	// Request ID with random suffix (using Uint64 for v2 instead of Int63)
	requestID := fmt.Sprintf("req_%x", r.Uint64())

	return OrderRequest{
		UserID:    userID,
		Price:     price,
		Status:    status,
		RequestID: requestID,
	}
}

// sendOrder POSTs the order data to the target API
func sendOrder(client *http.Client, url string, order OrderRequest) {
	data, err := json.Marshal(order)
	if err != nil {
		log.Printf("Failed to marshal order: %v", err)
		return
	}

	resp, err := client.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		log.Printf("Failed to send order (User: %s, Price: %d, Status: %s): %v", order.UserID, order.Price, order.Status, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		log.Printf("Successfully sent order: User=%-10s, Price=%-4d, Status=%-10s, ReqID=%s", order.UserID, order.Price, order.Status, order.RequestID)
	} else {
		log.Printf("API returned non-2xx status: %d (Order: User=%s, Price=%d)", resp.StatusCode, order.UserID, order.Price)
	}
}
