# The God-Tier CS & Software Engineering Roadmap
### Learn Everything by Building — 50 Projects from Zero to Mastery

---

## How to Use This Roadmap
- **Go in order within each phase.** Each project builds on skills from the previous one.
- **Don't just code — understand.** After each project, write a short README explaining what you learned.
- **Push everything to GitHub.** Build your portfolio as you go.
- **Time estimate:** ~2–3 years if done seriously alongside other work. Faster if full-time.

---

## Phase 1: Foundations (Months 1–3)
*Goal: Master a language deeply, understand how computers actually work.*

### P1. Custom Shell (C)
- Build a Unix shell from scratch: parsing, forking, piping, redirection, signal handling.
- **You'll learn:** Systems programming, process management, file descriptors, POSIX API.
- **Milestone:** Can run `ls | grep foo > out.txt` in your shell.

### P2. Memory Allocator (C)
- Implement `malloc()`, `free()`, `realloc()` using `sbrk`/`mmap`.
- **You'll learn:** Virtual memory, fragmentation, free lists, alignment, how the heap works.
- **Milestone:** Your allocator passes a stress test with thousands of alloc/free cycles.

### P3. Data Structures Library (C or Rust)
- Implement from scratch: dynamic array, linked list, hash map, BST, AVL tree, trie, heap, graph (adjacency list).
- **You'll learn:** Pointer manipulation, Big-O analysis, memory layout, iterator patterns.
- **Milestone:** Each structure has unit tests and benchmarks.

### P4. Sorting & Algorithm Visualizer (Python + Web)
- Implement 10+ sorting/search algorithms. Build a web UI that animates them step-by-step.
- **You'll learn:** Algorithm analysis, recursion, divide-and-conquer, basic web dev.
- **Milestone:** Interactive visualizer deployed on the web.

---

## Phase 2: Systems Programming (Months 3–6)
*Goal: Understand what happens beneath your code.*

### P5. HTTP Server from Scratch (C or Rust)
- Handle GET/POST, serve static files, parse headers, support keep-alive, chunked transfer.
- **You'll learn:** TCP/IP, sockets, HTTP protocol, concurrency (threads or async I/O).
- **Milestone:** Serves a static website and handles 1000 concurrent connections.

### P6. TCP/IP Stack (C)
- Implement a minimal TCP/IP stack over raw sockets or TUN/TAP: Ethernet frames, IP, TCP handshake, retransmission.
- **You'll learn:** Networking fundamentals at the deepest level.
- **Milestone:** Can do a 3-way handshake and transfer data reliably.

### P7. Mini Operating System Kernel (C + Assembly)
- Boot from scratch, enter protected mode, handle interrupts, basic memory management, simple filesystem, keyboard driver.
- **You'll learn:** How an OS actually works — boot sequence, page tables, context switching.
- **Resource:** Follow along with [OSDev Wiki](https://wiki.osdev.org/) or [Writing an OS in Rust](https://os.phil-opp.com/).
- **Milestone:** Boots in QEMU, shows a shell prompt, runs simple commands.

### P8. Debugger (C or Rust)
- Build a debugger for Linux using `ptrace`: set breakpoints, step through code, read registers, inspect memory.
- **You'll learn:** ELF format, DWARF debug info, system calls, low-level process control.
- **Milestone:** Can attach to a process, set a breakpoint on a function, and print variables.

---

## Phase 3: Databases & Storage (Months 6–8)
*Goal: Understand how data is stored, indexed, and queried.*

### P9. Key-Value Store (Rust or Go)
- Build a persistent KV store with a log-structured merge tree (LSM) or B-tree.
- **You'll learn:** Write-ahead logging, compaction, SSTables, memory-mapped I/O.
- **Milestone:** Survives crashes (durability) and passes ACID property tests.

### P10. SQL Database Engine (C++ or Rust)
- Tokenizer → Parser → Query Planner → Executor → Storage engine.
- Support: `CREATE TABLE`, `INSERT`, `SELECT` with `WHERE`, `JOIN`, `ORDER BY`, `INDEX`.
- **You'll learn:** Relational algebra, B+ trees, query optimization, buffer pools.
- **Resource:** Follow [CMU 15-445](https://15445.courses.cs.cmu.edu/) (BusTub project).
- **Milestone:** Run TPC-H-style queries on your database.

### P11. Distributed Key-Value Store (Go or Rust)
- Implement Raft consensus. Build a distributed, replicated, partition-tolerant KV store.
- **You'll learn:** Consensus algorithms, leader election, log replication, network partitions.
- **Milestone:** Survives node failures and maintains consistency.

---

## Phase 4: Compilers & Languages (Months 8–11)
*Goal: Understand how languages work from the inside.*

### P12. Interpreter for a Scripting Language (Java, Go, or Rust)
- Build a tree-walk interpreter: lexer, parser (Pratt parsing), AST, evaluator.
- Support: variables, functions, closures, control flow, basic types.
- **Resource:** [Crafting Interpreters](https://craftinginterpreters.com/) by Bob Nystrom.
- **Milestone:** Your language can run recursive Fibonacci and closures.

### P13. Bytecode Virtual Machine (C or Rust)
- Compile your language to bytecode, build a stack-based VM, add garbage collection (mark-and-sweep).
- **You'll learn:** Compilation, instruction sets, GC algorithms, performance optimization.
- **Milestone:** 10x faster than your tree-walk interpreter.

### P14. Compiler to Native Code (Rust or C++)
- Build a compiler for a C-like language targeting x86-64 or ARM.
- Lexer → Parser → IR (SSA) → Optimization passes → Code generation → ELF output.
- **You'll learn:** SSA form, register allocation, instruction selection, linking.
- **Milestone:** Compile and run a program that computes primes — no runtime dependency.

### P15. Regular Expression Engine (Any language)
- Thompson's construction → NFA → DFA → Optimizer.
- **You'll learn:** Automata theory, state machines, backtracking vs. linear-time matching.
- **Milestone:** Passes a regex test suite and handles pathological patterns without blowup.

---

## Phase 5: Web & Full-Stack Mastery (Months 11–14)
*Goal: Build production-quality web applications end-to-end.*

### P16. Full-Stack Social Platform (TypeScript, React, Node.js, PostgreSQL)
- Auth (OAuth + JWT), real-time feed, posts, comments, likes, notifications, image uploads (S3), search.
- **You'll learn:** REST API design, auth flows, ORMs, migrations, caching, full-stack architecture.
- **Milestone:** Deployed, handles 100+ concurrent users.

### P17. Real-Time Collaborative Editor (TypeScript)
- Like Google Docs: multiple users editing the same document simultaneously.
- Use CRDTs or Operational Transformation for conflict resolution. WebSockets for sync.
- **You'll learn:** Distributed systems concepts, CRDTs, WebSockets, state synchronization.
- **Milestone:** 3+ users can edit the same doc with no conflicts.

### P18. Build Your Own React (TypeScript)
- Virtual DOM, diffing algorithm, reconciliation, hooks (useState, useEffect), JSX transform.
- **You'll learn:** How UI frameworks actually work, DOM manipulation, fiber architecture.
- **Milestone:** Can render a simple interactive app with state and effects.

### P19. API Gateway & Rate Limiter (Go or Rust)
- Reverse proxy with: routing, rate limiting (token bucket), circuit breaker, request logging, auth middleware.
- **You'll learn:** Middleware patterns, distributed rate limiting, proxy architecture.
- **Milestone:** Sits in front of your social platform and handles traffic shaping.

---

## Phase 6: DevOps, Infrastructure & Cloud (Months 14–16)
*Goal: Understand the full deployment pipeline.*

### P20. Container Runtime (Go)
- Build a minimal Docker: namespaces, cgroups, chroot/pivot_root, overlay filesystem, image pulling.
- **You'll learn:** Linux internals, process isolation, resource limits.
- **Milestone:** Can pull and run a simple container image.

### P21. CI/CD Pipeline (Any)
- Build a CI server: watch Git repos, run builds on commits, parallel test execution, artifact storage, notifications.
- **You'll learn:** Build systems, test automation, webhooks, job scheduling.
- **Milestone:** Automatically tests and deploys your social platform on each push.

### P22. Infrastructure-as-Code Tool (Go or Python)
- Declarative config → Diff current state vs. desired → Apply changes to cloud resources (use AWS/GCP APIs).
- **You'll learn:** State management, dependency graphs, cloud APIs, idempotent operations.
- **Milestone:** Can provision a VPC + EC2 + RDS setup from a config file.

### P23. Load Balancer (Go or Rust)
- Layer 4 (TCP) and Layer 7 (HTTP) load balancing. Health checks, round-robin, least-connections, sticky sessions.
- **You'll learn:** Network programming, connection pooling, health monitoring.
- **Milestone:** Distributes traffic across 3+ backend instances with failover.

---

## Phase 7: Security & Cryptography (Months 16–17)
*Goal: Understand how to attack and defend systems.*

### P24. Cryptography Library (Rust or C)
- Implement: AES, SHA-256, RSA, Diffie-Hellman, HMAC, PBKDF2.
- **Warning:** For learning only — never use homegrown crypto in production.
- **You'll learn:** Number theory, block ciphers, asymmetric crypto, side-channel awareness.
- **Milestone:** Passes NIST test vectors.

### P25. TLS 1.3 Implementation (Rust)
- ClientHello → ServerHello → Key exchange → Handshake encryption → Application data.
- **You'll learn:** The full TLS handshake, certificate chains, AEAD ciphers, forward secrecy.
- **Milestone:** Your client can connect to google.com over TLS 1.3.

### P26. Vulnerability Scanner (Python)
- Port scanning, banner grabbing, known CVE matching, basic web vuln detection (SQLi, XSS, SSRF probes).
- **You'll learn:** Network security, common attack vectors, defensive scanning.
- **Milestone:** Scans your own deployed apps and generates a report.

---

## Phase 8: AI/ML Engineering (Months 17–20)
*Goal: Understand AI from first principles, not just API calls.*

### P27. Neural Network from Scratch (Python — NumPy only)
- Build a feedforward neural net: forward pass, backpropagation, SGD, batch norm, dropout.
- Train on MNIST, CIFAR-10.
- **You'll learn:** Calculus of backprop, gradient descent, loss functions, regularization.
- **Milestone:** >97% accuracy on MNIST without any ML library.

### P28. Autograd Engine (Python)
- Build automatic differentiation: computation graph, backward pass, tensor operations.
- Like a mini-PyTorch.
- **Resource:** Andrej Karpathy's [micrograd](https://github.com/karpathy/micrograd).
- **You'll learn:** How PyTorch/TensorFlow actually work inside.
- **Milestone:** Can train a small neural net using your autograd.

### P29. Transformer from Scratch (Python + PyTorch)
- Implement the full "Attention Is All You Need" architecture: multi-head attention, positional encoding, encoder-decoder.
- Train a small language model or machine translation model.
- **You'll learn:** Attention mechanisms, tokenization, training loops, GPU utilization.
- **Milestone:** Your model generates coherent text after training on a small corpus.

### P30. RAG Application (Python + Vector DB)
- Build a Retrieval-Augmented Generation system: document ingestion, chunking, embedding, vector search, LLM integration.
- **You'll learn:** Embeddings, vector databases, prompt engineering, LLM orchestration.
- **Milestone:** Chat with your own documents with cited sources.

### P31. ML Model Serving Platform (Python + Go)
- Model registry, A/B testing, canary deployments, autoscaling, monitoring, latency tracking.
- **You'll learn:** MLOps, model versioning, inference optimization, production ML.
- **Milestone:** Serves your transformer model with <100ms p99 latency.

---

## Phase 9: Distributed Systems & Scale (Months 20–23)
*Goal: Build systems that work at scale.*

### P32. Message Queue (Go or Rust)
- Persistent, distributed message queue: topics, partitions, consumer groups, at-least-once delivery.
- **You'll learn:** Durability, ordering guarantees, backpressure, zero-copy I/O.
- **Milestone:** Handles 100K messages/sec on a single node.

### P33. Distributed File System (Go or Rust)
- Chunked storage, replication, metadata server, client library.
- **You'll learn:** Data replication, consistency models, failure recovery.
- **Milestone:** Store and retrieve files across 3+ nodes, survive a node failure.

### P34. Search Engine (Rust or C++)
- Web crawler → HTML parser → Inverted index → TF-IDF ranking → Query engine → Web UI.
- **You'll learn:** Information retrieval, indexing, ranking algorithms, crawling etiquette.
- **Milestone:** Crawls 10K+ pages and returns relevant search results in <100ms.

### P35. MapReduce / Stream Processing Framework (Go or Java)
- Distributed computation: job scheduling, task distribution, fault tolerance, shuffle phase.
- **You'll learn:** Distributed computation models, fault tolerance, data partitioning.
- **Milestone:** Can run word-count across a multi-GB dataset on a cluster.

### P36. Time-Series Database (Rust or Go)
- Columnar storage, compression (gorilla encoding), downsampling, retention policies, query language.
- **You'll learn:** Specialized storage engines, compression, write-optimized data structures.
- **Milestone:** Ingests 1M data points/sec and supports aggregation queries.

---

## Phase 10: Graphics, Games & Low-Level Performance (Months 23–25)
*Goal: Push hardware to its limits.*

### P37. Software Rasterizer (C++ or Rust)
- 3D rendering without GPU APIs: projection, clipping, rasterization, z-buffer, texture mapping, Phong shading.
- **You'll learn:** Linear algebra, rendering pipeline, how GPUs actually work conceptually.
- **Milestone:** Renders a textured, lit 3D model in real-time.

### P38. Ray Tracer (C++ or Rust)
- Recursive ray tracer: reflections, refractions, soft shadows, BVH acceleration, multithreaded.
- **Resource:** [Ray Tracing in One Weekend](https://raytracing.github.io/) series.
- **You'll learn:** Physically-based rendering, spatial data structures, SIMD optimization.
- **Milestone:** Renders a complex scene with reflective/transparent objects.

### P39. Game Engine (C++ or Rust)
- ECS architecture, physics (collision detection, rigid body), renderer (OpenGL/Vulkan), audio, scripting, editor UI.
- **You'll learn:** Real-time systems, game loops, spatial partitioning, asset pipelines.
- **Milestone:** A playable 2D or 3D game running on your engine.

### P40. Physics Simulator (C++ or Rust)
- Rigid body dynamics, collision detection (GJK/EPA), constraint solver, cloth/fluid simulation.
- **You'll learn:** Numerical methods, computational geometry, simulation stability.
- **Milestone:** Simulates stacking boxes, ragdolls, or fluid in real-time.

---

## Phase 11: Mobile & Embedded (Months 25–26)
*Goal: Software for constrained environments.*

### P41. Mobile App with Offline Sync (React Native or Flutter)
- Full CRUD app with local SQLite, conflict resolution, background sync, push notifications.
- **You'll learn:** Mobile architecture, offline-first design, platform APIs.
- **Milestone:** Works completely offline and syncs when connection returns.

### P42. Embedded System — IoT Dashboard (C + Rust)
- Microcontroller (ESP32/RPi Pico) collecting sensor data → MQTT → Backend → Real-time dashboard.
- **You'll learn:** Embedded C, RTOS, communication protocols, hardware interfaces.
- **Milestone:** Real sensor data displayed on a live web dashboard.

---

## Phase 12: Software Engineering Excellence (Months 26–28)
*Goal: Write code that survives contact with the real world.*

### P43. Open Source Contribution (Any)
- Pick a large, active open-source project (Linux kernel, Postgres, Rust compiler, Chromium, etc.).
- Fix a real bug or implement a small feature. Go through the full PR review process.
- **You'll learn:** Reading large codebases, coding standards, collaboration, code review.
- **Milestone:** Merged PR in a project with 1000+ stars.

### P44. System Design Case Studies (Documentation)
- Design (on paper + diagrams) 5 real systems: URL shortener, Uber, Twitter, YouTube, Slack.
- For each: requirements, API design, data model, scaling strategy, failure modes.
- **You'll learn:** System design thinking, trade-off analysis, capacity planning.
- **Milestone:** Design docs that could pass a senior-level system design interview.

### P45. Chaos Engineering Tool (Go or Python)
- Inject failures into your distributed systems: kill processes, add latency, partition networks, corrupt data.
- **You'll learn:** Resilience testing, observability, failure analysis.
- **Milestone:** Runs chaos experiments on your distributed KV store and generates reports.

---

## Phase 13: Capstone Projects (Months 28–30)
*Goal: Combine everything into ambitious, portfolio-defining projects.*

### P46. Programming Language + Ecosystem
- Design your own language with a unique feature (ownership system, effect system, dependent types, etc.).
- Compiler + VM + standard library + package manager + LSP server (for editor support).
- **Milestone:** Someone else can write and run a non-trivial program in your language.

### P47. Cloud Platform (Mini-AWS)
- Compute (container orchestration), storage (object store), networking (virtual network), auth (IAM).
- Web console + CLI + SDK.
- **Milestone:** Deploy a multi-service application on your own cloud platform.

### P48. Distributed Database with SQL
- Combine everything: SQL parser + query optimizer + distributed storage + Raft consensus + transactions (2PC/Percolator).
- **Milestone:** Passes Jepsen-style consistency tests under network partitions.

### P49. Real-Time Multiplayer Game with Custom Engine
- Game engine + netcode (client-side prediction, server reconciliation, lag compensation) + matchmaking + leaderboard.
- **Milestone:** 10+ players in a smooth real-time game over the internet.

### P50. Full Monitoring & Observability Stack
- Metrics collection (like Prometheus), log aggregation (like Loki), distributed tracing (like Jaeger), alerting, dashboards.
- Deploy it to monitor all your previous projects.
- **Milestone:** Detect and diagnose a production incident across multiple services using your stack.

---

## Recommended Resources (Books)

| Topic | Book |
|---|---|
| C & Systems | *Computer Systems: A Programmer's Perspective* (CS:APP) |
| Operating Systems | *Operating Systems: Three Easy Pieces* (OSTEP) — free online |
| Networking | *Computer Networking: A Top-Down Approach* (Kurose & Ross) |
| Databases | *Designing Data-Intensive Applications* (Kleppmann) |
| Compilers | *Crafting Interpreters* (Nystrom) — free online |
| Algorithms | *Algorithm Design Manual* (Skiena) |
| Distributed Systems | *Designing Data-Intensive Applications* + MIT 6.824 lectures |
| System Design | *System Design Interview* Vol 1 & 2 (Alex Xu) |
| Security | *Serious Cryptography* (Aumasson) |
| ML | *Deep Learning* (Goodfellow) — free online |
| Software Eng | *A Philosophy of Software Design* (Ousterhout) |

---

## Principles for the Journey

1. **Build first, theory second.** Get stuck, then read the textbook — it'll make 10x more sense.
2. **Read source code.** Study Redis, SQLite, Nginx, Linux — they're masterclasses.
3. **Write tests.** Every project. No exceptions.
4. **Benchmark everything.** Know your numbers — latency, throughput, memory.
5. **Teach what you learn.** Blog posts, videos, talks. Teaching is the ultimate test of understanding.
6. **Contribute to open source.** It's the fastest way to level up by learning from great engineers.
7. **Depth > Breadth.** It's better to deeply understand 5 systems than to superficially know 50.

---

*Started: ___________*
*Current Phase: ___________*
*Projects Completed: __ / 50*
