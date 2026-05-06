# Resources: Backend Engineering — Zero to God Mode

---

## Phase 1: OS & Systems Fundamentals

### Books
- **Operating Systems: Three Easy Pieces (OSTEP)** — Free: https://pages.cs.wisc.edu/~remzi/OSTEP/
  - Read order: Ch 4–6 (Processes) → Ch 26–32 (Concurrency) → Ch 13–23 (Virtual Memory) → Ch 7–10 (Scheduling) → Ch 36–42 (Filesystems)
- **Computer Systems: A Programmer's Perspective (CS:APP)** — Ch 8 (Exceptional Control Flow), Ch 9 (Virtual Memory)
- **The Linux Programming Interface** — Processes, signals, mmap, epoll, sockets

### Tutorials
- **os-tutorial** — Step-by-step kernel in ~25 lessons: https://github.com/cfenollosa/os-tutorial
- **Writing a Simple OS from Scratch** — 80-page PDF: https://www.cs.bham.ac.uk/~exr/lectures/opsys/10_11/lectures/os-dev.pdf
- **OSDev Wiki** — Reference for GDT, IDT, paging, drivers: https://wiki.osdev.org/

### Source Code to Study
- **xv6** — MIT's teaching OS (~8000 lines, cleanest OS codebase): https://github.com/mit-pdos/xv6-public

### Courses
- **MIT 6.S081** — Full OS course with labs (free): https://pdos.csail.mit.edu/6.828/2021/schedule.html

### Videos / Blogs
- **Philipp Oppermann** — OS concepts with great diagrams: https://os.phil-opp.com/
- **Ben Eater** — Build a 6502 computer (hardware fundamentals): https://www.youtube.com/c/BenEater

---

## Phase 2: Networking & Protocols

### Books
- **Computer Networking: A Top-Down Approach** (Kurose & Ross) — TCP/IP, HTTP, DNS, sockets
- **TCP/IP Illustrated, Vol 1** (Stevens) — The deep-dive reference for TCP internals
- **High Performance Browser Networking** — Free: https://hpbn.co/ — HTTP/2, WebSocket, TLS, latency optimization

### RFCs (Read the Relevant Sections)
- **RFC 2616 / 7230–7235** — HTTP/1.1 specification
- **RFC 7540** — HTTP/2 specification
- **RFC 6455** — WebSocket protocol
- **RFC 793** — TCP specification

### Source Code to Study
- **Redis** — Event loop, RESP protocol, persistence: https://github.com/redis/redis
- **Nginx** — Event-driven architecture, connection handling: https://github.com/nginx/nginx

---

## Phase 3: Databases & Storage Engines

### Books
- **Designing Data-Intensive Applications** (Kleppmann) — **THE backend bible.** Storage engines, replication, partitioning, transactions. Read twice.
- **Database Internals** (Petrov) — B-trees, LSM trees, buffer pools, distributed DB internals
- **Architecture of a Database System** — Free paper: https://dsf.berkeley.edu/papers/fntdb07-architecture.pdf

### Courses
- **CMU 15-445: Database Systems** — Lectures + BusTub project (free): https://15445.courses.cs.cmu.edu/
- **CMU 15-721: Advanced Database Systems** — Query optimization, concurrency control: https://15721.courses.cs.cmu.edu/

### Source Code to Study
- **SQLite** — Single-file DB, incredibly readable: https://sqlite.org/src/doc/trunk/README.md
- **BoltDB** — B+ tree KV store in Go (~4000 lines): https://github.com/boltdb/bolt
- **LevelDB** — LSM tree reference implementation: https://github.com/google/leveldb
- **CockroachDB** — Distributed SQL (Go, uses Raft): https://github.com/cockroachdb/cockroach

### Papers
- **The Log-Structured Merge-Tree (LSM-Tree)** — O'Neil et al.
- **ARIES: A Transaction Recovery Method** — Mohan et al.

---

## Phase 4: API Design & Web Backend

### Books
- **RESTful Web APIs** (Richardson & Amundsen) — API design principles
- **API Design Patterns** (Geewax) — Google's API design approach
- **Bulletproof SSL and TLS** (Ristic) — TLS, certificates, HTTPS in depth

### References
- **OWASP Top 10** — Security vulnerabilities checklist: https://owasp.org/www-project-top-ten/
- **OAuth 2.0 RFC 6749** — The spec behind every "Login with Google" button
- **Google API Design Guide** — Free: https://cloud.google.com/apis/design

### Source Code to Study
- **Gin** — Go HTTP framework (trie router, middleware): https://github.com/gin-gonic/gin
- **Echo** — Another Go HTTP framework: https://github.com/labstack/echo
- **graphql-go** — GraphQL implementation in Go: https://github.com/graphql-go/graphql

---

## Phase 5: Async, Queues & Event-Driven

### Books
- **Enterprise Integration Patterns** (Hohpe & Woolf) — Messaging patterns (pub/sub, routing, dead letter)
- **Designing Data-Intensive Applications** — Ch 11 (Stream Processing), Ch 12 (Future of Data Systems)

### Papers
- **Kafka: A Distributed Messaging System for Log Processing** — LinkedIn's original Kafka paper
- **Event Sourcing** — Martin Fowler: https://martinfowler.com/eaaDev/EventSourcing.html
- **CQRS** — Martin Fowler: https://martinfowler.com/bliki/CQRS.html

### Source Code to Study
- **NATS** — High-performance message broker in Go: https://github.com/nats-io/nats-server
- **Asynq** — Task queue in Go (Redis-based): https://github.com/hibiken/asynq
- **Watermill** — Event-driven Go library: https://github.com/ThreeDotsLabs/watermill

---

## Phase 6: Infrastructure & DevOps

### Books
- **Container Security** (Rice) — Linux namespaces, cgroups, image security
- **Kubernetes in Action** (Luksa) — After you build your own container runtime

### Tutorials
- **Containers from Scratch** — Liz Rice's talk: https://www.youtube.com/watch?v=8fi7uSYlOdc
- **Build Your Own Docker** — CodeCrafters challenge: https://app.codecrafters.io/courses/docker/overview

### Source Code to Study
- **runc** — OCI container runtime (Go): https://github.com/opencontainers/runc
- **Traefik** — Reverse proxy in Go: https://github.com/traefik/traefik
- **Caddy** — HTTP server with automatic HTTPS (Go): https://github.com/caddyserver/caddy
- **Drone CI** — CI/CD engine in Go: https://github.com/harness/drone

---

## Phase 7: Distributed Systems

### Books
- **Designing Data-Intensive Applications** — Ch 5 (Replication), Ch 6 (Partitioning), Ch 7 (Transactions), Ch 8 (Distributed Faults), Ch 9 (Consistency & Consensus)
- **Understanding Distributed Systems** (Vitillo) — Practical, concise, modern
- **Distributed Systems** (van Steen & Tanenbaum) — Free: https://www.distributed-systems.net/

### Papers (Must-Read)
- **In Search of an Understandable Consensus Algorithm (Raft)** — Ongaro & Ousterhout: https://raft.github.io/raft.pdf
- **Dynamo: Amazon's Highly Available Key-Value Store** — DeCandia et al.
- **MapReduce: Simplified Data Processing on Large Clusters** — Dean & Ghemawat
- **The Google File System** — Ghemawat et al.
- **Spanner: Google's Globally-Distributed Database** — Corbett et al.
- **Time, Clocks, and the Ordering of Events in a Distributed System** — Lamport

### Courses
- **MIT 6.824: Distributed Systems** — Lectures + labs (Raft, MapReduce, KV store): https://pdos.csail.mit.edu/6.824/schedule.html

### Source Code to Study
- **etcd** — Raft-based KV store (Go): https://github.com/etcd-io/etcd
- **Hashicorp Raft** — Raft library in Go: https://github.com/hashicorp/raft
- **TiKV** — Distributed KV (Rust, but the design docs are gold): https://github.com/tikv/tikv

### Visualization
- **Raft Visualization** — Interactive: https://thesecretlivesofdata.com/raft/

---

## Phase 8: Observability & Reliability

### Books
- **Site Reliability Engineering** (Google) — Free: https://sre.google/sre-book/table-of-contents/
- **Observability Engineering** (Majors, Fong-Jones, Miranda) — Modern observability practices
- **Release It!** (Nygard) — Stability patterns: circuit breakers, bulkheads, timeouts
- **Chaos Engineering** (Rosenthal et al.) — Principles of chaos engineering

### Source Code to Study
- **Prometheus** — Metrics + time-series DB (Go): https://github.com/prometheus/prometheus
- **Jaeger** — Distributed tracing (Go): https://github.com/jaegertracing/jaeger
- **Loki** — Log aggregation (Go): https://github.com/grafana/loki

---

## System Design

### Books
- **System Design Interview Vol 1 & 2** (Alex Xu) — Practical scaling patterns
- **Web Scalability for Startup Engineers** (Ejsmont) — Caching, queues, sharding
- **A Philosophy of Software Design** (Ousterhout) — Writing clean, deep modules

### Blogs
- **The Morning Paper** — CS paper summaries: https://blog.acolyer.org/
- **High Scalability** — Architecture case studies: http://highscalability.com/
- **Netflix Tech Blog** — https://netflixtechblog.com/
- **Uber Engineering Blog** — https://www.uber.com/blog/engineering/
- **Cloudflare Blog** — https://blog.cloudflare.com/

---

## Go-Specific Resources

### Books
- **The Go Programming Language** (Donovan & Kernighan) — The Go bible
- **Concurrency in Go** (Cox-Buday) — Goroutines, channels, patterns
- **Let's Go / Let's Go Further** (Alex Edwards) — Production web apps in Go

### References
- **Go stdlib source code** — Read `net/http`, `sync`, `context`, `database/sql`: https://github.com/golang/go/tree/master/src
- **Effective Go** — https://go.dev/doc/effective_go
- **Go blog** — https://go.dev/blog/

---

## C++ Systems Programming Resources

### Books
- **The C++ Programming Language** (Stroustrup) — Reference
- **Effective Modern C++** (Meyers) — Write clean, modern C++

### References
- **cppreference.com** — https://en.cppreference.com/
- **Beej's Guide to Network Programming** — Sockets in C/C++: https://beej.us/guide/bgnet/

---

## General Engineering

### Papers
- **How to Read a Paper** (Keshav) — 3-pass approach to reading CS papers: https://svr-sk818-web.cl.cam.ac.uk/keshav/wiki/index.php/HTRAP

### Practice
- **CodeCrafters** — Build Redis, Docker, Git, DNS from scratch: https://codecrafters.io/
- **Challenging Projects Every Programmer Should Try** — https://austinhenley.com/blog/challengingprojects.html

---

*Bookmark this file. Come back to it as you start each phase.*
