# Module 7: The System Design Interview Framework

> The exact method to ace any system design interview, from L4 to Staff+.

---

## 7.1 — The 4-Step Framework

Every system design interview should follow this structure. Timing for a 45-minute interview:

| Step | Time | What |
|------|------|------|
| **1. Requirements & Scope** | 5 min | Clarify what to build |
| **2. Estimation & Constraints** | 5 min | Scale numbers, constraints |
| **3. High-Level Design** | 15 min | Architecture, components, data flow |
| **4. Deep Dives** | 20 min | Detailed design of 2-3 key components |

---

## Step 1: Requirements & Scope (5 minutes)

**Never start designing without clarifying requirements.** This is the #1 mistake.

### Functional Requirements
Ask: "What are the core features?"

Example for "Design Twitter":
```
Must-have:
  - Post tweets (text + media)
  - Follow/unfollow users
  - Home timeline (feed)
  - Search tweets

Nice-to-have (ask interviewer):
  - Trending topics?
  - Direct messages?
  - Notifications?
  - Analytics?
```

### Non-Functional Requirements
Ask explicitly:
```
- Scale: How many users? DAU?
- Latency: What's acceptable? (p99 < 200ms?)
- Availability: What SLA? (99.9%? 99.99%?)
- Consistency: Is eventual consistency OK or do we need strong?
- Read/write ratio: Read-heavy? Write-heavy?
- Data retention: How long to keep data?
- Geographic: Single region or multi-region?
```

### Scope Boxing

After gathering requirements, confirm scope:

> "So to summarize, we're building X with features A, B, C. We need to support N users with P latency. I'll focus on the core flow and then deep-dive into [the hardest part]. Does that sound right?"

This prevents:
- Going too broad (designing everything)
- Going too narrow (missing the hard part)
- Solving the wrong problem

---

## Step 2: Estimation & Constraints (5 minutes)

### The Estimation Template

```
Users:
  Total users:    ___
  DAU:            ___
  Peak concurrent: DAU / 5 typically

Traffic:
  Reads/sec:      ___
  Writes/sec:     ___
  Peak:           Average × 3-5x

Storage:
  Per record:     ___ bytes
  Daily:          ___ GB
  5-year total:   ___ TB

Bandwidth:
  Ingress:        ___ MB/s
  Egress:         ___ MB/s
```

### Example: Design Instagram

```
Users:
  2B total, 500M DAU
  Peak concurrent: 100M

Traffic:
  Photo uploads: 500M DAU × 2 photos/day = 1B/day = ~12K writes/sec
  Feed reads: 500M DAU × 10 reads/day = 5B/day = ~58K reads/sec
  Read:write = 5:1

Storage:
  Average photo: 2 MB (multiple resolutions stored)
  Daily storage: 1B × 2 MB = 2 PB/day
  5-year: 2 PB × 365 × 5 = 3.6 EB (need tiered storage!)

Bandwidth:
  Upload: 12K photos/sec × 2 MB = 24 GB/sec ingress
  Downloads: Much higher (CDN absorbs most of this)
```

**Key insight: These numbers guide your architecture.**
- 2 PB/day → object storage (S3), tiered storage essential
- 58K reads/sec → heavy caching, CDN mandatory
- 12K writes/sec → manageable with sharded writes

---

## Step 3: High-Level Design (15 minutes)

### Draw the Architecture

Start with the user and work inward:

```
1. [Client] → 
2. [DNS / CDN] → 
3. [Load Balancer] → 
4. [API Gateway] → 
5. [Services] → 
6. [Data Stores]
```

**For each component, explain WHY it's there.**

### Template for Common Patterns

```
Read-Heavy System:
  Client → CDN → LB → API Servers → Cache (Redis) → DB (read replicas)

Write-Heavy System:
  Client → LB → API Servers → Message Queue → Workers → DB

Real-Time System:
  Client ↔ WebSocket Servers → Pub/Sub (Redis/Kafka) → Presence/State

Search System:
  Write Path: API → DB → CDC → Search Index (Elasticsearch)
  Read Path: API → Search Index → DB (for full records if needed)
```

### Data Model

Sketch the core entities and relationships:

```
Twitter:
  User:    { id, name, email, created_at }
  Tweet:   { id, user_id, text, media_urls, created_at }
  Follow:  { follower_id, followee_id, created_at }
  Feed:    { user_id, tweet_id, score }
```

### API Design

Define the key endpoints:

```
POST /tweets           { text, media }         → { tweet_id }
GET  /feed             { cursor, page_size }    → { tweets[], next_cursor }
POST /follow/{user_id}                         → { success }
GET  /search           { query, filters }       → { tweets[] }
```

**Use cursor-based pagination**, not offset-based. Offset is broken at scale (expensive for deep pages, inconsistent with real-time data).

---

## Step 4: Deep Dives (20 minutes)

This is where you show **depth**. Pick the 2-3 most interesting/challenging components and go deep.

### How to Pick What to Deep-Dive

1. **What's the hardest part?** (The interviewer usually has this in mind)
2. **What requires a non-obvious solution?** (Not just "add more servers")
3. **What differentiates this system?** (The unique technical challenge)

### Deep Dive Checklist

For each component:
```
□ How does it work internally?
□ How does it scale?
□ How does it handle failure?
□ What are the trade-offs?
□ What alternatives did you consider?
```

### Example Deep Dives by System

| System | Good Deep Dives |
|--------|----------------|
| Twitter | Fan-out strategy (push vs pull), timeline ranking |
| Uber | Geospatial indexing (Geohash vs Quadtree), matching algorithm |
| YouTube | Video transcoding pipeline, adaptive bitrate streaming |
| Chat | WebSocket connection management, message ordering |
| Search | Inverted index design, ranking algorithm |
| Payment | Exactly-once processing, reconciliation |

---

## 7.2 — Communication Tips

### Signal What You're Doing

Narrate your thought process:
- "Let me start with the requirements..."
- "The bottleneck here is X, so I'll focus on that..."
- "There are two options: A and B. A is better because..."
- "If we need to scale further, we could..."

### Make Trade-Offs Explicit

**Bad:** "I'll use Cassandra for the database."
**Good:** "For this use case, I'm choosing Cassandra over PostgreSQL because we need high write throughput and can tolerate eventual consistency. The trade-off is that we lose complex query capability, but our access patterns are simple key-value lookups."

### Handle Unknown Territory

If you don't know something:
- "I'm not 100% sure of the exact algorithm, but I know it works by..."
- "I haven't used X directly, but based on similar systems, I'd approach it by..."
- Never bluff. Interviewers can tell.

### Drive the Discussion

Don't wait for the interviewer to tell you what to do next. Lead:
- "I think the most interesting challenge here is X. Let me dive into that."
- "We've covered the high-level design. Shall I deep-dive into the data model or the caching strategy?"

---

## 7.3 — Common Pitfalls

| Pitfall | Fix |
|---------|-----|
| Jumping straight into design | Always clarify requirements first |
| Over-engineering from day one | Start simple, then scale |
| Not considering failure modes | Ask "what if X fails?" for each component |
| Ignoring consistency requirements | Always state your consistency model |
| Using buzzwords without understanding | Only mention tech you can explain deeply |
| Designing everything, deep on nothing | Focus. 2-3 deep dives >> surface-level everything |
| Not doing math | Estimations drive architecture decisions |
| Ignoring database choice justification | Every DB choice needs a "because..." |
| Single point of failure | Every critical component needs redundancy |
| Not discussing monitoring/alerting | Show you think about operations |

---

## 7.4 — The Cheat Sheet

Copy this card and review before every interview.

```
┌─────────────────────────────────────────────────────────┐
│                SYSTEM DESIGN CHEAT SHEET                 │
├─────────────────────────────────────────────────────────┤
│                                                          │
│  1. CLARIFY (5 min)                                     │
│     □ Functional requirements (features)                │
│     □ Non-functional (scale, latency, consistency)      │
│     □ Scope and confirm                                 │
│                                                          │
│  2. ESTIMATE (5 min)                                    │
│     □ DAU → QPS (reads/writes)                          │
│     □ Storage (per record × volume × retention)         │
│     □ Bandwidth                                         │
│                                                          │
│  3. DESIGN (15 min)                                     │
│     □ API endpoints                                     │
│     □ Data model                                        │
│     □ High-level architecture                           │
│     □ Data flow for core use case                       │
│                                                          │
│  4. DEEP DIVE (20 min)                                  │
│     □ Pick 2-3 hard components                          │
│     □ Internals, scaling, failure modes                 │
│     □ Trade-offs and alternatives                       │
│                                                          │
│  ALWAYS MENTION:                                        │
│     □ Load balancing strategy                           │
│     □ Caching layer (what, where, invalidation)         │
│     □ Database choice + justification                   │
│     □ Consistency model                                 │
│     □ Failure handling                                  │
│     □ Monitoring / alerting                             │
│                                                          │
│  SCALING TOOLKIT:                                       │
│     Horizontal scaling · Sharding · Caching             │
│     CDN · Message queues · Read replicas                │
│     Rate limiting · Circuit breakers                    │
│     Async processing · Denormalization                  │
│                                                          │
│  NUMBERS:                                               │
│     1 day = 86K sec · 2^10 = 1K · 2^20 = 1M            │
│     2^30 = 1B · 2^40 = 1T                              │
│     Redis: 100K ops/s · Kafka: 1M msgs/s               │
│     pg: 10K-50K qps · HTTP server: 1-10K rps           │
│                                                          │
└─────────────────────────────────────────────────────────┘
```

---

## 7.5 — Practice Problems (Sorted by Difficulty)

### Easy
1. URL Shortener
2. Paste Bin
3. Rate Limiter
4. Key-Value Store (single node)

### Medium
5. Design Instagram
6. Design a Chat System (WhatsApp)
7. Design a Notification System
8. Design a Web Crawler
9. Design Typeahead/Autocomplete
10. Design a News Feed
11. Design a Unique ID Generator
12. Design a File Sharing System (Google Drive)

### Hard
13. Design YouTube
14. Design Uber
15. Design Twitter
16. Design a Search Engine
17. Design a Payment System
18. Design Google Maps
19. Design a Distributed Message Queue
20. Design a Distributed Cache

### Expert
21. Design Google Docs (Collaborative Editing — CRDTs/OT)
22. Design a Stock Exchange (Ultra-low latency matching engine)
23. Design a Distributed Consensus System
24. Design a Real-Time Gaming Backend
25. Design a Global DNS System

For each problem, use the 4-step framework. Time yourself (45 min). Practice with a friend who asks follow-up questions.

---

## 7.6 — What Interviewers Actually Look For

| Level | What they evaluate |
|-------|-------------------|
| **Junior (L3-L4)** | Can you design a working system? Basic components, data flow |
| **Mid (L5)** | Do you identify and solve the hard problems? Trade-off awareness |
| **Senior (L6)** | Can you drive the design independently? Deep expertise in 2-3 areas |
| **Staff (L7+)** | Do you see the big picture? Cross-system implications, organizational impact |

### The Evaluation Matrix

```
                        Weak          Okay           Strong
Requirements      Skipped/vague   Asked basics    Probed deeply, scoped well
Estimation        No numbers      Rough numbers   Precise, guided decisions
Design            Missing pieces  Functional      Elegant, justified, scalable
Deep Dive         Surface only    Good depth       Expert-level, alternatives
Trade-offs        "It depends"    Named a few     Every decision had a trade-off
Communication     Quiet/rambling  Clear           Led the discussion confidently
```

---

## 7.7 — Your Study Plan

### Week 1-2: Foundations
- Read Modules 1-2 (Fundamentals + Scaling Patterns)
- Practice 3 back-of-envelope estimations daily
- Design: URL Shortener, Rate Limiter

### Week 3-4: Data & Distributed Systems
- Read Modules 3-4 (Data Systems + Distributed Systems)
- Deep study: CAP, consistency models, sharding strategies
- Design: Chat System, News Feed, Unique ID Generator

### Week 5-6: Infrastructure
- Read Module 5 (Infrastructure & Reliability)
- Study Kafka, Kubernetes, observability
- Design: Notification System, Web Crawler, Typeahead

### Week 7-8: Hard Problems
- Read Module 6 (Real-World Designs)
- Design: YouTube, Uber, Twitter
- Focus on deep dives

### Week 9-10: Mock Interviews
- Practice with real humans (friends, paid mocks)
- Time yourself strictly (45 min)
- Record yourself and review

### Ongoing
- Read engineering blogs: Uber, Netflix, Stripe, Cloudflare, Meta
- Study open-source systems: Kafka, Redis, Cassandra source code
- Think about systems you use daily: How would you build Instagram? Spotify? Slack?

---

## 7.8 — Recommended Resources

### Books
- *Designing Data-Intensive Applications* by Martin Kleppmann — **The Bible. Read this.**
- *System Design Interview* by Alex Xu (Vol 1 & 2) — Good for interview-specific patterns
- *Building Microservices* by Sam Newman — Microservices done right

### Engineering Blogs
- [Uber Engineering](https://eng.uber.com)
- [Netflix Tech Blog](https://netflixtechblog.com)
- [Stripe Engineering](https://stripe.com/blog/engineering)
- [Cloudflare Blog](https://blog.cloudflare.com)
- [Meta Engineering](https://engineering.fb.com)
- [LinkedIn Engineering](https://engineering.linkedin.com/blog)
- [Shopify Engineering](https://shopify.engineering)

### Papers Worth Reading
- Google MapReduce (2004)
- Google File System (2003)
- Amazon Dynamo (2007)
- Google Bigtable (2006)
- Facebook TAO (2013)
- Kafka (LinkedIn, 2011)
- Raft Consensus (2014)
- Google Spanner (2012)

---

*You now have everything you need. The rest is practice. Design one system every day for 30 days and you'll be in the top 1%.*

**Go build something impossibly ambitious.**
