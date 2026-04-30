# Module 5: Infrastructure & Reliability

> Building systems that survive everything: traffic spikes, hardware failures, human error, and acts of nature.

---

## 5.1 — Message Queues & Event Streaming

### Why Message Queues?

```
Without queue:
  [Service A] → [Service B]  (A waits, coupled, B overloaded? A fails too)

With queue:
  [Service A] → [Queue] → [Service B]  (A doesn't wait, decoupled, B processes at its own pace)
```

**Benefits:**
- **Decoupling** — Services don't know about each other
- **Buffering** — Handle traffic spikes without dropping requests
- **Async processing** — Respond to user immediately, process later
- **Retry** — Failed messages go back on the queue
- **Fan-out** — One message consumed by multiple services

### Message Queue vs Event Stream

| Feature | Message Queue | Event Stream |
|---------|--------------|--------------|
| Pattern | Point-to-point | Pub/sub, log-based |
| Consumption | Message deleted after processing | Messages persist, replayable |
| Ordering | Usually FIFO within queue | Partition-ordered |
| Example | SQS, RabbitMQ | Kafka, Kinesis, Pulsar |

### Apache Kafka — Deep Dive

Kafka is the king of event streaming. Understanding it is essential.

```
Producer → Topic → Partition 0: [msg1, msg2, msg3, ...]
                → Partition 1: [msg4, msg5, msg6, ...]
                → Partition 2: [msg7, msg8, msg9, ...]
                         ↓
                    Consumer Group
                    Consumer A reads Partition 0
                    Consumer B reads Partition 1
                    Consumer C reads Partition 2
```

**Key concepts:**
- **Topic** — A category of messages (like a table)
- **Partition** — Ordered, immutable sequence within a topic. Unit of parallelism.
- **Offset** — Position of a message within a partition
- **Consumer Group** — Set of consumers that share work. Each partition assigned to exactly one consumer in a group.
- **Replication Factor** — Each partition replicated to N brokers for durability

**Guarantees:**
- Messages within a partition are strictly ordered
- At-least-once delivery by default
- Exactly-once semantics possible (with transactions)
- Messages persisted to disk, retained for configurable period (default 7 days)

**Performance:**
- Single Kafka cluster: millions of messages/sec
- Latency: single-digit ms typical
- LinkedIn processes 7+ trillion messages/day on Kafka

### Delivery Semantics

| Guarantee | Meaning | How |
|-----------|---------|-----|
| **At-most-once** | Message may be lost, never reprocessed | Fire and forget |
| **At-least-once** | Message never lost, may be reprocessed | ACK after processing + idempotent consumer |
| **Exactly-once** | Message processed exactly once | Kafka transactions or idempotent processing |

**In practice:** Use at-least-once + idempotent consumers. Exactly-once is expensive and usually not worth it.

### Queue Technologies Comparison

| System | Type | Throughput | Ordering | Persistence | Best For |
|--------|------|-----------|----------|-------------|----------|
| **Kafka** | Log-based stream | Very high (millions/sec) | Per-partition | Disk (days-forever) | Event streaming, CDC, logs |
| **RabbitMQ** | Traditional queue | Medium (50K/sec) | Per-queue | Optional | Task queues, RPC, routing |
| **SQS** | Managed queue | High | Best-effort (FIFO variant available) | Managed (14 days) | Simple async, AWS ecosystem |
| **Pulsar** | Log-based stream | Very high | Per-partition | Tiered storage | Multi-tenancy, geo-replication |
| **NATS** | Lightweight pub/sub | Very high | Subject-scoped | JetStream | Microservices, IoT, edge |

### Dead Letter Queue (DLQ)

```
Message fails processing after N retries
  → Move to Dead Letter Queue
  → Alert engineers
  → Engineers inspect and manually reprocess or discard
```

Every production queue system must have a DLQ strategy.

---

## 5.2 — Microservices Architecture

### Monolith vs Microservices

| Aspect | Monolith | Microservices |
|--------|----------|---------------|
| Deployment | All at once | Independent per service |
| Scaling | Entire app | Individual services |
| Technology | One stack | Polyglot (different languages) |
| Communication | Function calls (fast) | Network calls (slow) |
| Data | Shared database | Database per service |
| Complexity | In the code | In the infrastructure |
| Team size | Small teams | Large organizations |
| Debugging | Easy (one process) | Hard (distributed tracing) |

### When to Use Microservices

**Use microservices when:**
- Team is 50+ engineers
- Different components have different scaling needs
- You need independent deployment
- Components have clearly different domains

**Start with a monolith when:**
- Team < 20 engineers
- Product is still being figured out
- Speed of iteration matters more than scale
- You'll split later when boundaries become clear

### API Gateway

```
Client → API Gateway → Service A
                    → Service B
                    → Service C
```

Responsibilities:
- **Routing** — Route requests to correct service
- **Authentication** — Verify tokens centrally
- **Rate limiting** — Protect backend services
- **SSL termination** — Handle HTTPS at the edge
- **Response aggregation** — Combine results from multiple services
- **Protocol translation** — REST to gRPC, etc.

**Tools:** Kong, Envoy, AWS API Gateway, Netflix Zuul

### Service Mesh

```
[Service A] → [Sidecar Proxy A] → [Sidecar Proxy B] → [Service B]
```

A dedicated infrastructure layer for service-to-service communication.

**What the sidecar handles:**
- Mutual TLS (mTLS) — encryption between services
- Load balancing
- Circuit breaking
- Retries with backoff
- Observability (metrics, traces)
- Traffic splitting (canary deployments)

**Tools:** Istio, Linkerd, Consul Connect

### Service Discovery

How does Service A find Service B?

| Method | How | Example |
|--------|-----|---------|
| **DNS-based** | Service name → DNS record → IP | Kubernetes DNS, Consul |
| **Registry-based** | Services register with central registry | Eureka, Consul catalog |
| **Platform-native** | Container orchestrator handles it | Kubernetes Services |

### Circuit Breaker Pattern

Prevents cascade failures when a downstream service is unhealthy.

```
States:
  CLOSED (normal):     Requests flow through
    → If failure rate > threshold → OPEN

  OPEN (tripped):      All requests fail immediately (fast fail)
    → After timeout period → HALF-OPEN

  HALF-OPEN (testing): Allow a few requests through
    → If they succeed → CLOSED (recovered!)
    → If they fail → OPEN (still broken)
```

**Libraries:** Hystrix (Netflix, deprecated), resilience4j, Polly (.NET), gobreaker

### Retry with Exponential Backoff + Jitter

```
attempt 1: wait 1s
attempt 2: wait 2s + random(0, 1s)
attempt 3: wait 4s + random(0, 2s)
attempt 4: wait 8s + random(0, 4s)
give up after max_retries
```

**Without jitter:** All failed clients retry at the same time → thundering herd.
**With jitter:** Retries are spread out → system recovers gracefully.

---

## 5.3 — Containerization & Orchestration

### Docker (Containers)

```
Your Code + Dependencies + Runtime = Container Image
Container Image + Container Runtime = Running Container
```

- Consistent across environments (dev = staging = prod)
- Fast startup (seconds vs minutes for VMs)
- Lightweight (shared kernel, not full OS)

### Kubernetes (K8s)

The operating system for distributed applications.

```
Cluster:
  Control Plane (API Server, Scheduler, Controller Manager, etcd)
  Worker Nodes:
    Node 1:
      Pod A (Container 1, Container 2)
      Pod B (Container 3)
    Node 2:
      Pod C (Container 4)
      Pod D (Container 5)
```

**Key concepts:**
| Concept | What it is |
|---------|-----------|
| **Pod** | Smallest deployable unit. 1+ containers sharing network/storage |
| **Deployment** | Manages replica sets, rolling updates, rollbacks |
| **Service** | Stable network endpoint for a set of pods |
| **Ingress** | External HTTP(S) routing to services |
| **ConfigMap/Secret** | Configuration and sensitive data |
| **HPA** | Horizontal Pod Autoscaler |
| **StatefulSet** | For stateful workloads (databases, Kafka) |
| **DaemonSet** | Run one pod per node (monitoring agents, log collectors) |

**Why Kubernetes matters for system design:**
- Automatic bin-packing (efficient resource use)
- Self-healing (restart failed containers)
- Horizontal scaling
- Service discovery and load balancing built-in
- Rolling updates with zero downtime
- Multi-cloud portability

---

## 5.4 — Observability

> "You can't fix what you can't see."

### Three Pillars of Observability

#### 1. Metrics (What is happening?)

Numerical measurements over time.

```
Types:
  Counter    — Monotonically increasing (total requests, errors)
  Gauge      — Value that goes up and down (CPU %, queue depth)
  Histogram  — Distribution of values (request duration percentiles)
  Summary    — Similar to histogram, calculated client-side
```

**Key metrics (RED method):**
- **R**ate — Requests per second
- **E**rrors — Error rate
- **D**uration — Latency (p50, p95, p99)

**Key metrics (USE method, for resources):**
- **U**tilization — % of resource used
- **S**aturation — Work queued
- **E**rrors — Error count

**Tools:** Prometheus + Grafana, Datadog, CloudWatch, InfluxDB

#### 2. Logs (What happened?)

Structured records of events.

```json
{
  "timestamp": "2024-01-15T10:30:00Z",
  "level": "ERROR",
  "service": "payment-service",
  "trace_id": "abc123",
  "message": "Payment failed",
  "user_id": "user_456",
  "error": "insufficient_balance",
  "duration_ms": 230
}
```

**Always use structured logging (JSON).** Not plain text.

**Tools:** ELK Stack (Elasticsearch + Logstash + Kibana), Loki + Grafana, Splunk, Datadog

**Log levels:** DEBUG → INFO → WARN → ERROR → FATAL

#### 3. Traces (How did it happen?)

Follow a single request across multiple services.

```
Request ID: abc123
  → API Gateway (2ms)
    → Auth Service (15ms)
    → Order Service (50ms)
      → Inventory Service (20ms)
      → Payment Service (150ms)  ← SLOW!
        → Stripe API (140ms)     ← External dependency
    → Notification Service (5ms)

Total: 242ms, bottleneck = Stripe API
```

**Tools:** Jaeger, Zipkin, OpenTelemetry (the standard), Datadog APM

### Alerting

**Good alerts are:**
- **Actionable** — Someone needs to do something
- **Symptomatic** — Alert on user-facing symptoms, not internal causes
- **Deduplicated** — Don't page for the same issue repeatedly

**Alert hierarchy:**
1. P99 latency > 500ms for 5 minutes → Warning
2. Error rate > 5% for 2 minutes → Page on-call
3. Error rate > 50% for 1 minute → Page everyone

---

## 5.5 — Reliability Patterns

### Failover

#### Active-Passive (Hot Standby)
```
[Active Server] ← All traffic
[Passive Server] ← Idle, replicating state from active
  If active dies → Passive takes over (seconds to minutes)
```

#### Active-Active
```
[Server A] ← Traffic from US
[Server B] ← Traffic from EU
  If either dies → Other takes all traffic
```

Active-Active is better (no wasted capacity) but harder (state synchronization).

### Graceful Degradation

When the system is overloaded, reduce functionality rather than failing completely.

```
Normal:     Full features, personalized content, real-time updates
Degraded:   Static content, cached data, reduced features
Extreme:    Static page: "We're experiencing high traffic, try again soon"
```

**Examples:**
- Netflix: Lower video quality during peak
- Twitter: Show cached timeline instead of real-time
- E-commerce: Disable non-essential features (reviews, recommendations)

### Bulkheads

Isolate components so one failure doesn't take everything down.

```
Without bulkhead:
  Shared thread pool (100 threads)
  Service A (slow) uses all 100 → Service B, C starved

With bulkhead:
  Service A: 40 threads (isolated pool)
  Service B: 30 threads (isolated pool)
  Service C: 30 threads (isolated pool)
  Service A going slow → Only Service A is affected
```

### Backpressure

When a consumer can't keep up with a producer:

| Strategy | Behavior |
|----------|----------|
| **Drop** | Discard excess messages |
| **Buffer** | Queue messages (risk OOM) |
| **Sample** | Process every Nth message |
| **Backpressure** | Tell producer to slow down |

Reactive systems (RxJava, Reactor, Akka Streams) have backpressure built-in.

### Health Checks

```
GET /health → 200 OK (shallow: "process is running")
GET /health/ready → 200 OK (deep: "can serve traffic, DB connected")
GET /health/live → 200 OK (liveness: "not deadlocked")
```

Kubernetes distinction:
- **Liveness probe:** Is the process stuck? If failed → restart container
- **Readiness probe:** Can it serve traffic? If failed → remove from load balancer
- **Startup probe:** Has it finished initializing? Avoids killing slow-starting apps

---

## 5.6 — Deployment Strategies

| Strategy | How | Risk | Speed |
|----------|-----|------|-------|
| **Big Bang** | Replace everything at once | High (all or nothing) | Fastest |
| **Rolling** | Update instances one by one | Medium | Moderate |
| **Blue-Green** | Two environments, switch traffic | Low (instant rollback) | Fast |
| **Canary** | Route 1-5% traffic to new version, gradually increase | Very low | Slow |
| **Feature Flags** | Deploy code but enable features per user/% | Lowest | Flexible |

### Canary Deployment (Preferred)

```
Step 1: Deploy v2, route 1% traffic
Step 2: Monitor error rate, latency for 15 min
Step 3: If OK → 5% → 25% → 50% → 100%
Step 4: If NOT OK → route 100% back to v1 (instant rollback)
```

### Feature Flags

```python
if feature_flags.is_enabled("new_checkout_flow", user_id=user.id):
    return new_checkout(request)
else:
    return old_checkout(request)
```

**Powers:**
- Decouple deployment from release
- A/B testing
- Kill switch for broken features
- Gradual rollout by user segment

**Tools:** LaunchDarkly, Unleash, Flagsmith, custom (Redis + config)

---

## 5.7 — Chaos Engineering

> "The best way to avoid failure is to fail constantly." — Netflix

### Principles

1. Define "steady state" (normal behavior metrics)
2. Hypothesize that steady state will hold under stress
3. Introduce real-world failures (kill servers, inject latency, corrupt data)
4. Observe the difference between hypothesis and reality
5. Fix what breaks

### Common Chaos Experiments

| Experiment | What you learn |
|-----------|---------------|
| Kill a random instance | Does the LB route around it? |
| Kill an entire AZ | Does the system survive zone failure? |
| Inject 500ms latency on DB calls | Do circuit breakers trip? |
| Fill disk to 100% | Does the app handle it gracefully? |
| Block network between services | Does the system partition gracefully? |
| Corrupt DNS | Does the app cache DNS properly? |

### Tools
- **Chaos Monkey** (Netflix) — Randomly kills instances
- **Chaos Kong** (Netflix) — Simulates entire region failure
- **LitmusChaos** — Kubernetes-native chaos engineering
- **Gremlin** — Enterprise chaos-as-a-service

---

## 5.8 — Security at Scale

### Authentication & Authorization

```
Authentication (AuthN): "Who are you?"
  → JWT tokens, OAuth 2.0, SAML

Authorization (AuthZ): "What can you do?"
  → RBAC (Role-Based), ABAC (Attribute-Based), Policy engines (OPA)
```

### JWT Flow
```
1. User logs in → Auth Service validates credentials
2. Auth Service creates JWT (signed with private key)
3. Client stores JWT, sends with every request
4. API Gateway validates JWT (with public key)
5. No DB lookup needed for validation!
```

### API Security Checklist
- Rate limiting
- Input validation (prevent injection)
- HTTPS everywhere
- Authentication on all endpoints
- Authorization checks at service level
- API key rotation
- Request signing for sensitive operations
- Audit logging

### DDoS Protection Layers

```
Layer 1: DNS (Cloudflare, Route53) → Filter at edge
Layer 2: CDN → Absorb traffic
Layer 3: Load Balancer → Rate limiting, IP blocking
Layer 4: API Gateway → Authentication, throttling
Layer 5: Application → Business logic rate limits
```

---

## 5.9 — Exercises

1. **Queue design:** Your e-commerce platform processes 1M orders/day. Orders involve: validate payment, update inventory, send confirmation email. Design the message queue architecture. What happens if payment fails?

2. **Observability:** You get an alert that p99 latency spiked from 200ms to 2 seconds. Walk through your debugging process. What tools do you use? What do you check first?

3. **Deployment:** You're deploying a breaking database schema change that affects the user service. You can't do it atomically. How do you deploy safely?

4. **Chaos engineering:** Design 3 chaos experiments for a ride-sharing app. What's the hypothesis for each? What's the expected outcome?

---

**Next:** [Module 6 — Real-World System Designs](06-real-world-designs.md) →
