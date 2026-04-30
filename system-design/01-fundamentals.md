# Module 1: Fundamentals

> Before you can design systems, you need to think in the right units.

---

## 1.1 — Numbers Every Engineer Must Know

Burn these into your brain. They shape every design decision.

### Latency Numbers (approximate, 2024)

| Operation | Time |
|-----------|------|
| L1 cache reference | 0.5 ns |
| L2 cache reference | 7 ns |
| Main memory reference | 100 ns |
| SSD random read | 16 μs |
| Read 1 MB sequentially from memory | 250 μs |
| Read 1 MB sequentially from SSD | 1 ms |
| Disk seek (HDD) | 10 ms |
| Read 1 MB sequentially from HDD | 20 ms |
| Send packet CA → Netherlands → CA | 150 ms |

### Throughput Ballparks

| Resource | Throughput |
|----------|-----------|
| Single HDD | ~100 MB/s sequential |
| Single SSD | ~500 MB/s - 3 GB/s |
| 1 Gbps network | ~125 MB/s |
| 10 Gbps network | ~1.25 GB/s |
| Single web server (API) | 1K-10K requests/sec |
| Single Redis instance | ~100K ops/sec |
| Single PostgreSQL | ~10K-50K queries/sec (depends on query complexity) |

### Back-of-the-Envelope Estimations

This is a **critical skill**. You must be able to quickly estimate:

**Example: How much storage does Twitter need for tweets per day?**
- 500M tweets/day
- Average tweet: 280 chars × 2 bytes = ~560 bytes
- Metadata (user_id, timestamp, etc.): ~200 bytes
- Total per tweet: ~760 bytes ≈ 1 KB (round up)
- Daily: 500M × 1 KB = 500 GB/day
- Yearly: 500 GB × 365 = ~180 TB/year (just text, no media)

**Powers of 2 you must know:**
```
2^10 = 1 Thousand    (1 KB)
2^20 = 1 Million     (1 MB)
2^30 = 1 Billion     (1 GB)
2^40 = 1 Trillion    (1 TB)
```

**Time conversions:**
```
1 day    = 86,400 seconds ≈ 10^5 seconds
1 month  = 2.6M seconds   ≈ 2.5 × 10^6
1 year   = 31.5M seconds  ≈ 3 × 10^7
```

**Practice:** Before reading on, estimate: If YouTube gets 500 hours of video uploaded per minute, and 1 minute of 1080p video = 150 MB, how much raw storage per day?

*Answer: 500 hrs/min × 60 min × 150 MB × 60 min = 500 × 60 × 60 × 150 MB = 270 TB/day (before compression/encoding)*

---

## 1.2 — Availability & Reliability

### Availability = Uptime / (Uptime + Downtime)

Measured in "nines":

| Availability | Downtime/year | Downtime/month |
|-------------|---------------|----------------|
| 99% (two nines) | 3.65 days | 7.3 hours |
| 99.9% (three nines) | 8.76 hours | 43.8 minutes |
| 99.99% (four nines) | 52.6 minutes | 4.38 minutes |
| 99.999% (five nines) | 5.26 minutes | 26.3 seconds |

**Key insight:** Going from 99.9% to 99.99% is *10x harder* and *10x more expensive* than going from 99% to 99.9%. Diminishing returns are brutal.

### Availability in Series vs Parallel

**Series (both must work):**
```
A(total) = A(1) × A(2)
Example: Service A (99.9%) → Service B (99.9%)
Total = 0.999 × 0.999 = 99.8%  (worse!)
```

**Parallel (one must work):**
```
A(total) = 1 - (1 - A(1)) × (1 - A(2))
Example: Server A (99.9%) || Server B (99.9%)
Total = 1 - (0.001 × 0.001) = 99.9999%  (much better!)
```

**Design implication:** Every component in your critical path multiplies failure probability. Reduce the chain and add redundancy.

### SLA vs SLO vs SLI

| Term | Meaning | Example |
|------|---------|---------|
| **SLI** (indicator) | The metric you measure | p99 latency, error rate |
| **SLO** (objective) | Internal target for SLI | "p99 latency < 200ms" |
| **SLA** (agreement) | Contract with customers, with penalties | "99.95% uptime or we refund 10%" |

---

## 1.3 — CAP Theorem

The most important theorem in distributed systems.

### Statement

In a distributed data store, you can only guarantee **two out of three**:

- **C**onsistency — Every read receives the most recent write
- **A**vailability — Every request receives a response (not necessarily the latest data)
- **P**artition tolerance — System continues operating despite network partitions

### The Real-World Truth

**You don't actually "choose 2 of 3."** Network partitions *will* happen. They're not optional. So the real choice is:

> **When a partition occurs, do you sacrifice Consistency or Availability?**

| Choice | Name | Example | Behavior during partition |
|--------|------|---------|--------------------------|
| CP | Consistent | ZooKeeper, HBase, MongoDB (default) | Rejects writes/reads to maintain consistency |
| AP | Available | Cassandra, DynamoDB, CouchDB | Serves requests but may return stale data |

### PACELC (the better model)

CAP only describes behavior *during* partitions. PACELC extends it:

> **If Partition → choose A or C. Else → choose Latency or Consistency.**

| System | During Partition | Normal Operation |
|--------|-----------------|-----------------|
| DynamoDB | PA (available) | EL (low latency) |
| MongoDB | PC (consistent) | EC (consistent) |
| Cassandra | PA (available) | EL (low latency) |
| PostgreSQL (single node) | N/A | EC (consistent) |

---

## 1.4 — Networking Fundamentals for System Design

You don't need to be a networking expert, but you need these concepts.

### DNS (Domain Name System)

Converts `www.google.com` → `142.250.80.4`

```
User → Local DNS Resolver → Root DNS → TLD DNS (.com) → Authoritative DNS → IP Address
```

**Why it matters for system design:**
- DNS can be used for load balancing (return different IPs)
- DNS TTL affects how fast failover works
- GeoDNS routes users to nearest data center

### HTTP/HTTPS

- **HTTP/1.1** — One request per connection (head-of-line blocking)
- **HTTP/2** — Multiplexed streams over single connection
- **HTTP/3** — Built on QUIC (UDP), faster handshake, no head-of-line blocking

**Status codes you must know:**
```
200 OK           — Success
201 Created      — Resource created
301 Moved        — Permanent redirect (cacheable)
302 Found        — Temporary redirect
400 Bad Request  — Client error
401 Unauthorized — Not authenticated
403 Forbidden    — Authenticated but not authorized
404 Not Found    — Resource doesn't exist
429 Too Many     — Rate limited
500 Server Error — Something broke on the server
503 Unavailable  — Server overloaded or in maintenance
```

### TCP vs UDP

| Property | TCP | UDP |
|----------|-----|-----|
| Reliability | Guaranteed delivery, ordering | Best effort, no ordering |
| Speed | Slower (handshake, ack) | Faster (fire and forget) |
| Use case | HTTP, database connections | Video streaming, DNS, gaming |

### WebSockets

- Full-duplex communication over a single TCP connection
- Client and server can send messages at any time
- Perfect for: chat, live notifications, collaborative editing, real-time dashboards

### Long Polling vs Server-Sent Events vs WebSockets

| Method | Direction | Use Case |
|--------|-----------|----------|
| Long Polling | Server → Client (pull) | Simple notifications, compatibility |
| SSE (Server-Sent Events) | Server → Client (push) | Live feeds, stock tickers |
| WebSockets | Bidirectional | Chat, gaming, collaboration |

---

## 1.5 — API Design Paradigms

### REST

```
GET    /users/123        — Read user 123
POST   /users            — Create a user
PUT    /users/123        — Replace user 123
PATCH  /users/123        — Partially update user 123
DELETE /users/123        — Delete user 123
```

**Principles:**
- Stateless — Each request contains all needed info
- Resource-oriented — URLs represent nouns, not verbs
- Cacheable — GET responses can be cached

**When to use:** Most CRUD APIs, public APIs, when cacheability matters.

### GraphQL

```graphql
query {
  user(id: "123") {
    name
    email
    posts(first: 5) {
      title
      likes
    }
  }
}
```

**Advantages:** Client specifies exactly what data it needs. No over-fetching.
**Disadvantages:** Harder to cache, complex queries can be expensive, N+1 problem.
**When to use:** Complex, nested data models; mobile apps (bandwidth sensitive); multiple frontend clients needing different data shapes.

### gRPC

- Uses Protocol Buffers (binary serialization)
- 10x faster than JSON REST
- Strongly typed contracts
- Supports streaming (unary, server, client, bidirectional)

**When to use:** Internal service-to-service communication, low-latency requirements, polyglot environments.

### Comparison

| Feature | REST | GraphQL | gRPC |
|---------|------|---------|------|
| Format | JSON | JSON | Protobuf (binary) |
| Speed | Medium | Medium | Fast |
| Caching | Easy (HTTP caching) | Hard | Hard |
| Typing | Weak | Strong (schema) | Strong (proto) |
| Streaming | No (needs WebSocket) | Subscriptions | Native |
| Best for | Public APIs | Flexible frontends | Internal services |

---

## 1.6 — Proxies

### Forward Proxy
- Sits in front of **clients**
- Client → Forward Proxy → Internet → Server
- Use: VPN, content filtering, anonymity

### Reverse Proxy
- Sits in front of **servers**
- Client → Internet → Reverse Proxy → Server
- Use: Load balancing, SSL termination, caching, DDoS protection
- Examples: Nginx, HAProxy, Envoy, Cloudflare

---

## 1.7 — Consistent Hashing

**The problem:** You have N cache servers. Simple `hash(key) % N` breaks when you add/remove a server — almost every key remaps.

**The solution:** Consistent hashing.

### How It Works

1. Imagine a ring (hash space 0 to 2^32)
2. Hash each server to a point on the ring
3. Hash each key to a point on the ring
4. Each key is assigned to the first server clockwise from it

**When a server is added:** Only keys between the new server and the previous server move.
**When a server is removed:** Only its keys move to the next server.

### Virtual Nodes

Problem: With few physical servers, distribution is uneven.
Solution: Each physical server gets multiple "virtual nodes" on the ring.

```
Physical Server A → Virtual nodes: A1, A2, A3, A4, A5
Physical Server B → Virtual nodes: B1, B2, B3, B4, B5
```

This gives much more even distribution.

**Used by:** DynamoDB, Cassandra, Memcached, Akamai CDN, Discord.

---

## 1.8 — Hashing & Checksums

| Algorithm | Output | Use Case | Note |
|-----------|--------|----------|------|
| MD5 | 128-bit | Checksums (NOT security) | Broken for security |
| SHA-256 | 256-bit | Data integrity, blockchain | Secure |
| MurmurHash | 32/128-bit | Hash tables, consistent hashing | Very fast, not cryptographic |
| xxHash | 32/64/128-bit | Checksums, deduplication | Extremely fast |
| CRC32 | 32-bit | Network error detection | Fast, weak |

---

## 1.9 — Exercises

1. **Estimate:** You're building a photo-sharing app. 10M DAU, each user uploads 2 photos/day, average photo is 2 MB. How much storage per year? What's the bandwidth needed?

2. **Calculate availability:** Your system has 3 services in series (each 99.9%) and the database has 2 replicas in parallel (each 99.9%). What's the total availability?

3. **Choose CP or AP:** You're building a banking system. During a network partition, should you reject transactions (CP) or allow them and reconcile later (AP)? What about a social media feed?

4. **API Design:** Design REST endpoints for a ride-sharing app. What resources do you expose? Which operations on each?

---

**Next:** [Module 2 — Scaling Patterns](02-scaling-patterns.md) →
