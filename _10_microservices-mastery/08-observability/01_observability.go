// =============================================================================
// LESSON 8: OBSERVABILITY — Tracing, Logging, Metrics
// =============================================================================
//
// "You can't fix what you can't see."
//
// In a monolith: one log file, one profiler, one debugger.
// In microservices: a request touches 5-15 services. Where did it fail?
//                   Where is it slow? Which service is the bottleneck?
//
// THREE PILLARS OF OBSERVABILITY:
//   1. LOGS     — What happened? (structured events)
//   2. METRICS  — How much? How fast? How often? (numbers over time)
//   3. TRACES   — Where did the request go? (distributed path)
//
// TOOLS:
//   Logs:    Elasticsearch/Kibana (ELK), Loki/Grafana, CloudWatch
//   Metrics: Prometheus + Grafana, Datadog, CloudWatch
//   Traces:  Jaeger, Zipkin, Tempo, AWS X-Ray, OpenTelemetry
// =============================================================================

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// =============================================================================
// PILLAR 1: Structured Logging
// =============================================================================
//
// RULE: NEVER use fmt.Println for logs in production.
// ALWAYS use structured logging: key-value pairs, JSON format.
//
// WHY:
//   ✅ Machine-parseable (grep, Elasticsearch, Loki can index fields)
//   ✅ Consistent format across all services
//   ✅ Can filter by service, request_id, user_id, level
//   ✅ Correlation IDs connect logs across services
//
// BAD:  fmt.Printf("Error processing order %d for user %d\n", orderID, userID)
// GOOD: logger.Error("order processing failed", "order_id", 123, "user_id", 456, "error", err)

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

func (l LogLevel) String() string {
	return [...]string{"DEBUG", "INFO", "WARN", "ERROR"}[l]
}

type LogEntry struct {
	Timestamp string                 `json:"timestamp"`
	Level     string                 `json:"level"`
	Service   string                 `json:"service"`
	Message   string                 `json:"message"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
	TraceID   string                 `json:"trace_id,omitempty"`
	SpanID    string                 `json:"span_id,omitempty"`
	RequestID string                 `json:"request_id,omitempty"`
}

type Logger struct {
	service  string
	minLevel LogLevel
	mu       sync.Mutex
	entries  []LogEntry // in production: write to stdout/file, ship to ELK/Loki
}

func NewLogger(service string, minLevel LogLevel) *Logger {
	return &Logger{service: service, minLevel: minLevel}
}

func (l *Logger) log(level LogLevel, msg string, fields map[string]interface{}, ctx context.Context) {
	if level < l.minLevel {
		return
	}

	entry := LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Level:     level.String(),
		Service:   l.service,
		Message:   msg,
		Fields:    fields,
	}

	// Extract trace context if available
	if tc, ok := ctx.Value(traceContextKey{}).(*TraceContext); ok {
		entry.TraceID = tc.TraceID
		entry.SpanID = tc.SpanID
		entry.RequestID = tc.RequestID
	}

	l.mu.Lock()
	l.entries = append(l.entries, entry)
	l.mu.Unlock()

	// Print as JSON (in production: write to stdout, collected by Fluentd/Vector)
	jsonBytes, _ := json.Marshal(entry)
	fmt.Printf("  %s\n", jsonBytes)
}

func (l *Logger) Debug(ctx context.Context, msg string, fields map[string]interface{}) {
	l.log(DEBUG, msg, fields, ctx)
}

func (l *Logger) Info(ctx context.Context, msg string, fields map[string]interface{}) {
	l.log(INFO, msg, fields, ctx)
}

func (l *Logger) Warn(ctx context.Context, msg string, fields map[string]interface{}) {
	l.log(WARN, msg, fields, ctx)
}

func (l *Logger) Error(ctx context.Context, msg string, fields map[string]interface{}) {
	l.log(ERROR, msg, fields, ctx)
}

// =============================================================================
// PILLAR 2: Metrics (RED Method + USE Method)
// =============================================================================
//
// RED METHOD (for request-driven services — most microservices):
//   R = Rate       (requests per second)
//   E = Errors     (error count/rate)
//   D = Duration   (response time histogram)
//
// USE METHOD (for resources — CPU, memory, disk, connections):
//   U = Utilization (% of capacity used)
//   S = Saturation  (how much work is queued)
//   E = Errors      (error count)
//
// FOUR GOLDEN SIGNALS (Google SRE):
//   1. Latency       (time to serve a request)
//   2. Traffic        (demand: requests/sec)
//   3. Errors         (rate of failed requests)
//   4. Saturation     (how "full" the system is)
//
// METRIC TYPES (Prometheus):
//   Counter:   only goes UP (total requests, total errors)
//   Gauge:     goes up and down (CPU %, memory, goroutine count)
//   Histogram: distribution of values (request duration, response size)
//   Summary:   like histogram but calculates percentiles client-side

type MetricType int

const (
	CounterMetric MetricType = iota
	GaugeMetric
	HistogramMetric
)

type Metric struct {
	Name   string
	Type   MetricType
	Labels map[string]string
	Value  float64
}

type MetricsCollector struct {
	mu       sync.Mutex
	counters map[string]float64
	gauges   map[string]float64
	histos   map[string][]float64
}

func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		counters: make(map[string]float64),
		gauges:   make(map[string]float64),
		histos:   make(map[string][]float64),
	}
}

// Counter: increment only
func (mc *MetricsCollector) Inc(name string, labels map[string]string) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	key := formatMetricKey(name, labels)
	mc.counters[key]++
}

// Gauge: set to any value
func (mc *MetricsCollector) Set(name string, value float64, labels map[string]string) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	key := formatMetricKey(name, labels)
	mc.gauges[key] = value
}

// Histogram: observe a value
func (mc *MetricsCollector) Observe(name string, value float64, labels map[string]string) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	key := formatMetricKey(name, labels)
	mc.histos[key] = append(mc.histos[key], value)
}

func (mc *MetricsCollector) Report() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	fmt.Println("\n  --- Metrics Report ---")
	for k, v := range mc.counters {
		fmt.Printf("  COUNTER %s = %.0f\n", k, v)
	}
	for k, v := range mc.gauges {
		fmt.Printf("  GAUGE   %s = %.2f\n", k, v)
	}
	for k, values := range mc.histos {
		if len(values) == 0 {
			continue
		}
		sum := 0.0
		for _, v := range values {
			sum += v
		}
		avg := sum / float64(len(values))
		fmt.Printf("  HISTO   %s: count=%d avg=%.2fms\n", k, len(values), avg)
	}
}

func formatMetricKey(name string, labels map[string]string) string {
	if len(labels) == 0 {
		return name
	}
	parts := name + "{"
	first := true
	for k, v := range labels {
		if !first {
			parts += ","
		}
		parts += k + "=" + v
		first = false
	}
	return parts + "}"
}

// =============================================================================
// PILLAR 3: Distributed Tracing
// =============================================================================
//
// A TRACE follows a single request across ALL microservices.
//
// Trace ID: unique per request (generated at entry point)
// Span:     one operation within a trace (e.g., "call payment service")
// Parent:   spans form a tree (caller → callee)
//
// EXAMPLE TRACE:
//   Trace: abc-123
//   ├── Span: API Gateway (15ms)
//   │   ├── Span: Auth Service (3ms)
//   │   └── Span: Order Service (10ms)
//   │       ├── Span: DB Query (2ms)
//   │       └── Span: Payment Service (5ms)
//   │           └── Span: Stripe API Call (4ms)
//
// HOW IT WORKS:
//   1. Gateway generates Trace ID, creates root Span
//   2. Passes Trace ID + Span ID in HTTP headers (traceparent)
//   3. Each service creates a child Span
//   4. All spans are sent to Jaeger/Zipkin
//   5. Jaeger reconstructs the full trace tree
//
// OPENTELEMETRY STANDARD HEADERS:
//   traceparent: 00-<trace-id>-<span-id>-<flags>
//   tracestate:  <vendor-specific-data>

type traceContextKey struct{}

type TraceContext struct {
	TraceID   string
	SpanID    string
	ParentID  string
	RequestID string
}

type Span struct {
	TraceID   string            `json:"trace_id"`
	SpanID    string            `json:"span_id"`
	ParentID  string            `json:"parent_id,omitempty"`
	Service   string            `json:"service"`
	Operation string            `json:"operation"`
	StartTime time.Time         `json:"start_time"`
	Duration  time.Duration     `json:"duration"`
	Status    string            `json:"status"` // OK, ERROR
	Tags      map[string]string `json:"tags,omitempty"`
}

type Tracer struct {
	service string
	mu      sync.Mutex
	spans   []Span
}

func NewTracer(service string) *Tracer {
	return &Tracer{service: service}
}

func (t *Tracer) StartSpan(ctx context.Context, operation string) (context.Context, func(status string)) {
	spanID := fmt.Sprintf("span-%d", rand.Intn(10000))
	traceID := ""
	parentID := ""

	// Inherit trace context from parent
	if tc, ok := ctx.Value(traceContextKey{}).(*TraceContext); ok {
		traceID = tc.TraceID
		parentID = tc.SpanID
	} else {
		traceID = fmt.Sprintf("trace-%d", rand.Intn(100000))
	}

	start := time.Now()

	// Inject new span into context
	newCtx := context.WithValue(ctx, traceContextKey{}, &TraceContext{
		TraceID:  traceID,
		SpanID:   spanID,
		ParentID: parentID,
	})

	// Return finish function
	finish := func(status string) {
		span := Span{
			TraceID:   traceID,
			SpanID:    spanID,
			ParentID:  parentID,
			Service:   t.service,
			Operation: operation,
			StartTime: start,
			Duration:  time.Since(start),
			Status:    status,
		}
		t.mu.Lock()
		t.spans = append(t.spans, span)
		t.mu.Unlock()
	}

	return newCtx, finish
}

func (t *Tracer) PrintTrace() {
	t.mu.Lock()
	defer t.mu.Unlock()

	for _, span := range t.spans {
		indent := "  "
		if span.ParentID != "" {
			indent = "    "
		}
		fmt.Printf("%s[%s] %s/%s — %v (%s)\n",
			indent, span.TraceID, span.Service, span.Operation,
			span.Duration, span.Status)
	}
}

// =============================================================================
// CONCEPTS: Correlation IDs
// =============================================================================
//
// A Correlation ID is a unique string that follows a request through ALL services.
// Different from Trace ID:
//   - Trace ID: for tracing infrastructure (Jaeger, Zipkin)
//   - Correlation ID: for business logic (logs, error tracking, support tickets)
//
// Passed as HTTP header: X-Correlation-ID
// Every log entry includes it.
// Customer support: "Give me your request ID" → find all logs instantly.

// =============================================================================
// CONCEPTS: Structured Events vs Logs
// =============================================================================
//
// Traditional logs: println("user 123 placed order 456")
// Structured events: {user_id: 123, order_id: 456, action: "place_order", duration_ms: 45}
//
// HONEYCOMB APPROACH: wide events with 100+ fields per request.
// Instead of logging at every step, build up one rich event per request
// and emit it at the end. Query any dimension at any time.

func main() {
	// =========================================================================
	// DEMO 1: Structured Logging with Context
	// =========================================================================
	fmt.Println("=== STRUCTURED LOGGING ===")

	logger := NewLogger("order-service", INFO)

	// Create context with trace info
	ctx := context.WithValue(context.Background(), traceContextKey{}, &TraceContext{
		TraceID:   "trace-abc-123",
		SpanID:    "span-001",
		RequestID: "req-789",
	})

	logger.Info(ctx, "order created", map[string]interface{}{
		"order_id": 456,
		"user_id":  123,
		"total":    99.99,
	})

	logger.Warn(ctx, "inventory low", map[string]interface{}{
		"product_id": 42,
		"remaining":  3,
	})

	logger.Error(ctx, "payment failed", map[string]interface{}{
		"order_id": 456,
		"gateway":  "stripe",
		"error":    "card_declined",
	})

	// =========================================================================
	// DEMO 2: Metrics Collection
	// =========================================================================
	fmt.Println("\n=== METRICS (RED + USE) ===")

	metrics := NewMetricsCollector()

	// Simulate request processing with metrics
	endpoints := []string{"/users", "/orders", "/payments"}
	statuses := []string{"200", "200", "200", "500", "200"}

	for i := 0; i < 20; i++ {
		endpoint := endpoints[rand.Intn(len(endpoints))]
		status := statuses[rand.Intn(len(statuses))]
		duration := float64(rand.Intn(200) + 10) // 10-210ms

		// RED metrics
		metrics.Inc("http_requests_total", map[string]string{
			"endpoint": endpoint,
			"status":   status,
		})
		metrics.Observe("http_request_duration_ms", duration, map[string]string{
			"endpoint": endpoint,
		})

		if status == "500" {
			metrics.Inc("http_errors_total", map[string]string{
				"endpoint": endpoint,
			})
		}
	}

	// USE metrics
	metrics.Set("cpu_utilization_percent", 67.5, nil)
	metrics.Set("memory_used_bytes", 1024*1024*512, nil)
	metrics.Set("goroutines_active", 142, nil)
	metrics.Set("db_connection_pool_used", 8, map[string]string{"pool": "primary"})

	metrics.Report()

	// =========================================================================
	// DEMO 3: Distributed Tracing
	// =========================================================================
	fmt.Println("\n=== DISTRIBUTED TRACING ===")

	// Simulate a request flowing through multiple services
	gatewayTracer := NewTracer("api-gateway")
	orderTracer := NewTracer("order-service")
	paymentTracer := NewTracer("payment-service")
	dbTracer := NewTracer("database")

	// Gateway creates root span
	ctx1, finishGateway := gatewayTracer.StartSpan(context.Background(), "HandleRequest")

	// Gateway calls Order Service
	ctx2, finishOrder := orderTracer.StartSpan(ctx1, "CreateOrder")

	// Order Service queries database
	ctx3, finishDB := dbTracer.StartSpan(ctx2, "INSERT orders")
	time.Sleep(2 * time.Millisecond) // simulate DB query
	finishDB("OK")

	// Order Service calls Payment Service
	_, finishPayment := paymentTracer.StartSpan(ctx3, "ChargeCard")
	time.Sleep(5 * time.Millisecond) // simulate payment
	finishPayment("OK")

	time.Sleep(1 * time.Millisecond)
	finishOrder("OK")

	time.Sleep(1 * time.Millisecond)
	finishGateway("OK")

	fmt.Println("  Trace tree:")
	gatewayTracer.PrintTrace()
	orderTracer.PrintTrace()
	dbTracer.PrintTrace()
	paymentTracer.PrintTrace()

	// =========================================================================
	// DEMO 4: Alerting Rules (Conceptual)
	// =========================================================================
	fmt.Println("\n=== ALERTING RULES (Prometheus/Grafana) ===")
	fmt.Println("  Critical alerts:")
	fmt.Println("    - Error rate > 5% for 5 minutes          → page on-call")
	fmt.Println("    - P99 latency > 1s for 5 minutes          → page on-call")
	fmt.Println("    - Service health check DOWN for 2 minutes  → page on-call")
	fmt.Println("    - CPU > 90% for 10 minutes                → page on-call")
	fmt.Println("  Warning alerts:")
	fmt.Println("    - Error rate > 1% for 10 minutes          → Slack notification")
	fmt.Println("    - P95 latency > 500ms for 10 minutes      → Slack notification")
	fmt.Println("    - Disk > 80%                              → Slack notification")

	// =========================================================================
	// Summary
	// =========================================================================
	fmt.Println("\n=== OBSERVABILITY STACK ===")
	fmt.Println("┌──────────────────┬─────────────────────┬──────────────────────┐")
	fmt.Println("│ Pillar           │ Tool (OSS)          │ Tool (Cloud)         │")
	fmt.Println("├──────────────────┼─────────────────────┼──────────────────────┤")
	fmt.Println("│ Logging          │ ELK, Loki+Grafana   │ CloudWatch, Datadog  │")
	fmt.Println("│ Metrics          │ Prometheus+Grafana   │ Datadog, CloudWatch  │")
	fmt.Println("│ Tracing          │ Jaeger, Zipkin,Tempo │ AWS X-Ray, Datadog   │")
	fmt.Println("│ All-in-one       │ OpenTelemetry (SDK)  │ Datadog, New Relic   │")
	fmt.Println("└──────────────────┴─────────────────────┴──────────────────────┘")
	fmt.Println()
	fmt.Println("OPENTELEMETRY: The standard. Use OTel SDK to generate logs, metrics,")
	fmt.Println("traces — then export to ANY backend (Jaeger, Prometheus, Datadog).")
	fmt.Println("One instrumentation, many backends. This is the future.")
}
