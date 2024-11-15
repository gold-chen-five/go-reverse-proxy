package proxy

import (
	"log"
	"net/http"
)

type ConfigLoader struct {
	Config *Config
}

func NewConfigLoader(filename string) (*ConfigLoader, error) {
	cfg, err := LoadConfig(filename)
	if err != nil {
		return nil, err
	}

	return &ConfigLoader{Config: cfg}, nil
}

func (cl *ConfigLoader) CreateProxyServers() {
	// Create a router to handle different routes
	mux := http.NewServeMux()

	for _, server := range cl.Config.Servers {
		for _, route := range server.Routes {
			px, err := cl.CreateProxyServer(route)
			if err != nil {
				log.Fatal(err)
			}

			mux.HandleFunc(route.Match.Path, func(w http.ResponseWriter, r *http.Request) {
				// check the host header
				if r.Host == route.Match.Host {
					px.ServeHTTP(w, r)
				} else {
					http.NotFound(w, r)
				}
			})
		}
	}

	return mux
}

func (cl *ConfigLoader) CreateProxyServer(route RouteConfig) (*ProxyServer, error) {
	// 創建代理服務器
	px, err := NewProxyServer(route.Proxy.Upstream)
	if err != nil {
		return nil, err
	}

	// set strategy
	if route.Proxy.Strategy.Type != "" {
		px.LoadBalancer.UpdateStrategy(route.Proxy.Strategy.Type)

		// set weight for weightRR
		if route.Proxy.Strategy.Type == WeightedRR {
			if weights, ok := route.Proxy.Strategy.Config["weights"].(map[string]interface{}); ok {
				for url, weight := range weights {
					if w, ok := weight.(int); ok {
						px.LoadBalancer.SetServerWeight(url, int32(w))
					}
				}
			}

		}
	}

	return px, nil
}
