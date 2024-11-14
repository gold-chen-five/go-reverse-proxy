package proxy

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// Test the creation of the proxy server
func TestNewProxyServer(t *testing.T) {
	// Mock valid upstream URLs
	upstreamURLs := []string{
		"http://localhost:8080",
		"http://localhost:8081",
	}

	proxyServer, err := NewProxyServer(upstreamURLs)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(proxyServer.LoadBalancer.servers) != 2 {
		t.Fatalf("Expected 2 upstream servers, got %d", len(proxyServer.LoadBalancer.servers))
	}
}

// Test the ServeHTTP functionality
func TestProxyServer_ServeHTTP(t *testing.T) {
	// Create a test server to act as an upstream server
	testUpstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer testUpstream.Close()

	// Create a proxy server with the test upstream
	proxyServer, err := NewProxyServer([]string{testUpstream.URL})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Create a test HTTP request
	req := httptest.NewRequest("GET", "http://localhost", nil)
	rr := httptest.NewRecorder()

	// Call the ServeHTTP method
	proxyServer.ServeHTTP(rr, req)

	// Check that the response status code is 200 OK
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, status)
	}
}

// Mock the health check
func TestProxyServer_healthCheck(t *testing.T) {
	// Create a test upstream server that returns a 500 error (simulate failing health check)
	failingServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}))
	defer failingServer.Close()

	// Create a proxy server with the failing upstream server
	proxyServer, err := NewProxyServer([]string{failingServer.URL})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Override the config to reduce the health check interval and timeout for faster testing
	proxyServer.Config.HealthCheckInterval = 100 * time.Millisecond
	proxyServer.Config.Timeout = 100 * time.Millisecond
	proxyServer.Config.MaxFailCount = 1

	// Simulate running the health check by calling it directly
	go proxyServer.healthCheck()

	// Wait for health check to complete
	time.Sleep(500 * time.Millisecond)

	// Check if the server is marked as down
	server := proxyServer.LoadBalancer.servers[0]
	if server.Alive {
		t.Errorf("Expected server to be marked as down, but it is alive")
	}
}
