package proxy

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func mockConfig() *Config {
	return &Config{
		Servers: []ServerConfig{
			{
				Listen: ":8080",
				Ssl:    false,
				Host:   "localhost:8080",
				Routes: []RouteConfig{
					{
						Match: RouteMatch{
							Path: "",
						},
						Proxy: ProxyConfig{
							Upstream: []string{"http://localhost:9001", "http://localhost:9002"},
							Strategy: StrategyConfig{
								Type: "weighted-round-robin",
								Config: map[string]interface{}{
									"weights": map[string]int{
										"http://localhost:9001": 5,
										"http://localhost:9002": 3,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func TestCreateProxyServers(t *testing.T) {
	// Create a mock ConfigLoader
	cl := &ConfigLoader{Config: mockConfig()}

	// Test CreateProxyServers
	proxyServers, err := cl.CreateProxyServers()
	assert.NoError(t, err, "CreateProxyServers should not return an error")

	// Validate that the proxy server is created
	server, exist := proxyServers[":8080"]
	assert.True(t, exist, "server not exist")
	assert.NotNil(t, server, "Proxy server should not be nil")
	assert.False(t, server.Ssl, "Expected SSL to be false")

	// Test the HTTP handler for the route
	req := httptest.NewRequest("GET", "http://localhost:8080", nil)
	req.Host = "localhost:8080"
	rec := httptest.NewRecorder()

	server.HttpHandler.ServeHTTP(rec, req)

	// Check response
	assert.Equal(t, http.StatusMovedPermanently, rec.Code, "Expected status code 301. proxy will redirect the request")
}

func TestCreateProxyServer(t *testing.T) {
	// arrange
	cl := &ConfigLoader{Config: mockConfig()}
	route := cl.Config.Servers[0].Routes[0]

	// act
	_, err := cl.CreateProxyServer(route)

	// assert
	assert.NoError(t, err, "CreateProxyServer should not be error")
}
