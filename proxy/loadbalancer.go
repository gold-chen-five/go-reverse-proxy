package proxy

import (
	"fmt"
	"sync"
)

// LoadBalancer 負載均衡器
type LoadBalancer struct {
	servers []*UpstreamServer
	mu      sync.RWMutex
	// 負載均衡策略函數
	strategy        Strategy
	strategyHandler StrategyHandler
}

func NewLoadBalancer(servers []*UpstreamServer, strategy Strategy) *LoadBalancer {
	lb := &LoadBalancer{
		servers:         servers,
		strategy:        strategy,
		strategyHandler: NewStrategy(strategy),
	}
	return lb
}

// UpdateStrategy updates the load balancing strategy
func (lb *LoadBalancer) UpdateStrategy(strategy Strategy) {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	lb.strategy = strategy
	lb.strategyHandler = NewStrategy(strategy)
}

// GetNextServer returns the next server based on the current strategy
func (lb *LoadBalancer) GetNextServer(remoteAddr string) *UpstreamServer {
	lb.mu.RLock()
	defer lb.mu.RUnlock()
	return lb.strategyHandler.NextServer(lb.servers, remoteAddr)
}

// SetServerWeight sets the weight for weighted round-robin
func (lb *LoadBalancer) SetServerWeight(serverURL string, weight int32) error {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	for _, server := range lb.servers {
		if server.URL.String() == serverURL {
			server.Weight = weight
			return nil
		}
	}

	return fmt.Errorf("server not found: %s", serverURL)
}
