# Module 4: Distributed Systems

> The hardest part of system design. Master this and you've separated yourself from 95% of engineers.

---

## 4.1 — Why Distributed Systems Are Hard

### The Eight Fallacies of Distributed Computing

Engineers subconsciously assume:
1. The network is reliable → **It's not.** Packets get lost, connections drop.
2. Latency is zero → **It's not.** Cross-region calls are 50-300ms.
3. Bandwidth is infinite → **It's not.** You can saturate links.
4. The network is secure → **It's not.** Man-in-the-middle, eavesdropping.
5. Topology doesn't change → **It does.** Servers come and go.
6. There is one administrator → **There isn't.** Multiple teams, orgs, clouds.
7. Transport cost is zero → **It's not.** Data transfer costs money.
8. The network is homogeneous → **It's not.** Mixed hardware, protocols.

### Types of Failures

| Failure Type | Description | Example |
|-------------|-------------|---------|
| **Crash failure** | Process stops and stays stopped | Server loses power |
| **Omission failure** | Process fails to send/receive messages | Network drop |
| **Timing failure** | Response takes too long | Garbage collection pause |
| **Byzantine failure** | Process behaves arbitrarily (including maliciously) | Corrupted data, hacked server |

Most systems only handle crash and omission failures. Byzantine fault tolerance is needed for blockchain but is extremely expensive.

---

## 4.2 — Consistency Models

From strongest to weakest:

### Linearizability (Strongest)

- Every operation appears to take effect atomically at some point between invocation and response
- **All clients see the same order of operations**
- As if there's a single copy of the data
- Example: Single-machine database with proper locking
- **Cost:** High latency, low throughput, requires coordination

### Sequential Consistency

- All operations appear in some sequential order
- Each client's operations appear in the order they were issued
- But different clients might see different orderings from each other
- Still strong, but doesn't require real-time ordering

### Causal Consistency

- Operations that are causally related are seen in the same order by everyone
- Concurrent operations (no causal relation) can be seen in different orders
- Example: If A posts "Anyone want lunch?" and B replies "Yes!", everyone sees A before B
- But two unrelated posts can appear in different order for different users

### Eventual Consistency (Weakest practical)

- If no new writes occur, all replicas will **eventually** converge to the same value
- No guarantees on when
- Sufficient for many use cases (DNS, social media feeds, product catalogs)

### How to Choose

| Use Case | Consistency Needed | Why |
|----------|-------------------|-----|
| Bank transfers | Linearizable | Money can't appear/disappear |
| Inventory count | Linearizable | Can't oversell |
| Social media feed | Eventual | Stale feed for a few seconds is fine |
| User profile update | Read-your-writes | User should see their own changes |
| Collaborative editing | Causal | Related edits must be ordered |
| DNS | Eventual | TTL-based propagation is acceptable |

---

## 4.3 — Consensus Protocols

How do distributed nodes agree on a value?

### The Problem

```
Node A: "Set X = 5"
Node B: "Set X = 7"
Which value is correct? What if Node C is unreachable?
```

### Paxos

The original consensus algorithm (Lamport, 1989). Three roles:
- **Proposers** — Propose values
- **Acceptors** — Vote on proposals
- **Learners** — Learn the chosen value

Two phases:
1. **Prepare:** Proposer asks majority of acceptors to promise not to accept older proposals
2. **Accept:** If majority promise, proposer asks them to accept the value

**Problems:** Complex to implement, difficult to understand, single-decree (one decision at a time).

### Raft (The Understandable Consensus)

Designed to be easier than Paxos. Three states:
- **Leader** — Handles all writes, replicates to followers
- **Follower** — Passive, responds to leader
- **Candidate** — Trying to become leader

#### Leader Election
```
1. Follower's election timeout expires (no heartbeat from leader)
2. Follower becomes Candidate, increments term, votes for itself
3. Sends RequestVote to all other nodes
4. If majority votes → becomes Leader
5. Leader sends periodic heartbeats to maintain authority
```

#### Log Replication
```
1. Client sends write to Leader
2. Leader appends to its log
3. Leader sends AppendEntries to all Followers
4. Followers append to their logs, send ACK
5. Once majority ACK → Leader commits entry
6. Leader notifies Followers to commit
7. Leader responds to client
```

**Safety guarantee:** If a log entry is committed, all future leaders will have that entry.

**Used by:** etcd (Kubernetes), CockroachDB, TiDB, Consul.

### ZAB (ZooKeeper Atomic Broadcast)

Similar to Raft but designed for ZooKeeper:
- Leader handles all writes
- Atomic broadcast to followers
- Used by: ZooKeeper (Kafka depends on it, Hadoop, HBase)

### Comparison

| Property | Paxos | Raft | ZAB |
|----------|-------|------|-----|
| Understandability | Hard | Easy | Medium |
| Leader required | No | Yes | Yes |
| Implementation | Complex | Straightforward | Moderate |
| Used by | Google Chubby | etcd, CockroachDB | ZooKeeper |

---

## 4.4 — Distributed Clocks & Ordering

### The Problem with Wall Clocks

Two servers' clocks can differ by milliseconds to seconds. You **cannot** rely on timestamps for ordering in distributed systems.

### Logical Clocks

#### Lamport Clocks
```
Rules:
1. Before each event, increment local counter
2. When sending message, include counter
3. When receiving message, set counter = max(local, received) + 1
```

Gives you: If A happened before B, then Clock(A) < Clock(B)
Does NOT give you: If Clock(A) < Clock(B), then A happened before B

#### Vector Clocks

```
Each node maintains a vector of counters [N1: x, N2: y, N3: z]

Node 1 sends: [1:3, 2:0, 3:0]  (incremented its own counter)
Node 2 receives: [1:3, 2:1, 3:0]  (merged, incremented its own)
```

Gives you: True causality detection
- If V(A) < V(B) in all positions → A happened before B
- If neither ≤ the other → A and B are concurrent (conflict!)

**Used by:** DynamoDB, Riak

#### Hybrid Logical Clocks (HLC)

Combines physical time with logical clock:
- Uses physical timestamp when possible
- Falls back to logical counter when physical clocks are close
- Gives "close enough" to real time while maintaining causal ordering

**Used by:** CockroachDB, MongoDB

### Google TrueTime

Hardware-based solution. GPS + atomic clocks in every data center.

```
TrueTime.now() returns an interval: [earliest, latest]
Example: [12:00:00.001, 12:00:00.005]

Actual time is guaranteed to be within this interval.
Uncertainty is typically < 7ms.
```

Spanner uses this to implement linearizable reads across the globe:
- Wait out the uncertainty interval before committing
- If uncertainty is 7ms, wait 7ms after commit → guaranteed ordering

**This is how Google Spanner achieves global strong consistency.** No one else has this (as a managed service).

---

## 4.5 — Distributed Transactions

### Two-Phase Commit (2PC)

```
Phase 1 (Prepare):
  Coordinator → All Participants: "Can you commit?"
  Each Participant: Write to WAL, lock resources → "Yes" or "No"

Phase 2 (Commit):
  If all say Yes: Coordinator → All: "Commit"
  If any says No:  Coordinator → All: "Abort"
```

**Problems:**
- **Blocking:** If coordinator crashes between phases, participants hold locks forever
- **Performance:** All participants block during voting
- **Single point of failure:** Coordinator

### Three-Phase Commit (3PC)

Adds a "Pre-Commit" phase to reduce blocking. Rarely used in practice.

### Saga Pattern (Preferred for Microservices)

Instead of distributed transactions, use a sequence of local transactions with compensating actions.

```
Book Flight → Book Hotel → Charge Credit Card

If "Charge Credit Card" fails:
  Cancel Hotel Booking (compensation)
  Cancel Flight Booking (compensation)
```

#### Orchestration Saga
```
[Saga Orchestrator]
  → "Book flight" → Flight Service
  → "Book hotel"  → Hotel Service
  → "Charge card" → Payment Service
  
If failure at any step, orchestrator triggers compensating actions
```
- Central control, easy to reason about
- Single point of failure (the orchestrator)

#### Choreography Saga
```
Flight Service → (event: FlightBooked) → Hotel Service
Hotel Service → (event: HotelBooked) → Payment Service
Payment Service → (event: PaymentFailed) → Hotel Service (cancel)
                                         → Flight Service (cancel)
```
- No central coordinator
- Harder to debug and understand
- More resilient

### Outbox Pattern

Ensures exactly-once publishing of events alongside database writes.

```
Transaction:
  1. Write business data to DB
  2. Write event to "outbox" table in same DB (same transaction!)

Separate process:
  3. Read from outbox table
  4. Publish to message queue
  5. Mark as published
```

This avoids the dual-write problem (writing to DB and queue separately can fail partially).

---

## 4.6 — Failure Detection

### Heartbeat

```
Every node sends "I'm alive" every T seconds.
If no heartbeat for N*T seconds → consider node dead.
```

**Problem:** Network partition ≠ node failure. The node might be alive but unreachable.

### Phi Accrual Failure Detector

Instead of binary "alive/dead":
- Maintains a probability of failure (phi value)
- Based on historical heartbeat intervals
- Phi = 1 → 10% chance of failure
- Phi = 2 → 1% chance of failure
- Threshold is configurable

**Used by:** Cassandra, Akka

### Gossip Protocol

```
Every T seconds:
  1. Pick a random node
  2. Exchange state information (who's alive, who's dead)
  3. Merge information
```

- Epidemic-style information spreading
- Eventually all nodes know about all other nodes
- No single point of failure
- Scalable (O(log N) rounds to propagate)

**Used by:** Cassandra (membership), DynamoDB, Consul

### Split-Brain Problem

When a network partition creates two groups of nodes, each thinking the other is dead:

```
[A, B, C] ←partition→ [D, E]

Group 1 thinks: "D and E are dead, we're the cluster"
Group 2 thinks: "A, B, C are dead, we're the cluster"
Both accept writes → data divergence!
```

**Solutions:**
- **Quorum:** Need majority to operate (3/5 can, 2/5 can't)
- **Fencing tokens:** Monotonically increasing tokens; old leader's writes are rejected
- **STONITH:** "Shoot The Other Node In The Head" — one group forcibly kills the other

---

## 4.7 — Bloom Filters & Probabilistic Data Structures

### Bloom Filter

Space-efficient probabilistic membership test.

```
"Is element X in the set?"
  → "Definitely not" (100% certain)
  → "Probably yes" (small false positive rate)
```

**How it works:**
1. Bit array of m bits, k hash functions
2. To add: hash element k times, set those bits to 1
3. To check: hash element k times, check if all bits are 1

**Properties:**
- No false negatives (if it says no, it's definitely no)
- Small false positive rate (tunable, typically 1%)
- Cannot remove elements (use Counting Bloom Filter for that)
- Very space efficient (10 bits per element for 1% FP rate)

**Used for:**
- Cache penetration prevention (check if key exists before hitting DB)
- Spell checkers
- Network routers (packet filtering)
- Cassandra (check if SSTable might contain a key)
- Chrome Safe Browsing (check URL against blocklist)

### Count-Min Sketch

Estimate frequency of elements in a stream.

```
"How many times has element X appeared?"
Answer: At least N times (may overcount, never undercount)
```

**Used for:** Top-K queries, heavy hitters detection, rate limiting.

### HyperLogLog

Estimate cardinality (count distinct elements) using very little memory.

```
"How many unique visitors today?"
Answer: ~1,234,567 (±2% error) using only 12 KB of memory
```

**Used by:** Redis (PFCOUNT), BigQuery, Presto

---

## 4.8 — Leader Election

### Why Leader Election?

Many systems need a single leader for:
- Write coordination (databases)
- Task assignment (schedulers)
- Consensus (Raft/Paxos leader)

### Implementation Approaches

| Approach | How | Used By |
|----------|-----|---------|
| **ZooKeeper/etcd** | Ephemeral lock nodes | Kafka, HBase, Kubernetes |
| **Raft** | Built into consensus | CockroachDB, TiDB |
| **Database lock** | Advisory lock or row lock | Simple leader election |
| **Cloud-native** | DynamoDB conditional writes | Custom applications |

### ZooKeeper Leader Election
```
1. All candidates create ephemeral sequential znodes:
   /election/node_0001, /election/node_0002, ...
2. Node with smallest sequence number is leader
3. Others watch the node with next-lower sequence
4. If leader dies, ephemeral znode disappears
5. Watcher fires, new leader elected
```

---

## 4.9 — Idempotency

### Why It Matters

In distributed systems, messages can be delivered more than once. Operations must be safe to retry.

```
Client → "Transfer $100" → Server
Server processes, sends ACK
ACK is lost in network
Client retries → "Transfer $100" → Server
Without idempotency: $200 transferred!
With idempotency: $100 transferred (retry detected and ignored)
```

### Implementation

```
// Client includes unique idempotency key with each request
POST /payment
Idempotency-Key: "abc-123-def"
{ amount: 100, from: "Alice", to: "Bob" }

// Server logic:
1. Check if "abc-123-def" already processed
2. If yes → return cached response
3. If no → process payment, store result with key "abc-123-def"
```

### Naturally Idempotent Operations
```
SET balance = 500    ← Idempotent (same result every time)
DELETE user/123      ← Idempotent (deleting twice = same outcome)

INCREMENT balance    ← NOT idempotent (incrementing twice = wrong)
INSERT row           ← NOT idempotent (duplicate records)
```

---

## 4.10 — Exercises

1. **Consistency choice:** You're designing a global multiplayer game. What consistency model do you use for player positions? For in-game purchases?

2. **Clock problem:** Server A thinks it's 12:00:00.000 and Server B thinks it's 12:00:00.050. Both process a write at their "12:00:00.010". Which came first? How would you solve this with (a) vector clocks, (b) HLC?

3. **Saga design:** Design a saga for an e-commerce order: reserve inventory → process payment → schedule delivery. How do you handle each compensation?

4. **Bloom filter sizing:** You have 100M URLs and want a false positive rate of 0.1%. How many bits does your Bloom filter need? (Formula: m = -(n × ln(p)) / (ln(2))²)

---

**Next:** [Module 5 — Infrastructure & Reliability](05-infrastructure.md) →
