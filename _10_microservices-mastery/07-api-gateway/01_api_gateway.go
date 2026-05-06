// =============================================================================
// LESSON 7: API GATEWAY PATTERN
// =============================================================================
//
// The API Gateway is the SINGLE ENTRY POINT for all client requests.
// Instead of clients calling 10 microservices directly, they call ONE gateway.
//
// CLIENT → API Gateway → routes to correct microservice(s)
//
// WHAT THE GATEWAY DOES:
//   ✅ Request routing       (path-based, host-based)
//   ✅ Authentication/AuthZ  (validate JWT, API keys)
//   ✅ Rate limiting         (per-client, per-endpoint)
//   ✅ Request aggregation   (call multiple services, merge responses)
//   ✅ Protocol translation  (REST → gRPC, WebSocket → HTTP)
//   ✅ Caching               (cache common responses)
//   ✅ Load balancing        (distribute across instances)
//   ✅ Circuit breaking      (protect downstream services)
//   ✅ Logging & metrics     (centralized observability)
//   ✅ Request/response transformation (add headers, modify bodies)
//
// REAL GATEWAYS: Kong, Envoy, NGINX, AWS API Gateway, Traefik, Tyk
// =============================================================================

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// =============================================================================
// PART 1: Route Configuration
// =============================================================================

type Route struct {
	Path        string            `json:"path"`         // e.g., "/api/v1/users"
	Method      string            `json:"method"`       // GET, POST, etc. ("*" for all)
	ServiceName string            `json:"service_name"` // target microservice
	StripPrefix string            `json:"strip_prefix"` // remove from path before forwarding
	Middleware  []string          `json:"middleware"`   // auth, ratelimit, cache
	Timeout     time.Duration     `json:"timeout"`
	Headers     map[string]string `json:"headers"` // headers to add to request
}

type RouteConfig struct {
	routes []Route
}

func NewRouteConfig() *RouteConfig {
	return &RouteConfig{}
}

func (rc *RouteConfig) AddRoute(route Route) {
	rc.routes = append(rc.routes, route)
}

func (rc *RouteConfig) Match(method, path string) *Route {
	for _, r := range rc.routes {
		if (r.Method == "*" || r.Method == method) && strings.HasPrefix(path, r.Path) {
			return &r
		}
	}
	return nil
}

// =============================================================================
// PART 2: Middleware Chain
// =============================================================================
//
// Middleware = functions that run BEFORE and AFTER the actual request.
// They form a chain: Auth → RateLimit → Cache → Logging → Forward
//
// Each middleware can:
//   - Modify the request (add headers, validate tokens)
//   - Short-circuit (return 401, 429 without forwarding)
//   - Modify the response (add CORS headers, transform body)
//   - Log/measure (timing, error tracking)

type GatewayRequest struct {
	Method   string
	Path     string
	Headers  map[string]string
	Body     string
	ClientIP string
}

type GatewayResponse struct {
	StatusCode int
	Headers    map[string]string
	Body       string
}

type Middleware func(req *GatewayRequest, next func(*GatewayRequest) *GatewayResponse) *GatewayResponse

// --- Authentication Middleware ---
func AuthMiddleware(validTokens map[string]string) Middleware {
	return func(req *GatewayRequest, next func(*GatewayRequest) *GatewayResponse) *GatewayResponse {
		token := req.Headers["Authorization"]
		if token == "" {
			return &GatewayResponse{StatusCode: 401, Body: `{"error":"missing auth token"}`}
		}

		// Strip "Bearer " prefix
		token = strings.TrimPrefix(token, "Bearer ")

		userID, valid := validTokens[token]
		if !valid {
			fmt.Printf("  [Auth] ✗ Invalid token: %s\n", token[:min(8, len(token))]+"...")
			return &GatewayResponse{StatusCode: 401, Body: `{"error":"invalid token"}`}
		}

		// Inject user info into request headers (downstream services trust the gateway)
		req.Headers["X-User-ID"] = userID
		fmt.Printf("  [Auth] ✓ Authenticated user: %s\n", userID)

		return next(req)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// --- Rate Limiting Middleware ---
type rateLimitState struct {
	mu       sync.Mutex
	requests map[string][]time.Time // client IP → request timestamps
}

func RateLimitMiddleware(maxRequests int, window time.Duration) Middleware {
	state := &rateLimitState{requests: make(map[string][]time.Time)}

	return func(req *GatewayRequest, next func(*GatewayRequest) *GatewayResponse) *GatewayResponse {
		state.mu.Lock()

		now := time.Now()
		clientIP := req.ClientIP

		// Clean old timestamps
		var recent []time.Time
		for _, t := range state.requests[clientIP] {
			if now.Sub(t) < window {
				recent = append(recent, t)
			}
		}
		state.requests[clientIP] = recent

		if len(recent) >= maxRequests {
			state.mu.Unlock()
			fmt.Printf("  [RateLimit] ✗ Client %s exceeded %d req/%v\n", clientIP, maxRequests, window)
			return &GatewayResponse{
				StatusCode: 429,
				Body:       `{"error":"rate limit exceeded"}`,
				Headers:    map[string]string{"Retry-After": "60"},
			}
		}

		state.requests[clientIP] = append(state.requests[clientIP], now)
		state.mu.Unlock()

		fmt.Printf("  [RateLimit] ✓ Client %s: %d/%d requests\n", clientIP, len(recent)+1, maxRequests)
		return next(req)
	}
}

// --- Logging Middleware ---
func LoggingMiddleware() Middleware {
	return func(req *GatewayRequest, next func(*GatewayRequest) *GatewayResponse) *GatewayResponse {
		start := time.Now()

		resp := next(req)

		duration := time.Since(start)
		fmt.Printf("  [Log] %s %s → %d (%v)\n", req.Method, req.Path, resp.StatusCode, duration)

		return resp
	}
}

// --- CORS Middleware ---
func CORSMiddleware(allowedOrigins []string) Middleware {
	return func(req *GatewayRequest, next func(*GatewayRequest) *GatewayResponse) *GatewayResponse {
		resp := next(req)
		if resp.Headers == nil {
			resp.Headers = make(map[string]string)
		}
		resp.Headers["Access-Control-Allow-Origin"] = strings.Join(allowedOrigins, ",")
		resp.Headers["Access-Control-Allow-Methods"] = "GET,POST,PUT,DELETE,OPTIONS"
		resp.Headers["Access-Control-Allow-Headers"] = "Authorization,Content-Type"
		return resp
	}
}

// --- Request ID Middleware ---
func RequestIDMiddleware() Middleware {
	var counter uint64
	var mu sync.Mutex

	return func(req *GatewayRequest, next func(*GatewayRequest) *GatewayResponse) *GatewayResponse {
		mu.Lock()
		counter++
		requestID := fmt.Sprintf("req-%d", counter)
		mu.Unlock()

		req.Headers["X-Request-ID"] = requestID
		fmt.Printf("  [RequestID] Assigned: %s\n", requestID)

		resp := next(req)
		if resp.Headers == nil {
			resp.Headers = make(map[string]string)
		}
		resp.Headers["X-Request-ID"] = requestID
		return resp
	}
}

// =============================================================================
// PART 3: API Gateway
// =============================================================================

type APIGateway struct {
	routes     *RouteConfig
	middleware []Middleware
	services   map[string]func(path string, req *GatewayRequest) *GatewayResponse
}

func NewAPIGateway(routes *RouteConfig) *APIGateway {
	return &APIGateway{
		routes:   routes,
		services: make(map[string]func(string, *GatewayRequest) *GatewayResponse),
	}
}

func (gw *APIGateway) Use(mw Middleware) {
	gw.middleware = append(gw.middleware, mw)
}

func (gw *APIGateway) RegisterService(name string, handler func(string, *GatewayRequest) *GatewayResponse) {
	gw.services[name] = handler
}

func (gw *APIGateway) HandleRequest(req *GatewayRequest) *GatewayResponse {
	// Match route
	route := gw.routes.Match(req.Method, req.Path)
	if route == nil {
		return &GatewayResponse{StatusCode: 404, Body: `{"error":"no route matched"}`}
	}

	// Build middleware chain
	handler := func(r *GatewayRequest) *GatewayResponse {
		svc, ok := gw.services[route.ServiceName]
		if !ok {
			return &GatewayResponse{StatusCode: 502, Body: `{"error":"service unavailable"}`}
		}

		// Strip prefix if configured
		path := r.Path
		if route.StripPrefix != "" {
			path = strings.TrimPrefix(path, route.StripPrefix)
		}

		return svc(path, r)
	}

	// Wrap handler with middleware (reverse order so first middleware runs first)
	for i := len(gw.middleware) - 1; i >= 0; i-- {
		mw := gw.middleware[i]
		next := handler
		handler = func(r *GatewayRequest) *GatewayResponse {
			return mw(r, next)
		}
	}

	return handler(req)
}

// =============================================================================
// PART 4: Backend for Frontend (BFF) Pattern
// =============================================================================
//
// Problem: Mobile, Web, and IoT clients need DIFFERENT data from the same APIs.
//   - Mobile: compact, minimal fields, optimized images
//   - Web: full data, rich UI-specific fields
//   - IoT: tiny payloads, specific protocols
//
// Solution: One gateway per client type.
//   Mobile App  → Mobile BFF  → microservices
//   Web App     → Web BFF     → microservices
//   IoT Device  → IoT BFF     → microservices
//
// Each BFF is tailored to its client's needs:
//   - Aggregates data differently
//   - Returns different field sets
//   - Handles client-specific auth
//   - Optimizes payload size

type BFFResponse struct {
	Platform string      `json:"platform"`
	Data     interface{} `json:"data"`
}

type BFFGateway struct {
	platform  string
	transform func(fullData map[string]interface{}) interface{}
}

func NewBFFGateway(platform string, transform func(map[string]interface{}) interface{}) *BFFGateway {
	return &BFFGateway{platform: platform, transform: transform}
}

func (bff *BFFGateway) GetUserProfile(fullData map[string]interface{}) BFFResponse {
	return BFFResponse{
		Platform: bff.platform,
		Data:     bff.transform(fullData),
	}
}

// =============================================================================
// PART 5: Request Aggregation (Gateway Composition)
// =============================================================================
//
// Client needs data from 3 services. Instead of 3 calls from client:
//   Client → Gateway → calls User + Order + Payment in parallel → merge → return
//
// This reduces:
//   - Client complexity (one call instead of three)
//   - Latency (parallel calls on server side)
//   - Mobile bandwidth (one response instead of three)

type AggregatedResponse struct {
	User    interface{} `json:"user"`
	Orders  interface{} `json:"orders"`
	Payment interface{} `json:"payment"`
}

type Aggregator struct{}

func (a *Aggregator) GetDashboard(ctx context.Context) AggregatedResponse {
	var (
		wg          sync.WaitGroup
		userResp    interface{}
		ordersResp  interface{}
		paymentResp interface{}
	)

	// Parallel calls to multiple services
	wg.Add(3)

	go func() {
		defer wg.Done()
		// In production: HTTP call to user-service
		userResp = map[string]string{"id": "1", "name": "Vikram", "email": "vikram@example.com"}
	}()

	go func() {
		defer wg.Done()
		ordersResp = []map[string]interface{}{
			{"id": 1, "total": 99.99, "status": "delivered"},
			{"id": 2, "total": 49.99, "status": "shipping"},
		}
	}()

	go func() {
		defer wg.Done()
		paymentResp = map[string]string{"balance": "$500.00", "tier": "gold"}
	}()

	wg.Wait()

	return AggregatedResponse{
		User:    userResp,
		Orders:  ordersResp,
		Payment: paymentResp,
	}
}

func main() {
	// =========================================================================
	// DEMO 1: API Gateway with Middleware Chain
	// =========================================================================
	fmt.Println("=== API GATEWAY ===")

	// Configure routes
	routes := NewRouteConfig()
	routes.AddRoute(Route{
		Path: "/api/v1/users", Method: "*", ServiceName: "user-service",
		StripPrefix: "/api/v1", Timeout: 5 * time.Second,
	})
	routes.AddRoute(Route{
		Path: "/api/v1/orders", Method: "*", ServiceName: "order-service",
		StripPrefix: "/api/v1", Timeout: 5 * time.Second,
	})
	routes.AddRoute(Route{
		Path: "/api/v1/payments", Method: "POST", ServiceName: "payment-service",
		StripPrefix: "/api/v1", Timeout: 10 * time.Second,
	})

	// Create gateway
	gw := NewAPIGateway(routes)

	// Add middleware chain (order matters!)
	gw.Use(RequestIDMiddleware())
	gw.Use(LoggingMiddleware())
	gw.Use(AuthMiddleware(map[string]string{
		"token-abc-123": "user-1",
		"token-xyz-789": "user-2",
	}))
	gw.Use(RateLimitMiddleware(5, time.Minute))

	// Register mock backend services
	gw.RegisterService("user-service", func(path string, req *GatewayRequest) *GatewayResponse {
		return &GatewayResponse{
			StatusCode: 200,
			Body:       fmt.Sprintf(`{"user":"%s","path":"%s"}`, req.Headers["X-User-ID"], path),
		}
	})
	gw.RegisterService("order-service", func(path string, req *GatewayRequest) *GatewayResponse {
		return &GatewayResponse{
			StatusCode: 200,
			Body:       `{"orders":[{"id":1,"total":99.99}]}`,
		}
	})

	// Test 1: Authenticated request
	fmt.Println("\n--- Request 1: Valid auth ---")
	resp := gw.HandleRequest(&GatewayRequest{
		Method:   "GET",
		Path:     "/api/v1/users/profile",
		Headers:  map[string]string{"Authorization": "Bearer token-abc-123"},
		ClientIP: "192.168.1.1",
	})
	fmt.Printf("  Response: %d — %s\n", resp.StatusCode, resp.Body)

	// Test 2: Missing auth
	fmt.Println("\n--- Request 2: No auth ---")
	resp = gw.HandleRequest(&GatewayRequest{
		Method:   "GET",
		Path:     "/api/v1/orders",
		Headers:  map[string]string{},
		ClientIP: "192.168.1.2",
	})
	fmt.Printf("  Response: %d — %s\n", resp.StatusCode, resp.Body)

	// Test 3: No matching route
	fmt.Println("\n--- Request 3: Unknown route ---")
	resp = gw.HandleRequest(&GatewayRequest{
		Method:   "GET",
		Path:     "/api/v2/unknown",
		Headers:  map[string]string{"Authorization": "Bearer token-abc-123"},
		ClientIP: "192.168.1.1",
	})
	fmt.Printf("  Response: %d — %s\n", resp.StatusCode, resp.Body)

	// =========================================================================
	// DEMO 2: Backend for Frontend (BFF)
	// =========================================================================
	fmt.Println("\n=== BACKEND FOR FRONTEND (BFF) ===")

	fullUserData := map[string]interface{}{
		"id": 1, "name": "Vikram", "email": "vikram@dev.com",
		"avatar_url":  "https://cdn.example.com/avatars/vikram.jpg",
		"bio":         "Full-stack developer with 10 years of experience in Go and distributed systems.",
		"preferences": map[string]interface{}{"theme": "dark", "language": "en"},
		"address":     map[string]string{"city": "Bangalore", "country": "India"},
		"created_at":  "2020-01-15T10:30:00Z",
	}

	// Mobile BFF — compact response
	mobileBFF := NewBFFGateway("mobile", func(data map[string]interface{}) interface{} {
		return map[string]interface{}{
			"name":   data["name"],
			"avatar": data["avatar_url"],
		}
	})

	// Web BFF — full response
	webBFF := NewBFFGateway("web", func(data map[string]interface{}) interface{} {
		return data // return everything
	})

	mobileResp := mobileBFF.GetUserProfile(fullUserData)
	webResp := webBFF.GetUserProfile(fullUserData)

	mobileJSON, _ := json.MarshalIndent(mobileResp, "  ", "  ")
	webJSON, _ := json.MarshalIndent(webResp, "  ", "  ")

	fmt.Printf("  Mobile BFF response:\n  %s\n", mobileJSON)
	fmt.Printf("\n  Web BFF response:\n  %s\n", webJSON)

	// =========================================================================
	// DEMO 3: Request Aggregation
	// =========================================================================
	fmt.Println("\n=== REQUEST AGGREGATION ===")

	aggregator := &Aggregator{}
	dashboard := aggregator.GetDashboard(context.Background())

	dashJSON, _ := json.MarshalIndent(dashboard, "  ", "  ")
	fmt.Printf("  Dashboard (3 services merged):\n  %s\n", dashJSON)

	// =========================================================================
	// Summary
	// =========================================================================
	fmt.Println("\n=== API GATEWAY PATTERNS ===")
	fmt.Println("┌──────────────────────────────┬──────────────────────────────────┐")
	fmt.Println("│ Pattern                      │ Purpose                          │")
	fmt.Println("├──────────────────────────────┼──────────────────────────────────┤")
	fmt.Println("│ API Gateway                  │ Single entry, cross-cutting      │")
	fmt.Println("│ BFF (Backend for Frontend)   │ Client-specific gateways         │")
	fmt.Println("│ Request Aggregation          │ Merge multiple service calls     │")
	fmt.Println("│ Protocol Translation         │ REST→gRPC, WS→HTTP              │")
	fmt.Println("│ Edge Authentication          │ Validate auth once at gateway    │")
	fmt.Println("│ Response Caching             │ Cache at edge to reduce load     │")
	fmt.Println("└──────────────────────────────┴──────────────────────────────────┘")
	_ = http.StatusOK // imported for completeness
}
