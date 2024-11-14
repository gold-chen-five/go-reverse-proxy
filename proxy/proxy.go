package proxy

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"
)

// UpstreamServer 代表一個上游伺服器
type UpstreamServer struct {
	URL          *url.URL
	Alive        bool
	LastChecked  time.Time
	FailCount    int
	ReverseProxy *httputil.ReverseProxy

	// New fields for enhanced strategies
	Weight          int32        // for weighted round-robin
	CurrentWeight   int32        // for weighted round-robin
	ActiveConns     int32        // for least connections
	connectionsLock sync.RWMutex // protect connections counter
}

// ProxyServer 反向代理伺服器
type ProxyServer struct {
	LoadBalancer *LoadBalancer
	Config       struct {
		HealthCheckInterval time.Duration
		MaxFailCount        int
		Timeout             time.Duration
	}
}

// 創建新的反向代理伺服器
func NewProxyServer(upstreamURLs []string) (*ProxyServer, error) {
	servers := make([]*UpstreamServer, 0, len(upstreamURLs))

	for _, rawURL := range upstreamURLs {
		upstreamURL, err := url.Parse(rawURL)
		if err != nil {
			return nil, fmt.Errorf("invalid upstream URL %s: %v", rawURL, err)
		}

		proxy := httputil.NewSingleHostReverseProxy(upstreamURL)
		// 自定義錯誤處理
		proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			log.Printf("代理錯誤: %v", err)
			http.Error(w, "服務暫時不可用", http.StatusServiceUnavailable)
		}

		servers = append(servers, &UpstreamServer{
			URL:          upstreamURL,
			Alive:        true,
			ReverseProxy: proxy,
		})
	}

	lb := NewLoadBalancer(servers, RoundRobin)

	proxy := &ProxyServer{
		LoadBalancer: lb,
	}

	// 設置默認配置
	proxy.Config.HealthCheckInterval = 10 * time.Second
	proxy.Config.MaxFailCount = 3
	proxy.Config.Timeout = 5 * time.Second

	// 啟動健康檢查
	go proxy.healthCheck()

	return proxy, nil
}

// 健康檢查
func (p *ProxyServer) healthCheck() {
	ticker := time.NewTicker(p.Config.HealthCheckInterval)
	client := &http.Client{
		Timeout: p.Config.Timeout,
	}

	for range ticker.C {
		p.LoadBalancer.mu.Lock()
		for _, server := range p.LoadBalancer.servers {
			go func(server *UpstreamServer) {
				resp, err := client.Get(server.URL.String())
				if err != nil {
					server.FailCount++
					if server.FailCount >= p.Config.MaxFailCount {
						server.Alive = false
					}
					log.Printf("健康檢查失敗 %s: %v", server.URL, err)
					return
				}
				defer resp.Body.Close()

				if resp.StatusCode >= 200 && resp.StatusCode < 300 {
					server.Alive = true
					server.FailCount = 0
				} else {
					server.FailCount++
					if server.FailCount >= p.Config.MaxFailCount {
						server.Alive = false
					}
				}
				server.LastChecked = time.Now()
			}(server)
		}
		p.LoadBalancer.mu.Unlock()
	}
}

func (p *ProxyServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	server := p.LoadBalancer.GetNextServer(r.RemoteAddr)
	if server == nil {
		http.Error(w, "No available upstream servers", http.StatusServiceUnavailable)
		return
	}

	// Track active connections
	p.LoadBalancer.strategyHandler.IncrementConnections(server)
	defer p.LoadBalancer.strategyHandler.DecrementConnections(server)

	// Add proxy headers
	r.Header.Add("X-Forwarded-For", r.RemoteAddr)
	r.Header.Add("X-Real-IP", r.RemoteAddr)
	r.Header.Add("X-Proxy-Id", "go-reverse-engine")

	server.ReverseProxy.ServeHTTP(w, r)
}
