package proxy

import (
	"net/http"
	"strings"
)

type ConfigLoader struct {
	Config *Config
}

type TProxyServer struct {
	Ssl         bool
	HttpHandler http.Handler
}

type THostServer struct {
	path string
	px   *ProxyServer
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
	hostServers := make(map[string][]THostServer)

	for _, server := range cl.Config.Servers {
		if _, ok := hostServers[server.Host]; !ok {
			hostServers[server.Host] = []THostServer{}
		}

		// Create a router to handle different routes
		for _, route := range server.Routes {
			px, err := cl.createProxyServer(route)
			if err != nil {
				return nil, err
			}

			// Append the new THostServer to the list
			hostServers[server.Host] = append(hostServers[server.Host], THostServer{
				path: route.Match.Path,
				px:   px,
			})
		}
	}

	for _, server := range cl.Config.Servers {
		mux := cl.createMuxServer(&hostServers)
		proxyServers[server.Listen] = &TProxyServer{
			Ssl:         server.Ssl,
			HttpHandler: mux,
		}
	}

	return proxyServers, nil
}

func (cl *ConfigLoader) createMuxServer(hostServers *map[string][]THostServer) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		host := r.Host

		if hostServer, ok := (*hostServers)[host]; ok {
			for _, hs := range hostServer {
				if strings.HasPrefix(r.URL.Path, hs.path) {
					r.URL.Path = r.URL.Path[len(hs.path):]
					hs.px.ServeHTTP(w, r)
					return
				}
			}
		}
		http.NotFound(w, r)
	})

	return mux
}

func (cl *ConfigLoader) createProxyServer(route RouteConfig) (*ProxyServer, error) {
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
