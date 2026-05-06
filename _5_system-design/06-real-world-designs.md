# Module 6: Real-World System Designs

> Theory is worthless without practice. Here are 15+ complete system designs, from simple to planet-scale.

---

## How to Read These Designs

Each design follows the same structure:
1. **Requirements** — What we're building
2. **Estimations** — Scale numbers
3. **High-Level Design** — The architecture diagram
4. **Deep Dives** — Key components in detail
5. **Trade-offs** — What we chose and why

---

## Design 1: URL Shortener (Easy)

### Requirements
- Shorten long URLs (like bit.ly)
- Redirect short URL to original
- Custom aliases (optional)
- Analytics (click count)
- 100M URLs created per month, 10:1 read/write ratio

### Estimations
```
Writes: 100M / month = ~40 URLs/sec
Reads:  1B / month   = ~400 redirects/sec
Storage (5 years): 100M × 12 × 5 × 500 bytes = 3 TB
```

### High-Level Design
```
[Client] → [Load Balancer] → [API Servers] → [Database]
                                    ↕
                              [Cache (Redis)]
```

### Key Decisions

**Short URL generation:**
- Option A: Hash (MD5/SHA) the URL → take first 7 chars → base62 encode
  - Collision handling: check DB, if exists, append salt and rehash
- Option B: Pre-generate unique IDs (counter/Snowflake) → base62 encode
  - No collisions, but need distributed counter
- **Choose B:** Simpler, no collisions

**Base62:** `[a-zA-Z0-9]` = 62 chars. 7 chars = 62^7 = 3.5 trillion combinations. Enough.

**Database:** Simple KV lookup → DynamoDB or Redis + PostgreSQL

**Caching:** 80/20 rule — 20% of URLs get 80% of traffic. Cache top URLs in Redis.

**Read path:**
```
GET /abc1234
  → Check Redis cache
  → Cache hit → 301/302 redirect
  → Cache miss → Query DB → Cache result → Redirect
```

**301 vs 302 redirect:**
- 301 (Permanent): Browser caches. Better for SEO. Lose analytics.
- 302 (Temporary): Browser always hits server. Full analytics. **Choose this.**

---

## Design 2: Rate Limiter (Easy)

### Requirements
- Limit API requests per user/IP
- Distributed (multiple API servers share limits)
- Low latency (must not slow down requests)
- Return informative headers

### High-Level Design
```
[Client] → [API Gateway / Rate Limiter Middleware] → [Redis] → [API Servers]
```

### Implementation: Sliding Window Counter in Redis
```
Key: rate_limit:{user_id}:{minute_window}
Value: count

MULTI
  INCR rate_limit:user123:202401151030
  EXPIRE rate_limit:user123:202401151030 120  # 2 min TTL
EXEC

if count > limit → return 429 Too Many Requests
```

### Rules Engine
```yaml
rules:
  - name: "API general"
    key: "user_id"
    limit: 100
    window: 60s
  - name: "Login attempts"
    key: "ip"
    limit: 5
    window: 300s
  - name: "Premium users"
    key: "user_id"
    condition: "plan == premium"
    limit: 1000
    window: 60s
```

---

## Design 3: Notification System (Medium)

### Requirements
- Push notifications (iOS, Android), SMS, Email
- 100M notifications/day
- Pluggable providers
- Retry on failure
- Rate limiting per user (don't spam)
- User preferences (opt-in/out per channel)

### High-Level Design
```
[Notification Service API]
         ↓
[Validation + Rate Limiting + Preference Check]
         ↓
[Message Queue (Kafka)] → [Priority Queue: High | Medium | Low]
         ↓
[Worker Pool]
  → [Push Worker] → [APNS / FCM]
  → [SMS Worker]  → [Twilio / SNS]
  → [Email Worker] → [SendGrid / SES]
         ↓
[Delivery Status Tracker] → [Analytics DB]
```

### Key Components
- **Template engine:** Renders messages from templates + variables
- **User preference store:** Which channels per notification type
- **Deduplication:** idempotency key prevents duplicate sends
- **DLQ:** Failed notifications after retries → manual review
- **Analytics:** Open rates, click rates, delivery rates per channel

---

## Design 4: Chat System (Medium)

### Requirements
- 1-on-1 and group chat
- Online/offline status
- Sent/delivered/read receipts
- 50M DAU
- Message history

### Estimations
```
50M DAU, 40 messages/user/day = 2B messages/day
Peak: 2B / 86400 × 3 (peak factor) = ~70K messages/sec
Message size: ~200 bytes
Storage/day: 2B × 200B = 400 GB/day
```

### High-Level Design
```
[Mobile/Web Client]
       ↕ (WebSocket)
[WebSocket Servers (stateful)]
       ↓
[Chat Service]
  → [Message Queue (Kafka)]
    → [Message Storage (Cassandra)]
    → [Push Notification Service]
    → [Presence Service (Redis)]

[Connection Registry (Redis)]
  Maps: user_id → which WebSocket server
```

### Key Decisions

**WebSocket vs Long Polling:** WebSocket. Bidirectional, low latency, efficient.

**Message storage:** Cassandra
- Partition key: `conversation_id`
- Clustering key: `message_id` (time-based, e.g., Snowflake)
- Recent messages: hot in memory
- Old messages: on disk (Cassandra handles this naturally)

**Message ordering:**
- Use Snowflake IDs (timestamp + machine + sequence)
- Within a conversation, messages are ordered by Snowflake ID

**Online status:**
```
Redis:
  presence:{user_id} → { last_seen: timestamp, status: "online" }
  TTL: 30 seconds (no heartbeat → offline)

Client sends heartbeat every 10 seconds via WebSocket
```

**Group chat fan-out:**
- Small groups (< 100): Write message once, fan-out on read (each member queries)
- Large groups / channels: Write once, push notifications to active members

**Message delivery flow:**
```
1. Alice sends message via WebSocket to her WS server
2. WS server → Chat Service → Kafka (topic: messages)
3. Consumer reads from Kafka
4. Store in Cassandra
5. Lookup Bob's WebSocket server from Redis
6. Forward message to Bob's WS server → Bob's client
7. If Bob offline → Push notification
```

---

## Design 5: News Feed / Timeline (Medium)

### Requirements
- User posts content
- Feed shows posts from followed users
- 500M DAU, 5000 avg followers per user
- Feed should feel real-time (< 5 sec delay)

### The Central Problem: Fan-Out

**Fan-out on Write (Push model):**
```
User posts → For each follower, write to their feed cache

Pros: Feed reads are fast (pre-computed)
Cons: Celebrity problem — user with 100M followers = 100M writes per post
```

**Fan-out on Read (Pull model):**
```
User opens feed → Fetch latest posts from all followed users → Merge and sort

Pros: No hot-write problem
Cons: Slow reads (fetch from many sources)
```

**Hybrid (What Twitter/Instagram actually does):**
```
Regular users (< 10K followers): Fan-out on write
Celebrities (> 10K followers): Fan-out on read

When building feed:
  1. Read pre-computed feed from cache (regular users' posts)
  2. Fetch celebrity posts on-the-fly
  3. Merge, rank, return
```

### Architecture
```
Post Service:
  [User] → [Post API] → [Post DB] → [Fan-out Service]
                                          ↓
                               [Feed Cache (Redis)] (for each follower)

Feed Service:
  [User] → [Feed API] → [Feed Cache] + [Celebrity Posts] → Merge → Rank → Return
```

### Feed Ranking
```
score = f(
  recency,           # newer = higher
  engagement,        # likes, comments, shares
  relationship,      # close friends boost
  content_type,      # video > image > text (engagement-based)
  creator_quality,   # high-quality creators
  user_preferences   # ML model trained on user behavior
)
```

---

## Design 6: Distributed ID Generator (Medium)

### Requirements
- Globally unique IDs
- Roughly time-sortable
- High throughput (10K+ IDs/sec per node)
- 64-bit numeric (fit in a long)

### Twitter Snowflake Approach

```
64 bits:
  [1 bit: unused] [41 bits: timestamp] [5 bits: datacenter] [5 bits: machine] [12 bits: sequence]

  Timestamp: milliseconds since epoch → 69 years
  Datacenter: 32 datacenters
  Machine: 32 machines per DC
  Sequence: 4096 IDs per millisecond per machine

  Total capacity: 4096 × 1000 × 1024 machines = 4 billion IDs/sec
```

### Alternatives

| Approach | Pros | Cons |
|----------|------|------|
| **UUID v4** | No coordination, simple | 128-bit (not 64-bit), not sortable, bad for DB index |
| **UUID v7** | Time-sorted UUID | 128-bit |
| **Snowflake** | 64-bit, sortable, fast | Requires machine ID assignment |
| **DB auto-increment** | Simple, sequential | Single point of failure, bottleneck |
| **DB ticket server (Flickr)** | Two DBs, odd/even | Still limited throughput |
| **ULID** | Lexicographically sortable, 128-bit | Library support varies |

---

## Design 7: Web Crawler (Medium)

### Requirements
- Crawl billions of web pages
- Polite (respect robots.txt, rate limits)
- Handle duplicates
- Extract and store content

### Architecture
```
[Seed URLs]
     ↓
[URL Frontier (Priority Queue)]
     ↓
[DNS Resolver (cached)]
     ↓
[Fetcher Pool (distributed)]
     ↓
[Content Parser]
  → [Duplicate Detector (Bloom Filter + SimHash)]
  → [URL Extractor] → [URL Filter] → back to Frontier
  → [Content Store (S3)] → [Index (Elasticsearch)]
```

### Key Components

**URL Frontier:**
- Priority queue (important pages first: PageRank, freshness)
- Politeness queue (one queue per domain, rate limited)
- Separate front queue (priority) and back queue (politeness)

**Duplicate detection:**
- **URL dedup:** Bloom filter on normalized URLs
- **Content dedup:** SimHash or MinHash on page content (detect near-duplicates)

**Politeness:**
```
Per domain:
  Respect robots.txt directives
  Max 1 request per second per domain
  Exponential backoff on errors
```

---

## Design 8: Distributed Key-Value Store (Hard)

### Requirements
- Simple API: `get(key)`, `put(key, value)`
- Highly available (AP system)
- Tunable consistency
- Partition tolerant
- Scalable to petabytes

### This is essentially building a simplified DynamoDB/Cassandra.

### Architecture
```
[Client] → [Coordinator Node (any node)] → [Replicas for that key's partition]

Data placement:
  Consistent hashing ring → Partition → N replicas (e.g., 3)
```

### Key Components

**Consistent hashing with virtual nodes** — Data distribution

**Quorum reads/writes:**
```
N = 3 replicas
W = 2 (write to 2 before ACK)
R = 2 (read from 2, take latest)
W + R > N → Strong consistency
W = 1, R = 1 → Fastest, eventual consistency
```

**Conflict resolution:** Vector clocks + last-write-wins or application-resolved

**Failure handling:**
- **Sloppy quorum + hinted handoff:** If a replica is down, write to another node temporarily. When recovered, hand off the data.
- **Anti-entropy:** Merkle trees to sync replicas in background

**Data path:**
```
Write: Client → Coordinator → Write to memtable → WAL (write-ahead log)
       → When memtable full → Flush to SSTable on disk
       → Background compaction merges SSTables

Read:  Client → Coordinator → Check memtable → Check SSTables (newest first)
       → Bloom filter to skip SSTables that don't have the key
```

---

## Design 9: Search Autocomplete / Typeahead (Medium)

### Requirements
- Show top 5 suggestions as user types
- < 100ms latency
- Suggestions based on popularity/relevance
- Multi-language support
- 5B searches/day

### Data Structure: Trie

```
       root
      / | \
     t   b  ...
    /     |
   tr     be
   |      |
   tre    bee
   |      |
   tree   beer
  (freq:5000)  (freq:3000)
```

Each trie node stores:
- Character
- Children
- Top-K most frequent completions at this prefix (precomputed)

### Architecture
```
[Client] → [API Gateway] → [Autocomplete Service]
                                    ↓
                         [Distributed Trie (sharded by prefix)]
                                    ↓
                         [Redis Cache for hot prefixes]

Data Collection:
  [Search logs] → [Kafka] → [Analytics] → [Trie Builder (hourly/daily)]
```

### Key Decisions

**Trie sharding:**
- Shard by prefix range: a-f → Shard 1, g-m → Shard 2, etc.
- Uneven distribution → use weighted sharding based on query volume

**Freshness:**
- Base trie: rebuilt nightly from search analytics
- Trending overlay: updated every few minutes for trending/breaking topics

**Ranking:**
```
score = query_frequency × recency_weight × personalization_boost
```

---

## Design 10: YouTube / Video Streaming Platform (Hard)

### Requirements
- Upload and stream video
- 2B MAU, 500 hours uploaded/minute
- Adaptive bitrate streaming
- Global distribution

### Upload Pipeline
```
[Client Upload]
     ↓ (multipart, resumable)
[Upload Service] → [Object Storage (S3)]
     ↓
[Message Queue]
     ↓
[Transcoding Service (distributed, GPU)]
  → 1080p, 720p, 480p, 360p, 240p
  → Multiple codecs (H.264, VP9, AV1)
  → Audio extraction
  → Thumbnail generation
     ↓
[Transcoded files → CDN Origin → CDN Edge]
     ↓
[Metadata Service] → [Metadata DB (MySQL/Vitess)]
```

### Streaming Architecture
```
[Client] → [CDN Edge (closest PoP)]
              → Cache HIT → Stream from CDN
              → Cache MISS → Fetch from origin → Cache → Stream
```

**Adaptive Bitrate Streaming (ABR):**
- Video split into 2-10 second segments
- Each segment encoded at multiple bitrates
- Client measures bandwidth, switches quality dynamically
- Protocols: HLS (Apple), DASH (open standard)

```
manifest.m3u8:
  1080p/segment001.ts  (5 Mbps)
  720p/segment001.ts   (2.5 Mbps)
  480p/segment001.ts   (1 Mbps)
  360p/segment001.ts   (0.5 Mbps)

Client on fast WiFi → 1080p
Client enters tunnel → switches to 360p
Client exits tunnel → gradually back to 1080p
```

### Recommendation Engine
```
[User Activity] → [Kafka] → [Feature Store]
                         → [ML Training Pipeline (Spark/TF)]
                              ↓
                         [Model Serving (TF Serving)]
                              ↓
[Recommendation API] → [Blend: collaborative filtering + content-based + trending]
```

---

## Design 11: Uber / Ride-Sharing (Hard)

### Requirements
- Match riders with nearby drivers
- Real-time location tracking
- ETA calculation
- Surge pricing
- 20M rides/day

### Core Problem: Geospatial matching

### Architecture
```
Rider App                                 Driver App
    ↓                                          ↓
[API Gateway]                          [API Gateway]
    ↓                                          ↓
[Trip Service]                         [Location Service]
    ↓                                    (driver GPS every 3s)
[Matching Service]  ←  [Supply Service]       ↓
    ↓                   (available drivers) [Geospatial Index]
[Dispatch Service]
    ↓
[Notification] → Driver gets ride request
```

### Geospatial Indexing

**Option 1: Geohash**
```
Earth divided into grid cells with hierarchical string IDs:
  "9q8yyk" (San Francisco, ~1.2km × 0.6km)
  "9q8yy"  (larger area)
  "9q8y"   (even larger)

To find nearby drivers:
  1. Compute geohash of rider's location
  2. Query for drivers in same geohash + 8 neighboring geohashes
  3. Use Redis: GEORADIUS or sorted set of geohash → driver_ids
```

**Option 2: Quadtree**
```
Recursively divide space into 4 quadrants.
Split a cell when it contains > N drivers.
Dynamically adapts to driver density.
```

**Option 3: S2 Geometry (What Uber actually uses)**
- Maps Earth onto a cube, projects cube faces to squares
- Hierarchical cells at 30 levels of precision
- Excellent for covering regions with minimal cells

### Matching Algorithm
```
1. Rider requests ride at location (lat, lng)
2. Find available drivers within radius (expanding if needed)
3. Score each driver:
   score = f(distance, ETA, driver_rating, cancellation_rate, acceptance_rate)
4. Send request to best driver
5. Driver accepts/rejects (15 second timeout)
6. If rejected → next best driver
```

### ETA Calculation
- Graph of road network (weighted graph)
- Dijkstra's algorithm with real-time traffic weights
- Precomputed routing tables for common corridors
- ML model adjusting for time of day, weather, events

---

## Design 12: Twitter / Social Network (Hard)

### Requirements
- Post tweets (280 chars + media)
- Follow/unfollow users
- Home timeline (feed of followed users)
- Search tweets
- Trending topics
- 500M DAU

### Architecture Overview
```
[Client] → [API Gateway] → [Tweet Service]
                         → [Timeline Service]
                         → [Search Service]
                         → [User Service]
                         → [Notification Service]

[Tweet Service] → [Tweet DB (sharded by tweet_id)]
               → [Fan-out Service] → [Timeline Cache (Redis)]
               → [Search Indexer] → [Elasticsearch]

[Media] → [Object Storage + CDN]
```

### Timeline Service (The Hard Part)

See Design 5: News Feed for the hybrid fan-out approach.

### Trending Topics
```
[Tweet stream] → [Kafka] → [Stream Processor (Flink)]
  → Sliding window count of hashtags/topics
  → Weighted by: recency, acceleration (rate of growth), novelty
  → Per-region trending
  → Updated every 30 seconds
  → [Trending Cache (Redis)]
```

### Search
```
[Elasticsearch cluster]
  → Full-text index on tweet text
  → Inverted index: word → list of tweet_ids
  → Real-time indexing (< 10 sec from post to searchable)
  → Ranking: relevance + recency + engagement + user authority
```

---

## Design 13: Distributed Cache (Hard)

### Requirements
- Sub-millisecond reads
- 100M+ ops/sec across cluster
- Handle hot keys / thundering herd
- Consistent hashing for distribution
- Cache-aside pattern

### Architecture (Think: scaled Redis cluster)
```
[Client Library (consistent hashing)] → [Cache Node 1]
                                      → [Cache Node 2]
                                      → [Cache Node N]
```

### Hot Key Problem
```
A single key gets 50% of all traffic (e.g., celebrity tweet)

Solutions:
1. Local cache: L1 (in-process) → L2 (remote cache) → DB
2. Key replication: Replicate hot key to all cache nodes, random read
3. Key splitting: Split "hot_key" into "hot_key:1", "hot_key:2", ..., "hot_key:N"
   Client randomly picks one
```

### Thundering Herd (Cache Stampede)
```
Popular key expires → 10K requests simultaneously hit DB

Solutions:
1. Locking: First request sets lock, others wait
   SET lock:key 1 NX EX 5  (Redis atomic lock)
2. Early expiration: Refresh before actual TTL
3. Stagger TTL: Add random jitter to TTL
```

---

## Design 14: Payment System (Hard)

### Requirements
- Process payments reliably
- Exactly-once processing (cannot double-charge)
- Handle failures gracefully
- Compliance (PCI DSS)
- Reconciliation

### Architecture
```
[Checkout] → [Payment Service] → [Payment State Machine (DB)]
                    ↓
            [Payment Executor]
                    ↓
    [Payment Gateway (Stripe/Adyen)]
                    ↓
            [Callback/Webhook Handler]
                    ↓
            [Ledger Service (double-entry bookkeeping)]
                    ↓
            [Reconciliation (batch, daily)]
```

### Payment State Machine
```
CREATED → PROCESSING → APPROVED → CAPTURED → SETTLED
                    → DECLINED
                    → FAILED → RETRYING → PROCESSING
CAPTURED → REFUND_REQUESTED → REFUNDED
```

### Exactly-Once with Idempotency
```
POST /payments
Idempotency-Key: "order-12345-attempt-1"

Server:
  1. Check if idempotency key exists in DB
  2. If exists → return cached result
  3. If not → process payment
  4. Store result with idempotency key (in same transaction as payment record)
```

### Double-Entry Ledger
```
Every payment creates two entries:
  Debit:  Buyer's account  -$100
  Credit: Seller's account +$100

Every entry must balance. Sum of all debit = Sum of all credit.
This makes reconciliation and auditing possible.
```

---

## Design 15: Google Maps / Navigation (Expert)

### Requirements
- Map rendering (tiles)
- Geocoding (address → coordinates)
- Route planning (A to B)
- Real-time traffic
- ETA

### Map Rendering
```
World divided into tiles at each zoom level:
  Zoom 0: 1 tile (whole world)
  Zoom 1: 4 tiles
  Zoom 2: 16 tiles
  ...
  Zoom 18: 68 billion tiles

[Client] → [Tile CDN] → [Tile Rendering Service] → [Map Data DB]

Vector tiles (modern): Send geometry data, render on client (smaller, flexible styling)
Raster tiles (legacy): Pre-rendered images (simple but bandwidth-heavy)
```

### Route Planning
```
Road network = Weighted graph
  Nodes: intersections
  Edges: road segments
  Weights: distance, time (based on speed limit + traffic)

Algorithms:
  1. Dijkstra's: Basic shortest path. Too slow for continental routing.
  2. A*: Heuristic-guided Dijkstra. Better but still slow for long distances.
  3. Contraction Hierarchies: Precompute shortcuts for "important" nodes.
     Query: milliseconds for continental routes.
  4. ALD/ALT: Landmark-based heuristics.

What Google uses:
  Contraction Hierarchies + real-time traffic overlays + ML for ETA adjustment
```

### Real-Time Traffic
```
[Millions of phones sending GPS + speed data]
     ↓
[Kafka] → [Traffic Processing (Flink)]
     ↓
[Traffic State DB (road segment → current speed)]
     ↓
[Route Planning uses traffic weights]
[Map tiles colored by traffic speed]
```

---

## Design 16: Distributed Task Scheduler (Expert)

### Requirements
- Schedule one-time and recurring tasks
- At-least-once execution
- Handle millions of tasks
- Distributed, no single point of failure
- Priority support

### Architecture
```
[API] → [Task Store (MySQL, sharded by task_id)]

[Scheduler Nodes (multiple, using leader election)]
  → Scan for tasks due in next window
  → Place on execution queue (Kafka/SQS)

[Worker Fleet]
  → Pull tasks from queue
  → Execute
  → Report status back

[Dead Task Handler]
  → Tasks that failed N times
  → Alerting + manual intervention
```

### Timing: How to schedule at precise times
```
Option 1: Polling DB every second (simple, wasteful at scale)
Option 2: Hierarchical timing wheels
  - Tasks bucketed by time window
  - Multiple wheel levels (seconds, minutes, hours)
  - O(1) insertion and firing
  Used by: Kafka, Netty, Akka
```

---

## 6.1 — Exercises

Pick 3 designs from above and:
1. Draw the architecture diagram from memory
2. Identify the top 3 bottlenecks in each
3. Propose how you'd scale each to 10x its stated traffic
4. Identify what changes if you need 99.999% availability

---

**Next:** [Module 7 — Interview Framework](07-interview-framework.md) →
