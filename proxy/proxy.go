package proxy

import (
	"fmt"
	"log"
	"math/rand"
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
}

// LoadBalancer 負載均衡器
type LoadBalancer struct {
	servers []*UpstreamServer
	mu      sync.RWMutex
	// 負載均衡策略函數
	strategy func([]*UpstreamServer) *UpstreamServer
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

	lb := &LoadBalancer{
		servers:  servers,
		strategy: roundRobinStrategy, // 默認使用輪詢策略
	}

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

// 輪詢策略
func roundRobinStrategy(servers []*UpstreamServer) *UpstreamServer {
	var alive []*UpstreamServer
	for _, server := range servers {
		if server.Alive {
			alive = append(alive, server)
		}
	}
	if len(alive) == 0 {
		return nil
	}
	// 使用隨機選擇來實現簡單的輪詢
	return alive[rand.Intn(len(alive))]
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

// ServeHTTP 實現了 http.Handler 接口
func (p *ProxyServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	server := p.LoadBalancer.strategy(p.LoadBalancer.servers)
	if server == nil {
		http.Error(w, "沒有可用的上游伺服器", http.StatusServiceUnavailable)
		return
	}

	// 添加代理相關的 header
	r.Header.Add("X-Forwarded-For", r.RemoteAddr)
	r.Header.Add("X-Real-IP", r.RemoteAddr)
	r.Header.Add("X-Proxy-Id", "go-reverse-engine")

	server.ReverseProxy.ServeHTTP(w, r)
}
