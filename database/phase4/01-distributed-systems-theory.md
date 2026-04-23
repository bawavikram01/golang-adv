# 4.1 — Distributed Systems Theory

> You CANNOT master distributed databases without understanding the theory.  
> Every design decision in Cassandra, Spanner, DynamoDB, CockroachDB  
> traces back to these fundamental results and trade-offs.  
> This is the hardest chapter. It is also the most important.

---

## 1. Why Distributed?

```
Single-node databases hit walls:
  1. Storage limit:     A single server has finite disk
  2. Throughput limit:  A single CPU/NIC can only handle so much
  3. Availability:      Single server = single point of failure
  4. Latency:           Users across the globe → laws of physics (speed of light)

Distributed databases solve these by spreading data across multiple machines.
But distribution introduces an entirely new class of problems.
```

---

## 2. CAP Theorem

```
Brewer's CAP Theorem (2000, proved by Gilbert & Lynch 2002):

In a distributed system, during a NETWORK PARTITION, you can have at most:
  Consistency (C): every read returns the most recent write (linearizability)
  Availability (A): every request gets a non-error response (no timeouts)
  Partition tolerance (P): system operates despite network partitions

You MUST tolerate partitions (P) in a real distributed system.
So the real choice is: C or A during a partition.

                      C
                     / \
                    /   \
                   /     \
              CP systems  CA systems
              (Spanner,    (single-node
               CockroachDB, RDBMS — not
               HBase,       truly
               etcd)        distributed)
                  \       /
                   \     /
                    \   /
                     \ /
                      P ───────── A
                           AP systems
                           (Cassandra,
                            DynamoDB,
                            Riak,
                            CouchDB)

CRITICAL NUANCES — why "CAP" is misleading:

1. Partitions are RARE. CAP only matters during a partition.
   The rest of the time you can have C AND A.

2. Consistency is a SPECTRUM, not binary:
   Linearizable > Sequential > Causal > Read-your-writes > Eventual
   Most "AP" systems offer tunable consistency.

3. Availability is also a spectrum (99.9% vs 99.99% vs 99.999%).

4. CAP says nothing about LATENCY — the real trade-off in practice.
```

---

## 3. PACELC Theorem

```
PACELC (Daniel Abadi, 2012) — a more useful framework than CAP:

If there's a Partition (P):
  Choose Availability (A) or Consistency (C)
Else (E) — in normal operation:
  Choose Latency (L) or Consistency (C)

                 Partition?
                /          \
              YES           NO
             /               \
         A or C?          L or C?
         
Database       | P: A or C | E: L or C | Classification
─────────────────────────────────────────────────────────
Spanner        | C         | C         | PC/EC (always consistent)
CockroachDB    | C         | C         | PC/EC
Cassandra      | A         | L         | PA/EL (always fast)
DynamoDB       | A         | L         | PA/EL (default)
MongoDB        | C         | C         | PC/EC (with majority reads)
PostgreSQL     | C         | C         | PC/EC (single node, trivially)
Cosmos DB      | Tunable   | Tunable   | Tunable at per-request level

PACELC captures the REAL daily trade-off: latency vs consistency in normal operation.
CAP's partition scenario happens maybe once a year. Latency vs consistency happens on every request.
```

---

## 4. Consistency Models — The Full Spectrum

```
From STRONGEST to WEAKEST:

┌────────────────────────────────────────────────────────────┐
│ LINEARIZABILITY (strongest)                                 │
│ "As if there's a single copy and every operation is atomic" │
│ Every read sees the absolute latest write.                  │
│ Operations happen at a single point in real time.           │
│ Example: Spanner (with TrueTime), etcd, ZooKeeper          │
│ Cost: highest latency, cross-datacenter round-trips         │
├────────────────────────────────────────────────────────────┤
│ SEQUENTIAL CONSISTENCY                                      │
│ All nodes see operations in the SAME ORDER,                 │
│ but that order doesn't have to match real-time order.       │
│ Example: ZooKeeper reads (stale but ordered)                │
├────────────────────────────────────────────────────────────┤
│ CAUSAL CONSISTENCY                                          │
│ Operations that are causally related are seen in order.     │
│ Concurrent (unrelated) operations may be seen in any order. │
│ Example: MongoDB (causal sessions)                          │
│ Practical and performant — the "sweet spot" for many apps.  │
├────────────────────────────────────────────────────────────┤
│ READ-YOUR-WRITES                                            │
│ A process always sees its own writes.                       │
│ Other processes might see stale data.                       │
│ Example: most web apps with sticky sessions                 │
├────────────────────────────────────────────────────────────┤
│ EVENTUAL CONSISTENCY (weakest useful)                       │
│ If no new writes happen, all replicas EVENTUALLY converge.  │
│ No timing guarantee. Can read stale data for seconds/mins.  │
│ Example: DNS, Cassandra (at CL=ONE), DynamoDB (default)     │
│ Cost: lowest latency, highest availability                  │
└────────────────────────────────────────────────────────────┘

Linearizability vs Serializability (commonly confused):
  Linearizability: single-object, real-time ordering guarantee
  Serializability:  multi-object (transactions), global serial order
  
  Strict Serializability = Linearizability + Serializability
    → The gold standard. Spanner provides this.
    → Transactions execute as if serial, AND respect real-time order.
```

---

## 5. FLP Impossibility Result

```
Fischer, Lynch, Paterson (1985):

In an asynchronous distributed system where at least one process may crash:
  It is IMPOSSIBLE to guarantee consensus will be reached in finite time.

What this means:
  You CANNOT build a consensus algorithm that is:
    ✓ Always safe (never decides wrong)
    ✓ Always live (always terminates)
    ✓ Asynchronous (no timing assumptions)
  
  Real systems work around this by using PARTIAL SYNCHRONY:
    "The network is usually fast, and we use timeouts to detect failures."
    Paxos, Raft, etc. sacrifice liveness during async periods
    (they don't make wrong decisions, they just pause).

  FLP is why distributed consensus is fundamentally hard.
  Every consensus algorithm makes timing/failure assumptions to sidestep FLP.
```

---

## 6. Consensus Algorithms

### Paxos (Lamport, 1989)

```
Paxos solves: how can a group of unreliable nodes agree on a single value?

Roles:
  Proposer:  proposes a value
  Acceptor:  votes on proposals
  Learner:   learns the decided value
  (A node can play multiple roles)

Phase 1 (Prepare):
  Proposer → Acceptors: "PREPARE(n)"  (n = unique proposal number)
  Acceptors: if n > highest seen → PROMISE not to accept lower proposals
             respond with any previously accepted value

Phase 2 (Accept):
  Proposer (if majority promised) → Acceptors: "ACCEPT(n, value)"
  Acceptors: if no higher proposal seen → ACCEPT it
  
  When a majority of acceptors accept the same (n, value) → value is CHOSEN.

Why Paxos is hard:
  - Single-decree Paxos: agrees on ONE value (not useful alone)
  - Multi-Paxos: runs Paxos for each slot in a log (complex)
  - Leader election: need a stable leader for efficiency
  - The paper is famously difficult to understand ("Paxos Made Simple" helped)
  - Liveness: can live-lock if two proposers keep preempting each other

Used by: Google Chubby, Google Spanner (variant), older ZooKeeper
```

### Raft (Ongaro & Ousterhout, 2014)

```
Raft was designed to be UNDERSTANDABLE (Paxos is notoriously hard).

Key insight: decompose consensus into sub-problems:
  1. Leader election
  2. Log replication
  3. Safety

State machine replication:
  All nodes maintain a LOG of commands.
  If logs are identical → state machines are identical.

Roles:
  Leader:    handles all client requests, replicates to followers
  Follower:  passive, receives log entries from leader
  Candidate: trying to become leader (election)

Leader Election:
  - Term-based (monotonic counter)
  - If follower doesn't hear from leader → times out → becomes candidate
  - Candidate requests votes from all nodes
  - Majority vote → becomes leader for that term
  - At most one leader per term (safety guarantee)
  
  Election timeout: randomized (150-300ms) to prevent split votes

Log Replication:
  Client → Leader: "Execute command X"
  Leader: append X to local log
  Leader → Followers: AppendEntries(term, prevLogIndex, entries)
  When MAJORITY acknowledge → entry is COMMITTED
  Leader responds to client

  ┌──────────────────────────────────────┐
  │ Leader Log:  [1:SET x=1] [2:SET y=2] [3:SET x=3]  ← committed
  │ Follower A:  [1:SET x=1] [2:SET y=2] [3:SET x=3]  ← replicated
  │ Follower B:  [1:SET x=1] [2:SET y=2]               ← catching up
  │ Follower C:  [1:SET x=1] [2:SET y=2] [3:SET x=3]  ← replicated
  └──────────────────────────────────────┘
  Entry 3 committed: replicated on Leader + A + C (majority of 4)

Safety:
  - Leader Completeness: elected leader must have all committed entries
  - Log Matching: if two logs have an entry with same index+term → 
    all preceding entries are identical

Used by: etcd, CockroachDB, TiKV, Consul, RethinkDB, many modern systems
```

---

## 7. Clocks and Ordering

### Lamport Timestamps

```
Problem: in a distributed system, wall clocks are UNRELIABLE.
  Clock skew: different machines show different times.
  Clock drift: clocks run at slightly different rates.
  NTP synchronization: typically ±1-10ms accuracy.

Lamport Timestamps (1978):
  Logical clock — an integer counter per node.
  
  Rules:
  1. Before any event, increment local counter
  2. When sending a message, attach the counter
  3. When receiving a message: local = max(local, received) + 1
  
  Node A: [1] → [2] → [3] ───msg(3)──────→
  Node B:              [1] → [2] → [4 = max(2,3)+1] → [5]
  
  Property: if event a HAPPENED-BEFORE event b → L(a) < L(b)
  But NOT the reverse: L(a) < L(b) does NOT mean a happened before b
  (Concurrent events can have any ordering of timestamps)
```

### Vector Clocks

```
Vector clocks track causality precisely.
Each node maintains a vector of counters (one per node).

  Nodes: A, B, C
  Vector clock = [A: x, B: y, C: z]
  
  Node A event: increment A's entry
  Send message: attach full vector
  Receive message: element-wise max, then increment own entry

  A: [1,0,0] → [2,0,0] ─msg──→
  B:                      [0,1,0] → recv → [2,2,0] → [2,3,0]
  C: [0,0,1] ──────────────────────────── still [0,0,1]

  Comparing vectors:
  V1 < V2 (happened-before): every element of V1 ≤ V2, and at least one <
  V1 || V2 (concurrent):     neither V1 < V2 nor V2 < V1

  [2,2,0] vs [0,0,1] → concurrent! (A=2>0, but C=0<1)

  Used by: Riak (for conflict detection), Amazon DynamoDB (internal)
  Problem: vector size grows with number of nodes
```

### Google TrueTime

```
Spanner's TrueTime: GPS + atomic clocks provide a BOUNDED clock uncertainty.

TrueTime.now() returns an INTERVAL: [earliest, latest]
  "The real time is somewhere within this interval."
  Typical uncertainty: ε ≈ 1-7 ms

Commit protocol:
  On commit at time s:
    Wait until TrueTime.now().earliest > s
    → GUARANTEE that s is in the past for all observers
    → This is called "commit-wait"

  This gives EXTERNAL CONSISTENCY (linearizability) WITHOUT locking reads!
  → Read-only transactions are lock-free and globally consistent.

Why only Google has this (so far):
  Requires GPS receivers + atomic clocks in every datacenter.
  CockroachDB approximates this with NTP + "uncertainty intervals" (looser bounds).
```

---

## 8. Consistent Hashing

```
Problem: distributing data across N nodes.
  Naive: hash(key) % N → BUT adding/removing a node reshuffles EVERYTHING.

Consistent Hashing (Karger, 1997):
  Imagine a ring (0 to 2^128 - 1):
  
           Node A (hash=1000)
            ╱
  ─────────●─────────────────────●── Node B (hash=5000)
  │                               │
  │         Hash Ring             │
  │                               │
  ──────●────────────────●─────────
       Node D (hash=9000)  Node C (hash=7000)
  
  Key placement: hash the key → walk clockwise → first node you hit.
  
  Adding a node: only keys between the new node and its predecessor move.
  Removing a node: only that node's keys move to its successor.
  → Only K/N keys move on average (K=total keys, N=nodes).

Virtual nodes (vnodes):
  Each physical node gets many positions on the ring.
  → Better load distribution (avoids hotspots from uneven spacing).
  → A node with 2x capacity gets 2x virtual nodes.
  
  Cassandra default: 256 vnodes per node.
  DynamoDB: uses consistent hashing for partition routing.
```

---

## 9. Replication Strategies

### Single-Leader Replication

```
One leader accepts writes → replicates to followers.

Leader ──sync/async──→ Follower 1
       ──sync/async──→ Follower 2
       ──sync/async──→ Follower 3

Synchronous: leader waits for follower ACK before confirming to client.
  ✓ Follower guaranteed up-to-date  ✗ Slow (limited by slowest follower)

Semi-synchronous: wait for 1 or 2 followers (not all).
  PostgreSQL: synchronous_standby_names = 'FIRST 1 (...)'
  MySQL: semi-sync replication

Asynchronous: leader confirms immediately, replicates in background.
  ✓ Fast  ✗ Data loss if leader crashes before replication

Problems:
  - Single point of write bottleneck
  - Failover complexity (is the new leader caught up?)
  - Replication lag → stale reads from followers
```

### Multi-Leader Replication

```
Multiple nodes accept writes → replicate to each other.

Leader A ←──→ Leader B ←──→ Leader C

Use cases:
  - Multi-datacenter (one leader per DC)
  - Offline-capable apps (each device is a "leader")
  - Collaborative editing

The HARD PROBLEM: write conflicts.
  User 1 @ Leader A: UPDATE row SET name = 'Alice'
  User 2 @ Leader B: UPDATE row SET name = 'Bob'
  → After replication: which value wins?

Conflict resolution strategies:
  1. Last-writer-wins (LWW): highest timestamp wins → data loss!
  2. Application-level resolution: app merges conflicts (complex)
  3. CRDTs: mathematically guaranteed convergence (limited data types)
  4. Operational transform: Google Docs approach
```

### Leaderless Replication

```
ALL nodes accept reads and writes (no leader).

Client ──→ Node 1 (write)
       ──→ Node 2 (write)
       ──→ Node 3 (write)

Quorum reads and writes:
  N = number of replicas
  W = number of nodes that must acknowledge a write
  R = number of nodes that must respond to a read
  
  Safety rule: W + R > N
  → At least one node in the read set has the latest write.

  Example: N=3, W=2, R=2
    Write: 2 of 3 nodes acknowledge → success
    Read:  ask 2 of 3 nodes → at least 1 has latest value
    → Use version/timestamp to pick the newest.

  Tunable consistency:
    W=1, R=1: fast but eventually consistent (DynamoDB default)
    W=N, R=1: slow writes, fast reads, strong-ish consistency
    W=1, R=N: fast writes, slow reads
    W=⌈(N+1)/2⌉, R=⌈(N+1)/2⌉: balanced quorum

  Read repair: when a read finds stale replicas → update them
  Anti-entropy: background process syncs divergent replicas (Merkle trees)

  Used by: Cassandra, Riak, DynamoDB, Voldemort
```

---

## 10. Partitioning (Sharding)

```
Splitting data across multiple nodes so each node holds a SUBSET.

Strategies:

1. Range partitioning:
   Keys A-M → Shard 1, N-Z → Shard 2
   ✓ Efficient range scans  ✗ Hotspots (if data is skewed)
   Used by: HBase, Spanner, CockroachDB

2. Hash partitioning:
   hash(key) % num_shards → shard assignment
   ✓ Even distribution  ✗ Range scans require scatter-gather
   Used by: Cassandra, DynamoDB, MongoDB (hashed shard key)

3. Compound partitioning:
   Hash on partition key → range on sort key
   ✓ Balance + range queries within a partition
   Used by: Cassandra, DynamoDB

Rebalancing:
  When adding/removing nodes, data must move.
  - Fixed partition count: pre-allocate many partitions, assign to nodes
    (Riak, Elasticsearch, Couchbase)
    Adding node = reassign some partitions (no splitting)
  - Dynamic splitting: split a partition when too large
    (HBase, MongoDB, CockroachDB)
  - Consistent hashing: virtual nodes (Cassandra)

Cross-shard operations:
  Queries spanning shards = scatter-gather (slow)
  Transactions spanning shards = 2PC or Saga (complex)
  → Design your partition key to minimize cross-shard queries!
```

---

## 11. Distributed Transactions

### Two-Phase Commit (2PC)

```
Coordinator asks all participants to prepare, then commit or abort.

Phase 1 (Prepare/Vote):
  Coordinator → all Participants: "PREPARE"
  Participant: acquire locks, write redo/undo log
  Participant → Coordinator: "YES" (ready) or "NO" (abort)

Phase 2 (Commit/Abort):
  If ALL voted YES:
    Coordinator → all Participants: "COMMIT"
  If ANY voted NO:
    Coordinator → all Participants: "ABORT"

Problems:
  - BLOCKING: if coordinator crashes after phase 1, participants are STUCK
    (they've locked resources, waiting for commit/abort that never comes)
  - High latency: 2 round trips minimum
  - Coordinator is a single point of failure

  Used by: PostgreSQL (prepared transactions), MySQL (XA), distributed TPC-C
```

### Three-Phase Commit (3PC)

```
Adds a PRE-COMMIT phase to make it non-blocking:
  Phase 1: PREPARE (same as 2PC)
  Phase 2: PRE-COMMIT (coordinator tells everyone it will commit)
  Phase 3: COMMIT

  If coordinator crashes after pre-commit:
  → Participants can safely commit (they know everyone voted YES)

  Problem: doesn't work with network partitions (can violate safety)
  → Rarely used in practice. Raft/Paxos are preferred.
```

### Saga Pattern

```
For long-running distributed transactions where 2PC is impractical:

A saga is a sequence of local transactions with compensating actions:

T1 → T2 → T3 → T4 → T5
                 ↑ fails
                 C3 ← C2 ← C1  (compensating transactions, in reverse)

Example — book a trip:
  T1: Reserve flight       C1: Cancel flight reservation
  T2: Reserve hotel        C2: Cancel hotel reservation
  T3: Charge payment       C3: Refund payment
  T4: Send confirmation

  If T3 fails: execute C2 (cancel hotel), then C1 (cancel flight).

Orchestration: central coordinator directs each step
Choreography: each service publishes events, next service reacts

Trade-offs:
  ✓ No distributed locking → better availability and performance
  ✗ No isolation: intermediate states are visible (T1 committed, T2 not yet)
  ✗ Compensating actions must be idempotent
  ✗ Complex error handling

Used by: microservices architectures, e-commerce order processing
```

---

## 12. CRDTs (Conflict-free Replicated Data Types)

```
Data structures that can be replicated and MERGED without conflicts.
Mathematically guaranteed to converge — no conflict resolution needed.

Types:
  State-based CRDTs (CvRDT): send full state, merge with lattice join
  Operation-based CRDTs (CmRDT): send operations, commutative

Common CRDTs:

G-Counter (grow-only counter):
  Each node maintains its own counter.
  Value = sum of all node counters.
  Merge = element-wise max.
  
  Node A: [A:5, B:0, C:0]  value = 5
  Node B: [A:3, B:7, C:0]  value = 10 (A hasn't replicated yet)
  Merge:  [A:5, B:7, C:0]  value = 12 ✓

PN-Counter (increment + decrement):
  Two G-Counters: one for increments, one for decrements.
  Value = sum(increments) - sum(decrements).

G-Set (grow-only set):
  Only add, never remove. Merge = union.

OR-Set (observed-remove set):
  Add and remove with unique tags.
  Each add gets a unique ID; remove deletes specific IDs.
  Used by: Riak (bucket types), Redis CRDT module.

LWW-Register (last-writer-wins register):
  Value + timestamp. Higher timestamp wins.
  Not truly conflict-free (data loss), but simple.

Used by: Riak, Redis Enterprise (active-active), Automerge, Yjs (collaborative editing)
```

---

## 13. Gossip Protocols

```
Epidemic-style information dissemination:
  Each node periodically picks a random peer and exchanges state.
  
  Round 1: Node A tells Node B about update X
  Round 2: A tells D, B tells C about X
  Round 3: all nodes know about X
  
  Convergence: O(log N) rounds for N nodes.
  Probabilistic: not guaranteed, but extremely reliable in practice.

Used for:
  - Failure detection: "I haven't heard from Node C in 3 gossip rounds"
  - Membership: Cassandra uses gossip to track cluster topology
  - Load information: share load metrics for routing decisions
  - Amazon S3: anti-entropy with Merkle trees + gossip

Cassandra gossip:
  Every second, each node gossips with 1-3 random nodes.
  Shares: status (NORMAL/LEAVING/etc.), schema version, load, tokens.
  φ-accrual failure detector: adaptive threshold instead of fixed timeout.
```

---

## 14. Split-Brain and Fencing

```
Split-brain: a network partition causes two groups of nodes to both
think they're the leader → conflicting writes → data corruption.

  ┌───────────────┐     NETWORK      ┌───────────────┐
  │ Node A (leader)│     PARTITION    │ Node B (leader?)│
  │ Node C         │  ←─── ✂ ───→   │ Node D          │
  │ Node E         │                  │ Node F          │
  └───────────────┘                  └───────────────┘
  
  Both sides elect a leader → two leaders writing to the same data!

Prevention:
  1. Quorum: need majority to elect leader (odd number of nodes)
     Left side: 3/6 = not majority → cannot elect leader
     → Only partition with majority can proceed.
     
  2. Fencing tokens: each leader gets a monotonically increasing token.
     Storage layer rejects writes with old tokens.
     
  3. STONITH ("Shoot The Other Node In The Head"):
     Power off the old leader via IPMI/BMC before promoting new one.
     
  4. Witness/Tiebreaker: an odd node that breaks ties (Cassandra lightweight nodes).
```

---

## Key Takeaways

1. **CAP is about partitions; PACELC is about daily trade-offs.** In normal operation, the real choice is latency vs consistency. Use PACELC to reason about your system.

2. **Raft is the consensus algorithm you should learn deeply.** It's used everywhere (etcd, CockroachDB, TiKV, Consul) and is designed to be understandable. Paxos is equivalent but harder to grok.

3. **Quorum math: W + R > N ensures overlap.** This is the foundation of leaderless replication. Tuning W and R lets you trade consistency for latency.

4. **The hardest problem in distributed systems is time.** Lamport clocks give you ordering without real clocks. Vector clocks give you causality. TrueTime gives you real-time bounds but requires special hardware.

5. **Consistent hashing with virtual nodes** is how almost every distributed database places data on nodes. Understanding the ring concept is fundamental.

6. **2PC blocks, Sagas compensate, CRDTs merge.** These are three fundamentally different approaches to distributed transactions. Pick based on consistency requirements.

7. **FLP impossibility** means no perfect consensus algorithm exists. Every practical system makes timing assumptions (partial synchrony) to work around it.

---

Next: [02-distributed-sql.md](02-distributed-sql.md) →
