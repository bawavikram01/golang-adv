# 4.4 — Document Databases

> Document databases store data as semi-structured documents (JSON/BSON).  
> No rigid schema. Each document can have different fields.  
> The data model maps naturally to objects in your programming language.

---

## 1. MongoDB — The Document Database

### Data Model

```
Database → Collections → Documents (BSON)

{
  "_id": ObjectId("507f1f77bcf86cd799439011"),
  "name": "Alice",
  "email": "alice@example.com",
  "address": {                          ← embedded document
    "street": "123 Main St",
    "city": "Springfield",
    "state": "IL"
  },
  "orders": [                           ← embedded array
    { "product": "Widget", "qty": 3, "total": 29.97 },
    { "product": "Gadget", "qty": 1, "total": 49.99 }
  ],
  "tags": ["premium", "early-adopter"],
  "created_at": ISODate("2024-06-15T10:30:00Z")
}

BSON (Binary JSON):
  - Binary representation of JSON
  - Adds types: Date, ObjectId, Binary, Decimal128, Regex
  - More efficient to parse than text JSON
  - Max document size: 16 MB
```

### Schema Design — Embedding vs Referencing

```
The MOST IMPORTANT MongoDB design decision:
"Should I embed the related data or reference it?"

EMBED when:
  ✓ Data is always accessed together (1:1 or 1:few)
  ✓ Data doesn't change independently
  ✓ Bounded arrays (won't grow to millions)
  ✓ Need atomic operations on parent + children
  
  // User with embedded addresses (1:few):
  { _id: 1, name: "Alice", addresses: [{...}, {...}] }
  // One read to get user + addresses

REFERENCE when:
  ✓ Data is accessed independently
  ✓ Many-to-many relationships
  ✓ Unbounded growth (e.g., comments on a popular post)
  ✓ Data shared across multiple documents
  
  // Order references user:
  { _id: 101, user_id: 1, items: [...] }
  // Requires two queries or $lookup (join)

Anti-patterns:
  ✗ Massive arrays (>1000 elements — causes document growth, slow updates)
  ✗ Deeply nested documents (>3 levels — hard to query/update)
  ✗ Normalizing everything (you're not using SQL — denormalize!)

Schema design patterns:
  Bucket pattern:    group time-series into fixed-size buckets
  Outlier pattern:   handle exceptional documents differently
  Computed pattern:  pre-compute aggregations into documents
  Subset pattern:    embed only most recent/relevant subset
  Polymorphic:       same collection holds different "types" of documents
```

### Queries and Aggregation

```javascript
// === CRUD ===
db.users.insertOne({ name: "Alice", age: 30, tags: ["admin"] });
db.users.insertMany([{ name: "Bob" }, { name: "Carol" }]);

db.users.find({ age: { $gte: 25, $lt: 40 } });             // range
db.users.find({ tags: "admin" });                            // array contains
db.users.find({ "address.city": "Springfield" });            // nested field
db.users.find({ name: /^Al/i });                             // regex
db.users.find({ $or: [{ age: 30 }, { name: "Bob" }] });     // logical OR

db.users.updateOne(
  { _id: ObjectId("...") },
  { $set: { age: 31 }, $push: { tags: "verified" } }
);

db.users.deleteMany({ last_login: { $lt: ISODate("2023-01-01") } });

// === AGGREGATION PIPELINE (the power tool) ===
db.orders.aggregate([
  { $match: { status: "completed" } },                   // filter (like WHERE)
  { $unwind: "$items" },                                  // flatten array
  { $group: {                                             // like GROUP BY
      _id: "$items.product",
      total_revenue: { $sum: "$items.total" },
      order_count: { $sum: 1 }
  }},
  { $sort: { total_revenue: -1 } },                       // ORDER BY
  { $limit: 10 },                                         // LIMIT
  { $project: {                                           // SELECT columns
      product: "$_id",
      total_revenue: 1,
      order_count: 1,
      _id: 0
  }}
]);

// $lookup (LEFT OUTER JOIN):
db.orders.aggregate([
  { $lookup: {
      from: "users",
      localField: "user_id",
      foreignField: "_id",
      as: "user"
  }},
  { $unwind: "$user" }
]);

// Pipeline stages: $match, $group, $sort, $project, $unwind, $lookup,
// $addFields, $replaceRoot, $facet, $bucket, $graphLookup, $merge, $out
```

### Internals — WiredTiger Storage Engine

```
WiredTiger (default since MongoDB 3.2):
  - B-tree storage (not LSM — unlike many NoSQL stores)
  - Document-level concurrency control (not collection-level)
  - Snappy compression by default (also zstd, zlib)
  - MVCC with snapshot isolation
  - Checkpointing every 60 seconds
  - Journal (WAL) for durability between checkpoints

Data files:
  collection-*.wt    — collection B-tree
  index-*.wt         — index B-tree  
  journal/           — WAL files

Cache:
  wiredTiger.engineConfig.cacheSizeGB
  Default: 50% of (RAM - 1 GB), min 256 MB
  Like PostgreSQL's shared_buffers, this is the internal cache
```

### Replication and Sharding

```
Replica Set (HA):
  ┌──────────┐
  │ Primary   │──oplog──→ Secondary 1
  │ (R/W)     │──oplog──→ Secondary 2
  └──────────┘
  
  oplog: capped collection that records all changes
  Secondary replays oplog entries
  Automatic failover: if primary dies, secondaries elect a new primary
  Read preference: primary (default), primaryPreferred, secondary, nearest

Sharding (horizontal scaling):
  ┌──────────┐
  │  mongos   │  ← query router (stateless)
  └─────┬────┘
        │
  ┌─────▼───────────────────────────────┐
  │ Config servers (replica set)         │  ← shard metadata
  └─────────────────────────────────────┘
        │
  ┌─────┼─────┐
  │     │     │
  ▼     ▼     ▼
Shard1 Shard2 Shard3  ← each shard is a replica set

  Shard key: determines how documents are distributed.
  
  Shard key types:
    Hashed: hash(field) → even distribution, no range queries
    Ranged: field value ranges → range queries ok, possible hotspots
  
  Choosing shard key (critical — cannot change later in older versions!):
    ✓ High cardinality (many unique values)
    ✓ Even distribution
    ✓ Matches query patterns (queries include shard key → targeted)
    ✗ Monotonically increasing (timestamps) → all writes to last shard

  MongoDB 5.0+: resharding (change shard key — expensive but possible)

Change Streams:
  Watch for real-time changes to collections:
  db.orders.watch([{ $match: { "fullDocument.status": "shipped" } }])
  → Triggers on insert/update/delete/replace
  → Built on oplog tailing
  → Resumable with resume tokens
```

---

## 2. CouchDB — Multi-Master Replication

```
CouchDB: document database designed for replication.

Key differences from MongoDB:
  - Multi-master replication (any node accepts writes)
  - Conflict detection using revision trees (not conflict prevention)
  - HTTP/REST API (every operation is an HTTP request)
  - MapReduce views (pre-computed, incrementally updated)
  - Append-only B-tree storage (compaction to reclaim space)

Conflict handling:
  CouchDB allows conflicting writes to different replicas.
  Conflicts are STORED (both versions kept).
  Application must resolve conflicts.
  → Good for offline-first apps (mobile, edge)

  Revision tree:
    1-abc → 2-def → 3-ghi (resolved)
              ↘ 2-xyz (conflicting branch)

CouchDB is the inspiration for:
  - PouchDB (JavaScript CouchDB in the browser)
  - Couchbase (commercial successor)
```

---

## 3. Couchbase

```
Couchbase: memory-first document database with SQL-like query.

Architecture:
  - Data service: key-value operations (sub-millisecond)
  - Query service: N1QL (SQL for JSON)
  - Index service: GSI (Global Secondary Indexes)
  - Search service: full-text search (Bleve engine)
  - Analytics service: parallel analytics (columnar)
  - Eventing service: server-side JavaScript functions

N1QL (SQL for JSON):
  SELECT name, address.city
  FROM users
  WHERE ANY tag IN tags SATISFIES tag = 'admin' END
  ORDER BY name;

  -- JOINs across collections:
  SELECT u.name, o.total
  FROM users u
  JOIN orders o ON META(o).id = CONCAT('order::', o.user_id);

Key differentiator:
  - Built-in caching tier (managed cache, like Redis + MongoDB combined)
  - Sub-millisecond key-value gets (memory-first)
  - XDCR (Cross Data Center Replication): active-active multi-DC
  - Mobile sync (Couchbase Lite → Sync Gateway → Couchbase Server)
```

---

## 4. When Documents vs Relational

```
Choose documents when:
  ✓ Schema varies across records (product catalogs with different attributes)
  ✓ Hierarchical/nested data (naturally JSON-shaped)
  ✓ Rapid iteration (schema evolves frequently)
  ✓ Read-heavy access patterns reading entire documents
  ✓ Content management, user profiles, event logging

Choose relational when:
  ✓ Complex relationships (many-to-many, JOINs across entities)
  ✓ Strong consistency requirements (ACID across multiple tables)
  ✓ Complex queries with aggregation across entities
  ✓ Strict schema enforcement matters
  ✓ Financial data, inventory, booking systems

The pragmatic truth:
  PostgreSQL + JSONB gives you 80% of document database benefits
  MongoDB + $lookup gives you 60% of relational join benefits
  → The lines are blurring. Choose based on PRIMARY access pattern.
```

---

## Key Takeaways

1. **Embedding vs referencing** is the fundamental MongoDB design decision. Embed for 1:few, reference for 1:many or many-to-many.
2. **Aggregation pipeline** is MongoDB's power tool — it can express most analytical queries through staged transformations.
3. **Shard key selection is critical** — it determines data distribution AND query routing. Design for your access patterns.
4. **WiredTiger uses B-trees** (not LSM). Document-level locking provides good concurrency.
5. **Change streams** enable event-driven architectures on MongoDB (similar purpose to PostgreSQL's logical replication / LISTEN/NOTIFY).
6. **16 MB document limit** means you MUST design around bounded embedding. Unbounded arrays are an anti-pattern.
7. **CouchDB's conflict model** is brilliant for offline-first/edge computing — PouchDB in browsers syncing to CouchDB.
8. **PostgreSQL JSONB vs MongoDB**: for moderate document needs, JSONB avoids running a second database. For document-first workloads at scale, MongoDB is purpose-built.

---

Next: [05-wide-column-stores.md](05-wide-column-stores.md) →
