# Backend Engineering: Zero to God Mode
### 32 Projects to Become an Elite Backend Developer

---

## How This Works
- Each project is chosen to teach a **specific backend skill** you'll use in production.
- Go in order — each project builds on the last.
- Languages: **C++** (low-level/systems projects) + **Go** (everything else). That's it.
- You already know C++ from DSA and Go for backend — go deep, not wide.
- Push everything to GitHub with READMEs.

---

## Phase 1: The Language & Fundamentals (Weeks 1–6)
*You can't be a god if your foundation is sand.*

### P1. Custom Shell (C++)
- Parse commands, fork processes, handle pipes (`|`), redirection (`>`, `<`), background jobs (`&`), signal handling (Ctrl+C).
- **Why:** Forces you to understand processes, file descriptors, syscalls — the OS layer every backend runs on. C++ gives you the same syscall access as C with a language you already know.
- **Done when:** `ls -la | grep .go | wc -l > count.txt` works in your shell.

### P2. Memory Allocator (C++)
- Implement `malloc()`, `free()`, `realloc()` using `sbrk`/`mmap`. Support: first-fit, best-fit, and buddy allocation strategies. Handle alignment, coalescing free blocks, and splitting.
- **Why:** This is how the heap works. Every time you call `new` or `make`, this is what happens underneath. Understanding virtual memory, page tables, and fragmentation makes you dangerous.
- **Done when:** Passes a stress test with 100K+ alloc/free cycles, no memory leaks, and fragmentation stays bounded. Benchmark against glibc `malloc`.

### P3. Mini OS Kernel (C++ + Assembly)
- Boot from scratch in QEMU: enter protected/long mode, set up GDT/IDT, handle interrupts (keyboard, timer), physical + virtual memory manager (page allocator + paging), basic process scheduler (round-robin), simple filesystem (read files from a ramdisk).
- **Resource:** [OSDev Wiki](https://wiki.osdev.org/), [os-tutorial](https://github.com/cfenollosa/os-tutorial).
- **Why:** This is the ultimate OS flex. You'll understand context switching, page tables, interrupt handling, and the boot process — things most devs treat as magic. In interviews, this is a cheat code.
- **Done when:** Boots in QEMU, handles keyboard input, runs 2+ processes with preemptive scheduling, and reads files from a ramdisk.

### P4. Concurrent Task Runner (Go)
- A CLI tool that takes a list of shell commands and runs them concurrently with a configurable worker pool. Handle timeouts, retries, and output ordering.
- **Why:** Goroutines, channels, sync primitives, context cancellation — Go's concurrency model is *the* backend model.
- **Done when:** Runs 100 tasks across 10 workers with proper error handling and graceful shutdown.

### P5. Data Structures You'll Actually Use (Go)
- Implement: hash map (with open addressing), LRU cache, priority queue (heap), bloom filter, ring buffer, skip list.
- **Why:** These show up constantly in backend systems. You need to *feel* how they work, not just call library functions. Build them in Go since that's your backend language.
- **Done when:** Each has benchmarks + unit tests. You can explain the time/space trade-offs.

---

## Phase 2: Networking & Protocols (Weeks 7–12)
*Backend is networking. Master the wire.*

### P6. HTTP/1.1 Server from Scratch (C++)
- Parse HTTP requests, route to handlers, serve static files, handle keep-alive, chunked transfer encoding.
- No frameworks, no libraries — raw TCP sockets only.
- **Why:** You'll never be confused by an HTTP bug again. You'll *know* what's on the wire. C++ gives you raw socket control and teaches you exactly what Go's `net/http` does underneath.
- **Done when:** Serves a website, passes `curl` and `wrk` stress tests, handles 5K concurrent connections.

### P7. HTTP/2 + WebSocket Upgrade (Go)
- Build an HTTP/2 + WebSocket server using Go's `net` package (no `net/http`): HTTP/2 multiplexing, server push, WebSocket handshake + bidirectional messaging.
- **Why:** Modern backends speak HTTP/2. Real-time features need WebSockets. Know both protocols cold.
- **Done when:** A chat demo runs over WebSockets on your server, and HTTP/2 streams work with `h2load`.

### P8. Build a Redis Clone (C++)
- RESP protocol parser, event loop (epoll/kqueue), key-value storage, expiry (TTL), pub/sub, persistence (RDB snapshots + AOF).
- Support: `GET`, `SET`, `DEL`, `EXPIRE`, `LPUSH`, `LPOP`, `SUBSCRIBE`, `PUBLISH`.
- **Why:** Redis is *the* backend tool. Building it teaches you event loops, protocol design, in-memory data stores, and persistence strategies. C++ gives you manual memory control and epoll experience.
- **Done when:** Your Redis clone passes the official `redis-benchmark` tool and survives restarts without data loss.

### P9. RPC Framework (Go)
- Binary protocol (like Protocol Buffers), code generation from IDL, TCP transport with connection pooling, middleware support (logging, tracing, auth), load balancing (round-robin).
- **Why:** Microservices communicate via RPC. Build one to understand gRPC, Thrift, etc. at a deep level.
- **Done when:** Two services communicate through your RPC framework with serialization, retries, and timeouts.

---

## Phase 3: Databases & Storage Engines (Weeks 13–20)
*A backend dev who doesn't understand databases is just a CRUD monkey.*

### P10. Write-Ahead Log (Go)
- Append-only log with: fsync guarantees, segment rotation, log compaction, CRC checksums for corruption detection.
- **Why:** WAL is the foundation of every database. Understand it and you understand durability.
- **Done when:** Survives `kill -9` at any point and recovers all committed data.

### P11. Key-Value Storage Engine (Go)
- LSM tree: memtable (red-black tree or skip list) → immutable memtable → SSTable flush → compaction → bloom filters for read optimization.
- **Why:** This is how LevelDB, RocksDB, Cassandra, and many modern DBs store data.
- **Done when:** Read/write benchmarks, survives crashes, compaction runs correctly.

### P12. B+ Tree Storage Engine (Go)
- Disk-based B+ tree with: page management, split/merge, range scans, buffer pool with LRU eviction.
- **Why:** This is how PostgreSQL, MySQL, and SQLite store data. The other side of the coin from LSM.
- **Done when:** Supports millions of keys, range queries are fast, pages are properly managed.

### P13. SQL Database (Go)
- Tokenizer → Parser → Query Planner → Executor → Storage (use your B+ tree or LSM engine).
- Support: `CREATE TABLE`, `INSERT`, `SELECT ... WHERE ... ORDER BY ... LIMIT`, `JOIN` (nested loop + hash join), `CREATE INDEX`, `BEGIN`/`COMMIT`/`ROLLBACK`.
- **Why:** You'll understand query plans, why some queries are slow, how indexes work, and how transactions actually function.
- **Done when:** Runs complex queries with joins and indexes. `EXPLAIN` shows the query plan.

### P14. Connection Pooler (Go)
- Like PgBouncer: pool PostgreSQL connections, handle transaction/session/statement modes, health checking, idle timeout, max connections.
- **Why:** Connection management is a top-3 backend scaling problem. Know it inside out.
- **Done when:** Sits between your app and Postgres, handles 1000 app connections with 20 DB connections.

---

## Phase 4: API Design & Web Backend (Weeks 21–26)
*Build production-grade backends, not toy apps.*

### P15. RESTful API Framework (Go)
- Router (trie-based, with path params), middleware chain, request validation, error handling, content negotiation, structured logging, graceful shutdown.
- **Why:** Understand what Gin/Echo do under the hood. When the framework fails you, you won't be stuck.
- **Done when:** Build a CRUD API on it that matches the performance of Gin/Echo.

### P16. Authentication & Authorization System (Go + PostgreSQL)
- Full auth system: signup/login, password hashing (Argon2), JWT + refresh tokens, OAuth2 (Google/GitHub), RBAC (role-based access control), rate limiting per user.
- **Why:** Auth is the #1 most common backend task and the #1 source of security vulnerabilities. Master it.
- **Done when:** Secure against: timing attacks, token theft, privilege escalation, brute force. Passes OWASP checklist.

### P17. Production REST API (Go + PostgreSQL + Redis)
- A real app (e.g., project management tool like Linear): users, teams, projects, tasks, comments, file uploads (S3), full-text search, pagination (cursor-based), webhooks, API versioning.
- Use your auth system (P16). Add: database migrations, structured logging, request tracing, Dockerized deployment.
- **Why:** This is *the job*. Every backend role requires this. Do it properly once.
- **Done when:** Deployed with Docker Compose, has API docs (OpenAPI), integration tests, handles 500 req/s.

### P18. GraphQL Engine (Go)
- Schema parser, resolver execution, dataloader (N+1 prevention), subscriptions (WebSocket), query depth/complexity limiting.
- **Why:** Many companies use GraphQL. Understanding it deeply (not just calling a library) is a differentiator.
- **Done when:** Serves your P17 app's data via GraphQL with no N+1 queries.

---

## Phase 5: Async, Queues & Event-Driven (Weeks 27–32)
*Synchronous request-response is the easy part. The hard part is everything else.*

### P19. Job Queue / Task Worker (Go + Redis/Postgres)
- Enqueue jobs, worker pool, retries with exponential backoff, dead letter queue, priority queues, scheduled jobs (cron), job deduplication, admin dashboard.
- **Why:** Every serious backend has async work: emails, image processing, reports, webhooks. This is how.
- **Done when:** Processes 10K jobs/hour, retries failing jobs, no duplicate execution.

### P20. Message Broker (Go)
- Topics, partitions, consumer groups, offset tracking, at-least-once delivery, persistence, replication (single leader).
- Implement the producer/consumer protocol over TCP.
- **Why:** Kafka/RabbitMQ are in every backend stack. Build one to understand pub/sub, ordering, backpressure.
- **Done when:** 100K msg/sec throughput, survives broker restart, consumers rebalance on failure.

### P21. Event Sourcing + CQRS Application (Go)
- Event store, projections (materialized views), command handlers, eventual consistency, snapshotting.
- Build a domain (e.g., banking ledger) where the event log is the source of truth.
- **Why:** This architecture pattern is used at scale (banking, e-commerce, audit systems). It changes how you think about state.
- **Done when:** Full audit trail, projections rebuild from events, handles concurrent commands correctly.

### P22. Webhook Delivery System (Go)
- Reliable outbound webhooks: delivery attempts, exponential retry, circuit breaker per endpoint, signature verification (HMAC), delivery logs, event replay.
- **Why:** Every B2B SaaS product has webhooks. Building reliable delivery is harder than it looks.
- **Done when:** 99.9% delivery rate with proper retries, dead endpoints don't block others.

---

## Phase 6: Infrastructure & DevOps (Weeks 33–38)
*A god-tier backend dev owns the full stack from code to production.*

### P23. Reverse Proxy & Load Balancer (Go)
- HTTP reverse proxy: path-based routing, header manipulation, load balancing (round-robin, least-connections, consistent hashing), health checks, connection draining, rate limiting (token bucket), TLS termination.
- **Why:** You must understand what sits between the internet and your code (Nginx, HAProxy, Envoy).
- **Done when:** Routes traffic to 3+ backends, removes unhealthy instances, rate-limits abusive clients.

### P24. Container Runtime (Go)
- Minimal Docker: Linux namespaces (PID, NET, MNT, UTS), cgroups (CPU/memory limits), pivot_root, overlay filesystem, pull and unpack OCI images.
- **Why:** Containers are how all backend code ships. Understand the Linux primitives beneath Docker/K8s.
- **Done when:** Can pull `alpine` and run a command inside an isolated container.

### P25. Service Mesh Sidecar (Go)
- Transparent TCP proxy: service discovery, mTLS between services, retry policies, circuit breaker, distributed tracing header propagation, metrics emission.
- **Why:** Istio/Linkerd are standard in microservice architectures. Understand the sidecar pattern.
- **Done when:** Two services communicate through your sidecar with mTLS, retries, and tracing.

### P26. CI/CD Pipeline Engine (Go)
- YAML pipeline definitions, Git webhook triggers, isolated job execution (in containers from P24), parallel stages, artifact caching, secrets management, Slack notifications.
- **Why:** You should own your deployment pipeline, not just use it.
- **Done when:** Automatically tests + deploys your P17 API on every git push.

---

## Phase 7: Distributed Systems (Weeks 39–46)
*This is where mortals stop and gods begin.*

### P27. Raft Consensus Implementation (Go)
- Leader election, log replication, snapshotting, membership changes. Full implementation, not just the paper.
- **Why:** Consensus is the beating heart of distributed systems. Raft is used in etcd, CockroachDB, TiKV.
- **Done when:** 3-node cluster. Kill the leader → new leader elected → no data lost. Passes your own Jepsen-lite tests.

### P28. Distributed KV Store (Go)
- Build on top of your Raft implementation. Sharding (range or hash-based), shard rebalancing, distributed transactions (2PC), linearizable reads.
- **Why:** This is what DynamoDB, CockroachDB, TiKV are. Build it, and distributed databases become intuitive.
- **Done when:** Multi-node, survives node failures, shards rebalance, transactions work across shards.

### P29. Distributed Rate Limiter (Go + Redis)
- Sliding window rate limiting across multiple API server instances. Handle clock skew, race conditions, burst allowance.
- **Why:** Rate limiting in distributed systems is a real production problem. It's harder than it sounds.
- **Done when:** Rate limits are consistent across 3+ API server instances under concurrent load.

### P30. Distributed Cron / Job Scheduler (Go)
- Cluster-aware job scheduling: leader election (using your Raft), job assignment, at-most-once execution guarantees, failure detection, job handover.
- **Why:** Running scheduled jobs reliably in a distributed system is a classic hard problem.
- **Done when:** Jobs run exactly once even when nodes crash mid-execution.

---

## Phase 8: Observability & Reliability (Weeks 47–50)
*You can't fix what you can't see.*

### P31. Observability Stack (Go)
Build three components:
1. **Metrics collector** — Prometheus-compatible scraping, time-series storage (gorilla compression), PromQL-subset query engine.
2. **Log aggregator** — Structured log ingestion, indexing, full-text search, tail -f streaming.
3. **Distributed tracer** — Span collection, trace assembly, latency waterfall visualization, service dependency graph.

- Web dashboard that ties all three together.
- **Why:** Observability is what separates "it works on my machine" from "I can debug anything in production."
- **Done when:** Monitor all your previous projects. Find a performance bottleneck using your own tools.

### P32. Chaos Engineering Framework (Go)
- Inject failures into your distributed systems: kill processes, add network latency, drop packets, fill disks, exhaust connections.
- Automated experiments: define steady state → inject fault → verify steady state holds.
- **Why:** God-tier backend devs build systems that *survive* failures. This tool proves they do.
- **Done when:** Run chaos experiments against your distributed KV store (P28) and prove it maintains consistency.

---

## The Stack You'll Master

```
┌─────────────────────────────────────────────────┐
│              GOD-TIER BACKEND DEV                │
├─────────────────────────────────────────────────┤
│  Observability    │ Metrics, Logs, Traces       │
│  Chaos Eng        │ Fault injection, resilience │
├─────────────────────────────────────────────────┤
│  Distributed Sys  │ Raft, sharding, 2PC, CRDTs  │
│  Message Queues   │ Pub/sub, ordering, delivery  │
├─────────────────────────────────────────────────┤
│  Infrastructure   │ Containers, proxies, CI/CD   │
│  Service Mesh     │ mTLS, discovery, circuit brk │
├─────────────────────────────────────────────────┤
│  API Layer        │ REST, GraphQL, RPC, Auth     │
│  Async/Events     │ Queues, CQRS, event sourcing │
├─────────────────────────────────────────────────┤
│  Databases        │ SQL, KV, LSM, B+tree, WAL   │
│  Caching          │ Redis, LRU, bloom filters    │
├─────────────────────────────────────────────────┤
│  Networking       │ TCP, HTTP/1.1/2, WebSocket   │
│  OS / Linux       │ Processes, syscalls, epoll   │
├─────────────────────────────────────────────────┤
│  Languages        │ C++ (systems), Go (backend)  │
│  Fundamentals     │ Data structures, algorithms  │
└─────────────────────────────────────────────────┘
```

---

## Books (Backend-Specific, Read in This Order)

1. **Computer Systems: A Programmer's Perspective** — Your foundation. Understand the machine.
2. **The Linux Programming Interface** — Syscalls, processes, networking, IPC.
3. **Designing Data-Intensive Applications** (Kleppmann) — **THE backend bible.** Read it twice.
4. **Database Internals** (Petrov) — Storage engines, distributed DB internals.
5. **Understanding Distributed Systems** (Vitillo) — Practical distributed systems.
6. **System Design Interview Vol 1 & 2** (Alex Xu) — Scaling patterns, real-world architectures.
7. **Release It!** (Nygard) — Stability patterns, production readiness.
8. **Site Reliability Engineering** (Google) — Free online. How Google runs backends.
9. **Web Scalability for Startup Engineers** (Ejsmont) — Practical scaling guide.

---

## Daily Habits of a God-Tier Backend Dev

- **Read source code** — Study one file from Redis, PostgreSQL, or Go stdlib each week.
- **Profile everything** — `pprof`, `flamegraph`, `strace`, `tcpdump` should be second nature.
- **Write postmortems** — When your projects break, document what happened and why.
- **Benchmark before optimizing** — Numbers, not feelings. `wrk`, `hey`, `pprof`.
- **Think in failure modes** — For every feature: "What happens when this fails?"

---

*Started: ___________*
*Current Phase: ___________*
*Projects Completed: __ / 32*
