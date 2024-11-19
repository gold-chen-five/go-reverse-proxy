package proxy

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// func TestLoadConfig(t *testing.T) {
// 	sampleConfig := `
// servers:
//   - listen: ":8080"
//     routes:
//       - match:
//           host: "example.com"
//           path: "/"
//         proxy:
//           upstream:
//             - "http://localhost:8081"
//             - "http://localhost:8082"
// `

// 	// create temporary yaml file to test
// 	file, err := os.CreateTemp("", "test_*.yaml")
// 	if err != nil {
// 		t.Fatalf("Fail to create temporary file %v", err)
// 	}
// 	defer os.Remove(file.Name())

// 	_, err = file.WriteString(sampleConfig)
// 	if err != nil {
// 		t.Fatalf("Fail to write string to temp file %v", err)
// 	}

// 	file.Close()

// 	// use tenp file to test
// 	cfg, err := LoadConfig(file.Name())
// 	if err != nil {
// 		t.Fatalf("Fail to load config file %v", err)
// 	}

// 	assert.Equal(t, ":8080", cfg.Servers[0].Listen, "Expected listen address to be :8080")
// 	assert.Equal(t, "example.com", cfg.Servers[0].Routes[0].Match.Host, "Expected host to be example.com")
// 	assert.Len(t, cfg.Servers[0].Routes[0].Proxy.Upstream, 2, "Expected 2 upstreams")
// }

func TestGetAllDomains(t *testing.T) {
	// mock config
	cfg := &Config{
		Servers: []ServerConfig{
			{
				Listen: ":8080",
				Host:   "example.com",
				Routes: []RouteConfig{
					{
						Match: RouteMatch{
							Path: "/",
						},
						Proxy: ProxyConfig{
							Upstream: []string{"http://localhost:8081"},
						},
					},
					{
						Match: RouteMatch{
							Path: "/api",
						},
						Proxy: ProxyConfig{
							Upstream: []string{"http://localhost:8082"},
						},
					},
				},
			},
			{
				Listen: ":8080",
				Host:   "test.com",
				Routes: []RouteConfig{
					{
						Match: RouteMatch{
							Path: "/",
						},
						Proxy: ProxyConfig{
							Upstream: []string{"http://localhost:8081"},
						},
					},
					{
						Match: RouteMatch{
							Path: "/api",
						},
						Proxy: ProxyConfig{
							Upstream: []string{"http://localhost:8082"},
						},
					},
				},
			},
		},
	}

	expected := []string{"example.com", "test.com"}
	actualDomains := cfg.GetAllDomains()

	assert.ElementsMatch(t, expected, actualDomains, "Domains fail. expected: %v, actual %v", expected, actualDomains)
}
