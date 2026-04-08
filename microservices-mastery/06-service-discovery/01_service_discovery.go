// =============================================================================
// LESSON 6: SERVICE DISCOVERY & LOAD BALANCING
// =============================================================================
//
// In a monolith: everything is localhost.
// In microservices: services run on dynamic IPs, multiple instances, containers.
//
// HOW DOES SERVICE A FIND SERVICE B?
// Answer: Service Discovery.
//
// TWO APPROACHES:
//   1. CLIENT-SIDE DISCOVERY — client queries a registry, picks an instance
//   2. SERVER-SIDE DISCOVERY — client calls a load balancer, LB picks instance
//
// SERVICE REGISTRY: A database of service instances.
//   Examples: Consul, etcd, ZooKeeper, Kubernetes DNS, Eureka
//
// REGISTRATION:
//   Self-registration: Service registers itself on startup, deregisters on shutdown
//   Third-party:       A sidecar or platform registers the service (Kubernetes does this)
// =============================================================================

package main

import (
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

// =============================================================================
// PART 1: Service Instance & Registry
// =============================================================================

type ServiceInstance struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"` // logical service name
	Host      string            `json:"host"`
	Port      int               `json:"port"`
	Metadata  map[string]string `json:"metadata"` // version, zone, weight
	Healthy   bool              `json:"healthy"`
	LastCheck time.Time         `json:"last_check"`
}

func (si ServiceInstance) Address() string {
	return fmt.Sprintf("%s:%d", si.Host, si.Port)
}

// ServiceRegistry — in-memory simulation of Consul/etcd/Kubernetes service registry
type ServiceRegistry struct {
	mu       sync.RWMutex
	services map[string][]ServiceInstance // service name → instances
}

func NewServiceRegistry() *ServiceRegistry {
	return &ServiceRegistry{
		services: make(map[string][]ServiceInstance),
	}
}

// Register — service instance comes online
func (sr *ServiceRegistry) Register(instance ServiceInstance) {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	instance.Healthy = true
	instance.LastCheck = time.Now()
	sr.services[instance.Name] = append(sr.services[instance.Name], instance)
	fmt.Printf("  [Registry] Registered: %s @ %s\n", instance.ID, instance.Address())
}

// Deregister — service instance going offline (graceful shutdown)
func (sr *ServiceRegistry) Deregister(serviceID string) {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	for name, instances := range sr.services {
		for i, inst := range instances {
			if inst.ID == serviceID {
				sr.services[name] = append(instances[:i], instances[i+1:]...)
				fmt.Printf("  [Registry] Deregistered: %s\n", serviceID)
				return
			}
		}
	}
}

// GetHealthyInstances — returns only healthy instances of a service
func (sr *ServiceRegistry) GetHealthyInstances(serviceName string) []ServiceInstance {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	var healthy []ServiceInstance
	for _, inst := range sr.services[serviceName] {
		if inst.Healthy {
			healthy = append(healthy, inst)
		}
	}
	return healthy
}

// HealthCheck — marks unhealthy instances (runs periodically)
func (sr *ServiceRegistry) HealthCheck(timeout time.Duration) {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	for name, instances := range sr.services {
		for i, inst := range instances {
			if time.Since(inst.LastCheck) > timeout {
				instances[i].Healthy = false
				fmt.Printf("  [Registry] %s marked UNHEALTHY (no heartbeat)\n", inst.ID)
			}
		}
		sr.services[name] = instances
	}
}

// Heartbeat — service sends a heartbeat to stay registered
func (sr *ServiceRegistry) Heartbeat(serviceID string) {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	for name, instances := range sr.services {
		for i, inst := range instances {
			if inst.ID == serviceID {
				instances[i].LastCheck = time.Now()
				instances[i].Healthy = true
			}
		}
		sr.services[name] = instances
	}
}

// =============================================================================
// PART 2: Load Balancing Strategies
// =============================================================================
//
// Once you have multiple instances, HOW do you pick which one to call?
//
// ┌────────────────────────┬──────────────────────────────────────────────┐
// │ Strategy               │ Description                                  │
// ├────────────────────────┼──────────────────────────────────────────────┤
// │ Round Robin            │ Cycle through instances 1→2→3→1→2→3         │
// │ Weighted Round Robin   │ More traffic to beefier instances            │
// │ Random                 │ Pick a random instance                       │
// │ Least Connections      │ Pick the instance with fewest active calls   │
// │ Consistent Hashing     │ Same key always goes to same instance        │
// │ IP Hash                │ Client IP determines instance (sticky)       │
// └────────────────────────┴──────────────────────────────────────────────┘

type LoadBalancer interface {
	Pick(instances []ServiceInstance) *ServiceInstance
}

// --- Round Robin ---
type RoundRobinLB struct {
	counter atomic.Uint64
}

func (lb *RoundRobinLB) Pick(instances []ServiceInstance) *ServiceInstance {
	if len(instances) == 0 {
		return nil
	}
	idx := lb.counter.Add(1) % uint64(len(instances))
	return &instances[idx]
}

// --- Weighted Round Robin ---
type WeightedRoundRobinLB struct {
	counter atomic.Uint64
}

func (lb *WeightedRoundRobinLB) Pick(instances []ServiceInstance) *ServiceInstance {
	if len(instances) == 0 {
		return nil
	}

	// Build weighted list: instance with weight=3 appears 3 times
	var weighted []ServiceInstance
	for _, inst := range instances {
		weight := 1
		if w, ok := inst.Metadata["weight"]; ok {
			fmt.Sscanf(w, "%d", &weight)
		}
		for range weight {
			weighted = append(weighted, inst)
		}
	}

	idx := lb.counter.Add(1) % uint64(len(weighted))
	return &weighted[idx]
}

// --- Random ---
type RandomLB struct{}

func (lb *RandomLB) Pick(instances []ServiceInstance) *ServiceInstance {
	if len(instances) == 0 {
		return nil
	}
	return &instances[rand.Intn(len(instances))]
}

// --- Least Connections ---
type LeastConnectionsLB struct {
	mu          sync.RWMutex
	connections map[string]int // instance ID → active connections
}

func NewLeastConnectionsLB() *LeastConnectionsLB {
	return &LeastConnectionsLB{connections: make(map[string]int)}
}

func (lb *LeastConnectionsLB) Pick(instances []ServiceInstance) *ServiceInstance {
	if len(instances) == 0 {
		return nil
	}

	lb.mu.RLock()
	defer lb.mu.RUnlock()

	minConns := int(^uint(0) >> 1) // max int
	var selected *ServiceInstance

	for i, inst := range instances {
		conns := lb.connections[inst.ID]
		if conns < minConns {
			minConns = conns
			selected = &instances[i]
		}
	}
	return selected
}

func (lb *LeastConnectionsLB) Connect(instanceID string) {
	lb.mu.Lock()
	lb.connections[instanceID]++
	lb.mu.Unlock()
}

func (lb *LeastConnectionsLB) Disconnect(instanceID string) {
	lb.mu.Lock()
	lb.connections[instanceID]--
	lb.mu.Unlock()
}

// --- Consistent Hashing ---
// Used when you want the SAME key to always go to the SAME instance.
// Example: user-123 always goes to instance-2 (good for caching).
// When an instance is added/removed, only 1/N keys are remapped.
type ConsistentHashLB struct{}

func (lb *ConsistentHashLB) PickByKey(instances []ServiceInstance, key string) *ServiceInstance {
	if len(instances) == 0 {
		return nil
	}
	// Simple hash: sum of bytes mod instance count
	// Production: use a hash ring (like hashicorp/memberlist or ketama)
	hash := uint64(0)
	for _, b := range []byte(key) {
		hash = hash*31 + uint64(b)
	}
	idx := hash % uint64(len(instances))
	return &instances[idx]
}

// =============================================================================
// PART 3: Client-Side Discovery (like Netflix Ribbon)
// =============================================================================
//
// The CLIENT queries the registry and picks an instance.
//
// client → registry.GetInstances("payment") → [inst1, inst2, inst3]
// client → loadBalancer.Pick([inst1, inst2, inst3]) → inst2
// client → http.Get("http://inst2:8080/pay")
//
// PROS: No single point of failure (no LB proxy). Fewer hops.
// CONS: Every client needs discovery logic. Language-specific.

type ServiceClient struct {
	registry *ServiceRegistry
	lb       LoadBalancer
}

func NewServiceClient(registry *ServiceRegistry, lb LoadBalancer) *ServiceClient {
	return &ServiceClient{registry: registry, lb: lb}
}

func (sc *ServiceClient) Call(serviceName string) (string, error) {
	instances := sc.registry.GetHealthyInstances(serviceName)
	if len(instances) == 0 {
		return "", fmt.Errorf("no healthy instances of '%s'", serviceName)
	}

	instance := sc.lb.Pick(instances)
	if instance == nil {
		return "", fmt.Errorf("load balancer returned nil")
	}

	// In production: HTTP/gRPC call to instance.Address()
	return fmt.Sprintf("Response from %s @ %s", instance.ID, instance.Address()), nil
}

// =============================================================================
// PART 4: Server-Side Discovery (like AWS ALB, Kubernetes Service)
// =============================================================================
//
// A proxy/load balancer sits between client and services.
//
// client → load_balancer:8080 → picks instance → forwards request
//
// KUBERNETES does this automatically:
//   my-service.namespace.svc.cluster.local → kube-proxy → pod
//
// PROS: Client is dumb (just knows one URL). Language-agnostic.
// CONS: Extra hop. Load balancer is a single point of failure (need HA).

type ReverseProxy struct {
	registry *ServiceRegistry
	lb       LoadBalancer
}

func NewReverseProxy(registry *ServiceRegistry, lb LoadBalancer) *ReverseProxy {
	return &ReverseProxy{registry: registry, lb: lb}
}

func (rp *ReverseProxy) HandleRequest(serviceName, path string) (string, error) {
	instances := rp.registry.GetHealthyInstances(serviceName)
	if len(instances) == 0 {
		return "", fmt.Errorf("502 Bad Gateway: no instances of '%s'", serviceName)
	}

	instance := rp.lb.Pick(instances)
	return fmt.Sprintf("Proxied to %s @ %s%s", instance.ID, instance.Address(), path), nil
}

// =============================================================================
// CONCEPTS: DNS-Based Discovery
// =============================================================================
//
// Simplest approach: use DNS to resolve service names.
//
// Kubernetes: my-service.default.svc.cluster.local → cluster IP
//   - A records for ClusterIP services
//   - SRV records for headless services (returns all pod IPs)
//
// Consul DNS: payment.service.consul → returns healthy instance IPs
//
// LIMITATIONS:
//   - DNS TTL caching can return stale IPs
//   - No health checking (depends on the DNS provider)
//   - No load balancing metadata (weights, zones)
//   - DNS round-robin is not true load balancing

func main() {
	// =========================================================================
	// DEMO 1: Service Registry
	// =========================================================================
	fmt.Println("=== SERVICE REGISTRY ===")

	registry := NewServiceRegistry()

	// Register multiple instances of "payment-service"
	registry.Register(ServiceInstance{
		ID: "payment-1", Name: "payment", Host: "10.0.1.1", Port: 8080,
		Metadata: map[string]string{"version": "v2", "zone": "us-east-1a", "weight": "3"},
	})
	registry.Register(ServiceInstance{
		ID: "payment-2", Name: "payment", Host: "10.0.1.2", Port: 8080,
		Metadata: map[string]string{"version": "v2", "zone": "us-east-1b", "weight": "1"},
	})
	registry.Register(ServiceInstance{
		ID: "payment-3", Name: "payment", Host: "10.0.1.3", Port: 8080,
		Metadata: map[string]string{"version": "v1", "zone": "us-west-2a", "weight": "2"},
	})

	instances := registry.GetHealthyInstances("payment")
	fmt.Printf("  Healthy 'payment' instances: %d\n", len(instances))

	// =========================================================================
	// DEMO 2: Load Balancing Strategies
	// =========================================================================
	fmt.Println("\n=== ROUND ROBIN ===")
	rrLB := &RoundRobinLB{}
	for i := 0; i < 6; i++ {
		inst := rrLB.Pick(instances)
		fmt.Printf("  Request %d → %s (%s)\n", i+1, inst.ID, inst.Metadata["zone"])
	}

	fmt.Println("\n=== WEIGHTED ROUND ROBIN ===")
	wrrLB := &WeightedRoundRobinLB{}
	counts := make(map[string]int)
	for i := 0; i < 12; i++ {
		inst := wrrLB.Pick(instances)
		counts[inst.ID]++
	}
	for id, count := range counts {
		fmt.Printf("  %s: %d requests\n", id, count)
	}

	fmt.Println("\n=== CONSISTENT HASHING ===")
	chLB := &ConsistentHashLB{}
	// Same key always goes to same instance
	for i := 0; i < 3; i++ {
		inst := chLB.PickByKey(instances, "user-123")
		fmt.Printf("  user-123 → %s (attempt %d)\n", inst.ID, i+1)
	}
	inst := chLB.PickByKey(instances, "user-456")
	fmt.Printf("  user-456 → %s\n", inst.ID)

	// =========================================================================
	// DEMO 3: Client-Side Discovery
	// =========================================================================
	fmt.Println("\n=== CLIENT-SIDE DISCOVERY ===")
	client := NewServiceClient(registry, &RoundRobinLB{})

	for i := 0; i < 3; i++ {
		response, err := client.Call("payment")
		if err != nil {
			fmt.Printf("  Error: %v\n", err)
		} else {
			fmt.Printf("  %s\n", response)
		}
	}

	// =========================================================================
	// DEMO 4: Server-Side Discovery (Reverse Proxy)
	// =========================================================================
	fmt.Println("\n=== SERVER-SIDE DISCOVERY (Reverse Proxy) ===")
	proxy := NewReverseProxy(registry, &RoundRobinLB{})

	for i := 0; i < 3; i++ {
		response, err := proxy.HandleRequest("payment", "/api/v1/charge")
		if err != nil {
			fmt.Printf("  Error: %v\n", err)
		} else {
			fmt.Printf("  %s\n", response)
		}
	}

	// =========================================================================
	// DEMO 5: Health Check & Deregistration
	// =========================================================================
	fmt.Println("\n=== HEALTH CHECK & FAILOVER ===")
	// Simulate payment-2 going down (no heartbeat)
	registry.mu.Lock()
	for i, inst := range registry.services["payment"] {
		if inst.ID == "payment-2" {
			registry.services["payment"][i].LastCheck = time.Now().Add(-10 * time.Minute)
		}
	}
	registry.mu.Unlock()

	registry.HealthCheck(30 * time.Second) // timeout: 30s without heartbeat = unhealthy
	healthy := registry.GetHealthyInstances("payment")
	fmt.Printf("  Healthy instances after check: %d\n", len(healthy))

	registry.Deregister("payment-3")
	healthy = registry.GetHealthyInstances("payment")
	fmt.Printf("  After deregistering payment-3: %d healthy\n", len(healthy))

	// =========================================================================
	// Summary
	// =========================================================================
	fmt.Println("\n=== SERVICE DISCOVERY DECISION GUIDE ===")
	fmt.Println("┌──────────────────────────────┬────────────────────────────────────┐")
	fmt.Println("│ Approach                     │ Best For                           │")
	fmt.Println("├──────────────────────────────┼────────────────────────────────────┤")
	fmt.Println("│ Client-side (Ribbon/Consul)  │ High performance, Go/Java services │")
	fmt.Println("│ Server-side (ALB/Nginx)      │ Polyglot, simple clients           │")
	fmt.Println("│ DNS-based (K8s, Consul DNS)  │ Simple setups, no code changes     │")
	fmt.Println("│ Service Mesh (Istio/Linkerd) │ Advanced: mTLS, canary, observ.    │")
	fmt.Println("└──────────────────────────────┴────────────────────────────────────┘")
	fmt.Println()
	fmt.Println("KUBERNETES CHEAT SHEET:")
	fmt.Println("  ClusterIP Service → stable internal DNS + load balancing")
	fmt.Println("  Headless Service  → DNS returns all pod IPs (client-side LB)")
	fmt.Println("  NodePort/LB       → external access")
	fmt.Println("  Ingress           → HTTP routing (path/host-based)")
}
