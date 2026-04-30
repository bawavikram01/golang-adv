# Module 2: Scaling Patterns

> How to go from 1 user to 1 billion users.

---

## 2.1 — Vertical vs Horizontal Scaling

### Vertical Scaling (Scale Up)
- Bigger machine: more CPU, RAM, disk
- **Pros:** Simple, no code changes, strong consistency
- **Cons:** Hardware limits (you can't buy a 10,000-core machine), single point of failure, expensive at high end
- **When:** Databases (up to a point), simple apps, startups

### Horizontal Scaling (Scale Out)
- More machines of the same size
- **Pros:** Nearly infinite scale, redundancy built-in, commodity hardware
- **Cons:** Complexity (distributed state, network failures), eventual consistency challenges
- **When:** Web servers, caches, storage (anything stateless scales horizontally trivially)

### The Golden Rule

> **Make your compute stateless, push state to dedicated systems (databases, caches, queues).**

Stateless services can be cloned infinitely behind a load balancer. Stateful services require careful partitioning.

---

## 2.2 — Load Balancing

A load balancer distributes incoming traffic across multiple servers.

### Where Load Balancers Live

```
Internet → DNS LB → L4/L7 Load Balancer → Web Servers → Internal LB → App Servers → DB
```

### Layer 4 vs Layer 7

| Feature | L4 (Transport) | L7 (Application) |
|---------|----------------|-------------------|
| Operates on | TCP/UDP packets | HTTP requests |
| Speed | Faster (no payload inspection) | Slower (reads headers/body) |
| Routing | IP + port based | URL path, headers, cookies |
| SSL termination | No | Yes |
| Example | AWS NLB, IPVS | Nginx, HAProxy, AWS ALB, Envoy |

### Load Balancing Algorithms

| Algorithm | How it Works | Best For |
|-----------|-------------|----------|
| **Round Robin** | Each server gets a turn | Homogeneous servers, stateless |
| **Weighted Round Robin** | Servers get turns proportional to weight | Mixed hardware |
| **Least Connections** | Route to server with fewest active connections | Varying request durations |
| **Least Response Time** | Route to fastest responding server | Latency-sensitive |
| **IP Hash** | Hash(client_ip) → server | Sticky sessions without cookies |
| **Consistent Hashing** | Hash ring distribution | Cache servers, distributed stores |
| **Random** | Pick a random server | Surprisingly effective at scale |
| **Power of Two Choices** | Pick 2 random, choose the one with fewer connections | Best of random + least connections |

### Health Checks

Load balancers must know which backends are healthy:
- **Passive:** Monitor responses (5xx errors, timeouts)
- **Active:** Periodically ping a health endpoint (`GET /health`)
- **Deep health checks:** Verify downstream dependencies too

### Session Persistence (Sticky Sessions)

**Problem:** User logs in on Server A, next request goes to Server B — session lost.

**Solutions (from best to worst):**
1. **Stateless sessions** — Store session in client (JWT) or shared store (Redis) ← **Best**
2. **Consistent hashing** — Same user always hits same server
3. **Cookie-based affinity** — LB sets a cookie mapping to server
4. **IP affinity** — Hash(client_ip) → server (breaks with NAT/proxy)

### Global Server Load Balancing (GSLB)

For multi-region deployments:
- **GeoDNS** — Route users to nearest data center based on IP geolocation
- **Anycast** — Multiple servers share same IP; network routes to nearest one
- **Latency-based routing** — Route to the region with lowest measured latency

---

## 2.3 — Caching

> "There are only two hard things in Computer Science: cache invalidation and naming things." — Phil Karlton

### Cache Hierarchy

```
Client Cache (browser, app) → CDN → API Gateway Cache → Application Cache (Redis) → Database Cache (query cache)
```

Each layer closer to the user is faster but harder to invalidate.

### Caching Strategies

#### 1. Cache-Aside (Lazy Loading)
```
Read:
  1. Check cache
  2. Cache miss → Read from DB
  3. Write to cache
  4. Return data

Write:
  1. Write to DB
  2. Invalidate/delete cache entry
```
- **Pros:** Only requested data is cached, cache failure doesn't break reads
- **Cons:** Cache miss penalty (3 calls), data can be stale
- **Used by:** Most applications

#### 2. Read-Through
```
Read:
  1. Check cache
  2. Cache miss → Cache itself loads from DB
  3. Return data
```
- Similar to cache-aside but the cache library handles DB loading
- Simplifies application code

#### 3. Write-Through
```
Write:
  1. Write to cache
  2. Cache synchronously writes to DB
  3. Return success
```
- **Pros:** Cache is always consistent with DB
- **Cons:** Write latency (2 writes), cache may store data that's never read

#### 4. Write-Behind (Write-Back)
```
Write:
  1. Write to cache
  2. Return success immediately
  3. Cache asynchronously writes to DB (batched)
```
- **Pros:** Very fast writes, DB writes can be batched
- **Cons:** Data loss risk if cache crashes before flushing

#### 5. Refresh-Ahead
```
Cache proactively refreshes entries before they expire
(based on access patterns and TTL)
```
- **Pros:** Reduced latency for hot data
- **Cons:** Wasted resources if prediction is wrong

### Cache Eviction Policies

| Policy | Strategy | Best For |
|--------|----------|----------|
| **LRU** (Least Recently Used) | Evict oldest accessed item | General purpose (most common) |
| **LFU** (Least Frequently Used) | Evict least accessed item | When frequency matters (CDN) |
| **FIFO** (First In First Out) | Evict oldest inserted item | Time-sensitive data |
| **TTL** (Time To Live) | Evict after fixed time | Data with known staleness window |
| **Random** | Evict random item | When LRU overhead is too high |

### Cache Invalidation Patterns

This is the **hardest part** of caching.

| Pattern | Approach | Consistency |
|---------|----------|-------------|
| **TTL-based** | Data expires after N seconds | Eventually consistent |
| **Event-driven** | DB change triggers cache invalidation | Near real-time |
| **Write-invalidate** | Delete cache on write | Strong (next read is fresh) |
| **Write-update** | Update cache on write | Strong (but race condition risk) |
| **Version-based** | Cache key includes version number | Strong |

### Redis vs Memcached

| Feature | Redis | Memcached |
|---------|-------|-----------|
| Data structures | Strings, lists, sets, sorted sets, hashes, streams | Strings only |
| Persistence | RDB snapshots, AOF log | None |
| Replication | Built-in primary-replica | None (client-side) |
| Clustering | Redis Cluster (auto-sharding) | Client-side sharding |
| Memory efficiency | Less efficient | More efficient for simple KV |
| Use case | Feature-rich caching, queues, pub/sub | Simple, high-throughput caching |

**Default choice: Redis** (unless you specifically need Memcached's simplicity and memory efficiency for pure string caching).

### Caching Anti-Patterns

1. **Cache stampede** — Cache expires, 1000 requests hit DB simultaneously
   - Fix: Locking, staggered TTLs, refresh-ahead
2. **Hot key** — One key gets 90% of traffic
   - Fix: Replicate across shards, local caching
3. **Cache penetration** — Queries for non-existent data always miss
   - Fix: Cache null results, Bloom filter in front of cache
4. **Big key** — One cache entry is huge (e.g., 100MB)
   - Fix: Split into chunks, compress

---

## 2.4 — Content Delivery Networks (CDN)

### What is a CDN?

A geographically distributed network of proxy servers that cache content **close to end users**.

```
User in Tokyo → Tokyo CDN Edge → Cache HIT → Return immediately (5ms)
User in Tokyo → Tokyo CDN Edge → Cache MISS → Origin in US → Return (200ms), cache for next user
```

### Push vs Pull CDN

| Type | How it works | Best for |
|------|-------------|----------|
| **Pull** | CDN fetches from origin on first request, caches it | Dynamic content, large catalogs |
| **Push** | You upload content to CDN proactively | Static assets you control (JS, CSS, images) |

### What to Put on a CDN

- Static assets (images, CSS, JS, fonts)
- Video/audio files
- API responses (with appropriate cache headers)
- HTML pages (for static or pre-rendered sites)

### CDN Providers & Scale
- **Cloudflare** — 300+ PoPs, free tier, DDoS protection
- **AWS CloudFront** — Integrated with AWS, Lambda@Edge
- **Akamai** — Largest CDN, serves ~30% of web traffic
- **Fastly** — Edge computing (Compute@Edge), real-time purging

---

## 2.5 — Database Scaling

### Read Replicas

```
Writes → Primary DB
Reads  → Replica 1, Replica 2, Replica 3
```

- Replication: primary writes propagated to replicas
- **Synchronous replication:** strong consistency, higher latency
- **Asynchronous replication:** eventual consistency, lower latency
- Typical: 1 primary + 2-5 read replicas handles most read-heavy workloads

### Connection Pooling

Opening DB connections is expensive. Use pools:
- **PgBouncer** for PostgreSQL
- **ProxySQL** for MySQL
- Application-level pools (HikariCP for Java, etc.)

A single PostgreSQL server can handle ~5,000 connections max. With PgBouncer, you can multiplex 50,000 app connections over 100 real DB connections.

### Database Proxy

```
App Servers → Database Proxy (ProxySQL/PgBouncer) → Primary + Replicas
```

The proxy routes reads to replicas and writes to primary transparently.

---

## 2.6 — Rate Limiting

Protects your system from abuse, DDoS, and cascade failures.

### Algorithms

#### 1. Token Bucket
```
- Bucket holds N tokens (capacity)
- Tokens added at rate R per second
- Each request consumes 1 token
- No tokens? Request denied
```
- **Pros:** Allows bursts (up to bucket capacity), smooth rate
- **Used by:** AWS, Stripe

#### 2. Leaky Bucket
```
- Requests enter a queue (bucket)
- Processed at fixed rate R
- Queue full? Request dropped
```
- **Pros:** Perfectly smooth output rate
- **Cons:** No burst handling

#### 3. Fixed Window
```
- Count requests in fixed time window (e.g., per minute)
- Exceeds limit? Reject until next window
```
- **Pros:** Simple
- **Cons:** Burst at window boundaries (2x limit possible across 2 windows)

#### 4. Sliding Window Log
```
- Keep timestamp of each request
- Count requests in past N seconds
- Exceeds limit? Reject
```
- **Pros:** Accurate
- **Cons:** Memory intensive (storing all timestamps)

#### 5. Sliding Window Counter
```
- Combine current window count + weighted previous window count
- count = current_count + previous_count × overlap_percentage
```
- **Pros:** Accurate + memory efficient
- **Used by:** Most production systems

### Rate Limiting in Distributed Systems

Challenge: 10 API servers, how do they share rate limit state?

| Approach | How | Trade-off |
|----------|-----|-----------|
| **Centralized (Redis)** | All servers check Redis counter | Accurate but Redis becomes bottleneck |
| **Local + sync** | Each server has local counter, periodically syncs | Fast but slightly inaccurate |
| **Token bucket in Redis** | Distributed token bucket using Redis atomic ops | Good balance |

### Response Headers

```
X-RateLimit-Limit: 100        # Max requests per window
X-RateLimit-Remaining: 47     # Requests left
X-RateLimit-Reset: 1625097600 # When the window resets (Unix timestamp)
Retry-After: 30               # Seconds to wait (on 429 response)
```

---

## 2.7 — Auto-Scaling

### Types

| Type | Scale based on | Example |
|------|---------------|---------|
| **Reactive** | Current metrics exceed threshold | CPU > 70% for 5 min → add instance |
| **Predictive** | Historical patterns | Scale up before Monday morning traffic |
| **Scheduled** | Known events | Scale up before Super Bowl kickoff |

### Key Metrics for Auto-Scaling

- CPU utilization
- Memory utilization
- Request count / throughput
- Queue depth
- Custom metrics (e.g., processing lag)

### Scaling Policies

- **Step scaling:** Add 1 server if CPU > 60%, add 3 if CPU > 80%
- **Target tracking:** "Keep average CPU at 60%" — auto-scaler figures out how many instances

### Cool-Down Period

After scaling, wait N seconds before allowing another scaling action. Prevents thrashing (rapid scale up/down cycles).

---

## 2.8 — The Scaling Playbook

Here's how to actually scale a system, step by step:

### Stage 1: Single Server (0 → 1K users)
```
[Users] → [Single Server: Web + App + DB]
```
- Monolith is fine
- Single database, single server
- Focus on shipping features

### Stage 2: Separate DB (1K → 10K users)
```
[Users] → [Web Server] → [Database Server]
```
- Separate compute from storage
- Add database backups

### Stage 3: Load Balancer + Multiple Servers (10K → 100K users)
```
[Users] → [Load Balancer] → [Server 1, Server 2, Server 3]
                                     ↓
                              [Primary DB] → [Read Replica]
```
- Stateless app servers behind LB
- Read replicas for read-heavy workload
- Add Redis cache

### Stage 4: CDN + Caching (100K → 1M users)
```
[Users] → [CDN] → [Load Balancer] → [App Servers]
                                          ↓
[Redis Cache] ← → [Primary DB] → [Replicas × 3]
```
- CDN for static assets
- Redis for hot data
- Multiple read replicas
- Database connection pooling

### Stage 5: Microservices + Message Queues (1M → 10M users)
```
[CDN] → [API Gateway] → [Service A] → [Queue] → [Service B]
                      → [Service C] → [DB C]
                      → [Service D] → [DB D]
```
- Break monolith into services
- Async processing via queues
- Each service has its own DB
- Add monitoring and observability

### Stage 6: Sharding + Multi-Region (10M → 1B users)
```
Region US:
  [CDN] → [LB] → [Services] → [Sharded DB (shards 0-99)]

Region EU:
  [CDN] → [LB] → [Services] → [Sharded DB (shards 100-199)]

Region Asia:
  [CDN] → [LB] → [Services] → [Sharded DB (shards 200-299)]

Cross-region replication for disaster recovery
```
- Database sharding
- Multi-region deployment
- Global load balancing (GeoDNS/Anycast)
- Chaos engineering for resilience testing

---

## 2.9 — Exercises

1. **Design a caching strategy:** You're building a social media feed. Posts are read 100x more than written. What caching strategy do you use? How do you handle cache invalidation when a user edits a post?

2. **Choose a LB algorithm:** You have 3 servers: one beefy (32 cores) and two small (8 cores). Which algorithm and why?

3. **Rate limiting:** Your API gets 10K req/sec normally but during a flash sale, it spikes to 500K req/sec. How do you protect your backend?

4. **Scaling estimation:** Your app has 50M DAU with an average of 10 API calls per user per day, concentrated in 8 peak hours. What's the peak request rate? How many app servers do you need if each handles 5K req/sec?

---

**Next:** [Module 3 — Data Systems](03-data-systems.md) →
