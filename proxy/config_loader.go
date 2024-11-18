package proxy

import (
	"net/http"
)

type ConfigLoader struct {
	Config *Config
}

type TProxyServer struct {
	Ssl         bool
	HttpHandler http.Handler
}

func NewConfigLoader(filename string) (*ConfigLoader, error) {
	cfg, err := LoadConfig(filename)
	if err != nil {
		return nil, err
	}

	return &ConfigLoader{Config: cfg}, nil
}

func (cl *ConfigLoader) CreateProxyServers() (map[string]*TProxyServer, error) {
	proxyServers := make(map[string]*TProxyServer)

	for _, server := range cl.Config.Servers {
		// Create a router to handle different routes
		mux := http.NewServeMux()

		for _, route := range server.Routes {
			px, err := cl.CreateProxyServer(route)
			if err != nil {
				return nil, err
			}

			mux.HandleFunc(route.Match.Path, func(w http.ResponseWriter, r *http.Request) {
				// check the host header
				if r.Host == route.Match.Host {
					if len(r.URL.Path) >= len(route.Match.Path) && r.URL.Path[:len(route.Match.Path)] == route.Match.Path {
						r.URL.Path = r.URL.Path[len(route.Match.Path):]
						px.ServeHTTP(w, r)
						return
					}
				}
				http.NotFound(w, r)

			})
			proxyServers[server.Listen] = &TProxyServer{
				Ssl:         server.Ssl,
				HttpHandler: mux,
			}
		}
	}

	return proxyServers, nil
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
