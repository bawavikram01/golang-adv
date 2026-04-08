// =============================================================================
// LESSON 2: SYNCHRONOUS COMMUNICATION — REST, gRPC, and Service-to-Service
// =============================================================================
//
// Synchronous = caller WAITS for a response.
// Two main options: REST (HTTP/JSON) and gRPC (HTTP/2 + Protobuf).
//
// THIS FILE BUILDS:
//   - A REST microservice with proper patterns
//   - An HTTP client with timeouts, retries, connection pooling
//   - gRPC concepts and comparison
//
// WHEN TO USE SYNC:
//   ✅ Query/read operations ("get user profile")
//   ✅ Operations where caller needs immediate response
//   ✅ Simple request-response patterns
//
// WHEN TO AVOID:
//   ❌ Long-running operations (use async)
//   ❌ Fan-out to many services (latency adds up)
//   ❌ Fire-and-forget (use events)
// =============================================================================

package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// =============================================================================
// PATTERN 1: REST Service with proper structure
// =============================================================================

// --- Domain ---
type Product struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Price       float64 `json:"price"`
	Description string  `json:"description,omitempty"`
}

// --- Repository (data layer) ---
type ProductRepository struct {
	mu       sync.RWMutex
	products map[int64]*Product
	nextID   int64
}

func NewProductRepository() *ProductRepository {
	return &ProductRepository{
		products: map[int64]*Product{
			1: {ID: 1, Name: "Go Book", Price: 49.99},
			2: {ID: 2, Name: "Keyboard", Price: 149.99},
			3: {ID: 3, Name: "Monitor", Price: 399.99},
		},
		nextID: 4,
	}
}

func (r *ProductRepository) FindByID(id int64) (*Product, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.products[id]
	if !ok {
		return nil, fmt.Errorf("product %d not found", id)
	}
	return p, nil
}

func (r *ProductRepository) FindAll() []*Product {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*Product, 0, len(r.products))
	for _, p := range r.products {
		result = append(result, p)
	}
	return result
}

// --- API Response Types ---
// ALWAYS wrap responses for consistency

type APIResponse struct {
	Data  interface{} `json:"data,omitempty"`
	Error *APIError   `json:"error,omitempty"`
}

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// --- HTTP Handlers ---
type ProductHandler struct {
	repo *ProductRepository
}

func (h *ProductHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "invalid product ID")
		return
	}

	product, err := h.repo.FindByID(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{Data: product})
}

func (h *ProductHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	products := h.repo.FindAll()
	writeJSON(w, http.StatusOK, APIResponse{Data: products})
}

// --- Health Check (CRITICAL for microservices) ---
func healthCheck(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status":  "healthy",
		"service": "product-service",
		"time":    time.Now().UTC().Format(time.RFC3339),
	})
}

// --- JSON Helpers ---
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, APIResponse{Error: &APIError{Code: code, Message: message}})
}

// =============================================================================
// PATTERN 2: HTTP Client for service-to-service calls
// =============================================================================
//
// CRITICAL RULES for service clients:
// 1. ALWAYS set timeouts (connection, request, response)
// 2. Reuse http.Client (connection pooling)
// 3. Close response body
// 4. Use context for per-request timeout
// 5. Handle partial failures gracefully

type ProductClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewProductClient(baseURL string) *ProductClient {
	return &ProductClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second, // overall request timeout
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				MaxConnsPerHost:     100,
				IdleConnTimeout:     90 * time.Second,
			},
		},
	}
}

func (c *ProductClient) GetProduct(ctx context.Context, id int64) (*Product, error) {
	// Per-request timeout (more granular than client timeout)
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	url := fmt.Sprintf("%s/products/%d", c.baseURL, id)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	// Add service identity headers
	req.Header.Set("X-Service-Name", "order-service")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, fmt.Errorf("product service timeout: %w", err)
		}
		return nil, fmt.Errorf("product service call: %w", err)
	}
	defer resp.Body.Close()

	// Limit body read to prevent OOM from malicious responses
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // 1MB max
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("product service returned %d: %s", resp.StatusCode, body)
	}

	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	// Re-decode the Data field into Product
	dataBytes, _ := json.Marshal(apiResp.Data)
	var product Product
	if err := json.Unmarshal(dataBytes, &product); err != nil {
		return nil, fmt.Errorf("decode product: %w", err)
	}

	return &product, nil
}

// =============================================================================
// PATTERN 3: REST vs gRPC Comparison
// =============================================================================
//
// ┌─────────────────┬───────────────────────┬────────────────────────┐
// │                  │ REST (HTTP/JSON)       │ gRPC (HTTP/2+Protobuf) │
// ├─────────────────┼───────────────────────┼────────────────────────┤
// │ Encoding         │ JSON (text, ~10x slower)│ Protobuf (binary, fast)│
// │ Transport        │ HTTP/1.1 or HTTP/2    │ HTTP/2 only            │
// │ Schema           │ OpenAPI/Swagger       │ .proto files (strict)  │
// │ Streaming        │ SSE, WebSocket        │ Native bidirectional   │
// │ Browser support  │ ✅ Native             │ ❌ Needs grpc-web      │
// │ Code gen         │ Optional (openapi-gen) │ Required (protoc)      │
// │ Debugging        │ curl, browser, Postman│ grpcurl, Postman       │
// │ Performance      │ Good                  │ 2-10x faster           │
// │ Learning curve   │ Low                   │ Medium                 │
// │ Best for         │ Public APIs, web      │ Internal service-to-svc│
// └─────────────────┴───────────────────────┴────────────────────────┘
//
// RECOMMENDATION:
//   External API (public, web, mobile) → REST
//   Internal service-to-service        → gRPC (or REST, both work)
//   Streaming (real-time data)          → gRPC or WebSocket
//   Simple internal calls               → REST is fine

// =============================================================================
// PATTERN 4: API Versioning
// =============================================================================
//
// Three approaches:
//
// 1. URL versioning:    /v1/products, /v2/products
//    ✅ Simple, obvious  ❌ Creates multiple route trees
//
// 2. Header versioning: Accept: application/vnd.myapi.v2+json
//    ✅ Clean URLs  ❌ Easy to miss, harder to test
//
// 3. Query param:       /products?version=2
//    ✅ Easy to add  ❌ Messy, not RESTful
//
// RECOMMENDATION: URL versioning for simplicity (/v1/...)

// =============================================================================
// PATTERN 5: Idempotency — Making operations safe to retry
// =============================================================================
//
// In microservices, network can fail AFTER the server processed the request
// but BEFORE the client received the response. The client will retry,
// causing the operation to execute TWICE.
//
// SOLUTION: Idempotency keys
//
//   Client sends: POST /orders  + Header: Idempotency-Key: abc-123-def
//   Server: "Have I seen abc-123-def before?"
//     Yes → return cached response
//     No  → process, store result keyed by abc-123-def, return response

type IdempotencyStore struct {
	mu    sync.Mutex
	store map[string][]byte // key → cached JSON response
}

func NewIdempotencyStore() *IdempotencyStore {
	return &IdempotencyStore{store: make(map[string][]byte)}
}

func (s *IdempotencyStore) Get(key string) ([]byte, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.store[key]
	return v, ok
}

func (s *IdempotencyStore) Set(key string, response []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.store[key] = response
}

func idempotencyMiddleware(store *IdempotencyStore, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only for non-safe methods (POST, PUT, PATCH)
		if r.Method == http.MethodGet {
			next.ServeHTTP(w, r)
			return
		}

		key := r.Header.Get("Idempotency-Key")
		if key == "" {
			next.ServeHTTP(w, r)
			return
		}

		// Check if we've seen this key before
		if cached, ok := store.Get(key); ok {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-Idempotency-Replay", "true")
			w.Write(cached)
			return
		}

		// Process request and cache response
		// (In production, use a response recorder)
		next.ServeHTTP(w, r)
	})
}

func main() {
	repo := NewProductRepository()
	handler := &ProductHandler{repo: repo}
	idempotency := NewIdempotencyStore()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthCheck)
	mux.HandleFunc("GET /products", handler.ListProducts)
	mux.HandleFunc("GET /products/{id}", handler.GetProduct)

	// Wrap with idempotency middleware
	wrappedMux := idempotencyMiddleware(idempotency, mux)

	server := &http.Server{
		Addr:         ":8081",
		Handler:      wrappedMux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in background
	go func() {
		fmt.Println("Product Service starting on :8081")
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Demonstrate client calling the service
	client := NewProductClient("http://localhost:8081")
	ctx := context.Background()

	product, err := client.GetProduct(ctx, 1)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Got product: %+v\n", product)
	}

	// Try non-existent product
	_, err = client.GetProduct(ctx, 999)
	fmt.Printf("Non-existent product error: %v\n", err)

	// Shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.Shutdown(ctx)

	fmt.Println("\n=== KEY PATTERNS ===")
	fmt.Println("1. REST for public APIs, gRPC for internal service-to-service")
	fmt.Println("2. Always set timeouts on HTTP clients AND servers")
	fmt.Println("3. Use idempotency keys for non-GET operations")
	fmt.Println("4. Health check endpoints for every service")
	fmt.Println("5. Structured error responses with error codes")
	fmt.Println("6. API versioning via URL prefix (/v1/...)")
}
