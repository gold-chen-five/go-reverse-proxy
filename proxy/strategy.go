package proxy

import (
	"hash/fnv"
	"sync"
	"sync/atomic"
)

type Strategy string

const (
	RoundRobin       Strategy = "round-robin"
	LeastConnections Strategy = "least-connections"
	IPHash           Strategy = "ip-hash"
	WeightedRR       Strategy = "weighted-round-robin"
)

type StrategyHandler interface {
	NextServer(servers []*UpstreamServer, remoteAddr string) *UpstreamServer
	IncrementConnections(server *UpstreamServer)
	DecrementConnections(server *UpstreamServer)
}

type BaseStrategy struct {
	mu      sync.RWMutex
	counter uint32
}

type RoundRobinStrategy struct {
	BaseStrategy
}

type LeastConnectionsStrategy struct {
	BaseStrategy
}

type IPHashStrategy struct {
	BaseStrategy
}

type WeightedRoundRobinStrategy struct {
	BaseStrategy
}

func NewStrategy(strategyType Strategy) StrategyHandler {
	switch strategyType {
	case RoundRobin:
		return &RoundRobinStrategy{}
	case LeastConnections:
		return &LeastConnectionsStrategy{}
	case IPHash:
		return &IPHashStrategy{}
	case WeightedRR:
		return &WeightedRoundRobinStrategy{}
	default:
		return &RoundRobinStrategy{}
	}
}

// get the alive server
func getAliveServers(servers []*UpstreamServer) []*UpstreamServer {
	var alive []*UpstreamServer
	for _, server := range servers {
		if server.Alive {
			alive = append(alive, server)
		}
	}
	return alive
}

// 輪詢策略
func (s *RoundRobinStrategy) NextServer(servers []*UpstreamServer, _ string) *UpstreamServer {
	aliveServer := getAliveServers(servers)
	if len(aliveServer) == 0 {
		return nil
	}
	next := atomic.AddUint32(&s.counter, 1)
	return aliveServer[next%uint32(len(aliveServer))]
}

// 最少連接策略
func (s *LeastConnectionsStrategy) NextServer(servers []*UpstreamServer, _ string) *UpstreamServer {
	var minServer *UpstreamServer
	minConnections := int32(1<<31 - 1) // Max int32

	alive := getAliveServers(servers)
	if len(alive) == 0 {
		return nil
	}

	for _, server := range alive {
		connections := atomic.LoadInt32(&server.ActiveConns)
		if connections < minConnections {
			minConnections = connections
			minServer = server
		}
	}

	return minServer
}

// IP Hash
func (s *IPHashStrategy) NextServer(servers []*UpstreamServer, remoteAddr string) *UpstreamServer {
	alive := getAliveServers(servers)
	if len(alive) == 0 {
		return nil
	}

	h := fnv.New32a()
	h.Write([]byte(remoteAddr))
	hash := h.Sum32()

	return alive[hash%uint32(len(alive))]
}

// 加權輪詢策略
func (s *WeightedRoundRobinStrategy) NextServer(servers []*UpstreamServer, _ string) *UpstreamServer {
	s.mu.Lock()
	defer s.mu.Unlock()

	alive := getAliveServers(servers)
	if len(alive) == 0 {
		return nil
	}

	var totalWeight int32
	var maxServer *UpstreamServer
	var maxWeight int32 = -1

	// First pass - calculate total weight and find highest weight server
	for _, server := range alive {
		totalWeight += server.Weight
		server.CurrentWeight += server.Weight

		if server.CurrentWeight > maxWeight {
			maxWeight = server.CurrentWeight
			maxServer = server
		}
	}

	if maxServer == nil {
		return nil
	}

	// Decrease current weight
	maxServer.CurrentWeight -= totalWeight

	return maxServer
}

// Helper functions for server management
func (s *BaseStrategy) UpdateWeight(server *UpstreamServer, weight int32) {
	atomic.StoreInt32(&server.Weight, weight)
}

func (s *BaseStrategy) IncrementConnections(server *UpstreamServer) {
	atomic.AddInt32(&server.ActiveConns, 1)
}

func (s *BaseStrategy) DecrementConnections(server *UpstreamServer) {
	atomic.AddInt32(&server.ActiveConns, -1)
}

func (s *BaseStrategy) GetConnections(server *UpstreamServer) int32 {
	return atomic.LoadInt32(&server.ActiveConns)
}
