# Microservices Mastery — God-Level Architecture Patterns in Go

Every microservices pattern that matters, implemented as runnable Go code with heavy commentary.

## Modules

| # | Module | File | What You Learn |
|---|--------|------|----------------|
| 01 | **Fundamentals** | `01_decomposition.go` | Monolith vs micro, decomposition strategies, DDD, when NOT to use microservices |
| 02 | **Sync Communication** | `01_rest_service.go` | REST patterns, HTTP clients, idempotency, API versioning, gRPC comparison |
| 03 | **Async Communication** | `01_event_driven.go` | Event-driven architecture, pub/sub, message queues, delivery guarantees, broker comparison |
| 04 | **Data Patterns** | `01_data_patterns.go` | Saga (orchestration/choreography), CQRS, Event Sourcing, API composition, DB-per-service |
| 05 | **Resilience** | `01_resilience_patterns.go` | Circuit breaker, retry+backoff, bulkhead, rate limiter, timeout, fallback, health checks |
| 06 | **Service Discovery** | `01_service_discovery.go` | Registry, health checks, round robin, weighted LB, consistent hashing, client vs server-side |
| 07 | **API Gateway** | `01_api_gateway.go` | Gateway pattern, middleware chain, BFF, request aggregation, auth at edge |
| 08 | **Observability** | `01_observability.go` | Structured logging, RED/USE metrics, distributed tracing, OpenTelemetry, alerting |
| 09 | **Security** | `01_security.go` | JWT, API keys, RBAC, mTLS, OAuth2/OIDC, zero trust architecture |
| 10 | **Testing** | `01_testing_strategies.go` | Unit/mocks/fakes, contract testing, integration, chaos engineering, testing pyramid |
| 11 | **Deployment** | `01_deployment_patterns.go` | Docker, Kubernetes, blue-green, canary, service mesh, GitOps, CI/CD |
| 12 | **Advanced Patterns** | `01_advanced_patterns.go` | Strangler fig, sidecar, ambassador, ACL, outbox, CDC, distributed locking |

## How to Run

```bash
# Run any module
cd 05-resilience
go run 01_resilience_patterns.go

# Or from project root
go run ./03-async-communication/01_event_driven.go
```

## Learning Order

```
Recommended progression:

01 Fundamentals ──────────────────────────────────┐
                                                   ▼
02 Sync Communication ──► 03 Async Communication ──► 04 Data Patterns
                                                   │
                                                   ▼
05 Resilience ──────────► 06 Service Discovery ──► 07 API Gateway
                                                   │
                                                   ▼
08 Observability ──────► 09 Security ────────────► 10 Testing
                                                   │
                                                   ▼
11 Deployment ─────────► 12 Advanced Patterns ──► GOD MODE 🏆
```

## Quick Reference

| Problem | Pattern | Module |
|---------|---------|--------|
| Breaking apart a monolith | Decomposition strategies, Strangler Fig | 01, 12 |
| Cross-service transactions | Saga (orchestration/choreography) | 04 |
| Read/write optimization | CQRS | 04 |
| Complete audit trail | Event Sourcing | 04 |
| Service keeps failing | Circuit Breaker | 05 |
| Traffic spikes | Rate Limiter, Bulkhead | 05 |
| Finding services | Service Discovery + Load Balancing | 06 |
| Client-specific APIs | BFF (Backend for Frontend) | 07 |
| Request tracing | Distributed Tracing + Correlation IDs | 08 |
| Service-to-service auth | mTLS, JWT, OAuth2 | 09 |
| API compatibility | Contract Testing | 10 |
| Zero-downtime deploys | Canary, Blue-Green | 11 |
| Reliable event publishing | Outbox + CDC | 12 |
| Legacy integration | Anti-Corruption Layer | 12 |

## God-Level Stack

```
Event-Driven + CQRS + Event Sourcing + Saga + Outbox + CDC
+ Service Mesh (Istio) + GitOps (ArgoCD) + Observability (OpenTelemetry)
= The ultimate microservices architecture
```
