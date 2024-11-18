package proxy

// import "testing"

// func mockConfig() *Config {
// 	return &Config{
// 		Servers: []ServerConfig{
// 			{
// 				Listen: ":8080",
// 				Ssl:    false,
// 				Routes: []RouteConfig{
// 					{
// 						Match: RouteMatch{
// 							Host: "testproxy.local",
// 							Path: "/api",
// 						},
// 						Proxy: ProxyConfig{
// 							Upstream: []string{"http://localhost:9001", "http://localhost:9002"},
// 							Strategy: StrategyConfig{
// 								Type: "weighted-round-robin",
// 								Config: map[string]interface{}{
// 									"weights": map[string]int{
// 										"http://localhost:9001": 5,
// 										"http://localhost:9002": 3,
// 									},
// 								},
// 							},
// 						},
// 					},
// 				},
// 			},
// 		},
// 	}
// }

// func TestCreateProxyServers(t *testing.T) {

// }
