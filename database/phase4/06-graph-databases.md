# 4.6 — Graph Databases

> Some data is fundamentally about RELATIONSHIPS.  
> Social networks, fraud rings, knowledge graphs, recommendation engines.  
> When your queries are "find all paths" or "who is connected to whom,"  
> relational JOINs become exponentially slow. Graph databases thrive here.

---

## 1. Graph Models

### Property Graph Model

```
The dominant model (Neo4j, Amazon Neptune, ArangoDB, JanusGraph).

Nodes (vertices):
  - Have labels (types): (:Person), (:Company), (:Product)
  - Have properties: {name: "Alice", age: 30}

Edges (relationships):
  - Have a type: [:WORKS_AT], [:KNOWS], [:PURCHASED]
  - Have direction: (Alice)-[:KNOWS]->(Bob)
  - Have properties: {since: 2020, strength: 0.8}

  (Alice:Person) ──KNOWS{since:2020}──→ (Bob:Person)
        │                                     │
   WORKS_AT{role:"Engineer"}           WORKS_AT{role:"Manager"}
        │                                     │
        ▼                                     ▼
  (Acme:Company) ──PARTNER_OF──→ (Globex:Company)
```

### RDF (Resource Description Framework)

```
W3C standard — data as triples: (subject, predicate, object)

<http://example.org/Alice> <http://xmlns.com/foaf/0.1/knows> <http://example.org/Bob> .
<http://example.org/Alice> <http://xmlns.com/foaf/0.1/name> "Alice" .
<http://example.org/Alice> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://xmlns.com/foaf/0.1/Person> .

Queried with SPARQL:
  SELECT ?name WHERE {
    ?person foaf:knows <http://example.org/Bob> .
    ?person foaf:name ?name .
  }

RDF is used for:
  - Semantic web, linked data
  - Knowledge graphs (Wikidata, DBpedia)
  - Government/scientific data integration

Property graph is used for:
  - Application development
  - Real-time traversals
  - Most commercial use cases
```

---

## 2. Neo4j — The Leading Graph Database

### Cypher Query Language

```cypher
// Create nodes:
CREATE (alice:Person {name: 'Alice', age: 30})
CREATE (bob:Person {name: 'Bob', age: 35})
CREATE (acme:Company {name: 'Acme Corp'})

// Create relationships:
CREATE (alice)-[:KNOWS {since: 2020}]->(bob)
CREATE (alice)-[:WORKS_AT {role: 'Engineer'}]->(acme)
CREATE (bob)-[:WORKS_AT {role: 'Manager'}]->(acme)

// Pattern matching (the core of Cypher):
// Find Alice's friends:
MATCH (alice:Person {name: 'Alice'})-[:KNOWS]->(friend)
RETURN friend.name

// Friends of friends (2 hops):
MATCH (alice:Person {name: 'Alice'})-[:KNOWS*2]->(fof)
RETURN DISTINCT fof.name

// Variable-length paths (1 to 5 hops):
MATCH path = (alice:Person {name: 'Alice'})-[:KNOWS*1..5]->(target)
RETURN target.name, length(path) AS distance

// Shortest path:
MATCH path = shortestPath(
  (alice:Person {name: 'Alice'})-[:KNOWS*..10]-(bob:Person {name: 'Bob'})
)
RETURN path

// Aggregation:
MATCH (p:Person)-[:WORKS_AT]->(c:Company)
RETURN c.name, count(p) AS employee_count
ORDER BY employee_count DESC

// Subqueries and OPTIONAL MATCH:
MATCH (p:Person)
OPTIONAL MATCH (p)-[:PURCHASED]->(product)
RETURN p.name, collect(product.name) AS purchases

// UNWIND (expand a list into rows):
UNWIND ['Alice', 'Bob', 'Carol'] AS name
MATCH (p:Person {name: name})
RETURN p

// MERGE (create if not exists):
MERGE (p:Person {email: 'alice@example.com'})
ON CREATE SET p.name = 'Alice', p.created = datetime()
ON MATCH SET p.last_seen = datetime()
```

### Graph Algorithms

```cypher
// Neo4j Graph Data Science library (GDS):

// PageRank (find influential nodes):
CALL gds.pageRank.stream('myGraph')
YIELD nodeId, score
RETURN gds.util.asNode(nodeId).name AS name, score
ORDER BY score DESC LIMIT 10

// Community detection (Louvain):
CALL gds.louvain.stream('myGraph')
YIELD nodeId, communityId
RETURN communityId, collect(gds.util.asNode(nodeId).name) AS members

// Shortest path (Dijkstra):
CALL gds.shortestPath.dijkstra.stream('myGraph', {
  sourceNode: startNode,
  targetNode: endNode,
  relationshipWeightProperty: 'cost'
})
YIELD path, totalCost

// Betweenness centrality (find bridge nodes):
CALL gds.betweenness.stream('myGraph')
YIELD nodeId, score

// Node similarity (Jaccard):
CALL gds.nodeSimilarity.stream('myGraph')
YIELD node1, node2, similarity
```

### Neo4j Architecture

```
Storage:
  Native graph storage — nodes and relationships are stored as
  linked records with direct pointers to adjacent nodes.
  
  Node store:  fixed-size records with pointers to:
    → First relationship
    → First property
    → Labels
  
  Relationship store:  doubly-linked list per node
    → Start node, end node
    → Next relationship for start node
    → Next relationship for end node
    → Type, properties
  
  This means: traversing a relationship is a POINTER FOLLOW, not a JOIN.
  Relational JOIN: index lookup → O(log N) per hop
  Neo4j traversal: pointer dereference → O(1) per hop (index-free adjacency)
  
  At 10 hops: relational = 10 × O(log N) index lookups
              graph = 10 × O(1) pointer follows
  → Graph databases win DRASTICALLY for deep traversals.

Limitations:
  - Not designed for heavy analytics/aggregations (use a columnar DB)
  - Sharding is hard for graph workloads (cutting edges is expensive)
  - Memory-hungry (best when graph fits in RAM)
```

---

## 3. Other Graph Databases

### Amazon Neptune

```
Managed graph database supporting BOTH property graph AND RDF.

  Property graph: openCypher + Gremlin (Apache TinkerPop)
  RDF: SPARQL
  
  Storage: distributed, replicated across 3 AZs
  Read replicas: up to 15
  
  Good for: AWS-native teams, dual-model (property graph + RDF) needs
```

### ArangoDB (Multi-Model)

```
ArangoDB: document + graph + key-value in ONE database.

  AQL (ArangoDB Query Language):
    FOR v, e, p IN 1..3 OUTBOUND 'users/alice' GRAPH 'social'
      RETURN {vertex: v.name, edge: e.type, path_length: LENGTH(p.edges)}
  
  Advantage: no need to copy data between a doc store and a graph store.
  The same document collection can be traversed as a graph.
```

### Apache TinkerPop / Gremlin

```
TinkerPop: standard graph computing framework.
Gremlin: the traversal language (works with Neo4j, Neptune, JanusGraph, etc.).

g.V().has('name', 'Alice')           // start at Alice
  .out('KNOWS')                       // traverse KNOWS edges outward
  .out('KNOWS')                       // friends of friends
  .values('name')                     // return their names
  .dedup()                            // unique results

g.V().has('name', 'Alice')
  .repeat(out('KNOWS')).times(3)      // 3 hops
  .path()                             // return full path
  .by('name')                         // use name property

Gremlin is more procedural (step-by-step traversal).
Cypher is more declarative (pattern matching).
```

---

## 4. Graph Use Cases

```
Social networks:
  "Find friends of friends who also like hiking"
  (user)-[:FRIENDS]->()-[:FRIENDS]->(fof)-[:LIKES]->(hiking)

Fraud detection:
  "Find circular money transfers (indicating money laundering)"
  MATCH path = (a)-[:TRANSFERRED*3..6]->(a)  -- cycle back to sender
  WHERE ALL(r IN relationships(path) WHERE r.amount > 10000)
  RETURN path

Recommendation engines:
  "People who bought X also bought..."
  MATCH (user)-[:PURCHASED]->(product)<-[:PURCHASED]-(other)-[:PURCHASED]->(rec)
  WHERE NOT (user)-[:PURCHASED]->(rec)
  RETURN rec.name, count(*) AS score ORDER BY score DESC

Knowledge graphs:
  "What drugs treat diseases that affect the BRCA1 gene?"
  MATCH (gene:Gene {name:'BRCA1'})-[:ASSOCIATED_WITH]->(disease)
        -[:TREATED_BY]->(drug)
  RETURN drug.name, disease.name

Network/infrastructure:
  "What servers are affected if this switch goes down?"
  MATCH (switch:Switch {id: 'sw-42'})<-[:CONNECTED_TO*]-(server:Server)
  RETURN server

Access control:
  "Does user X have permission to resource Y through any group membership?"
  MATCH path = (user:User {id:'X'})-[:MEMBER_OF*]->(group)
        -[:HAS_PERMISSION]->(resource {id:'Y'})
  RETURN path IS NOT NULL AS has_access
```

---

## 5. When Graph vs Relational

```
Use a graph database when:
  ✓ Queries involve variable-length paths (1..N hops)
  ✓ Relationships are as important as entities
  ✓ Schema is highly connected and evolving
  ✓ Deep traversals (>3 JOINs) are common
  ✓ Pattern matching across relationships

Stick with relational when:
  ✓ Fixed-depth JOINs (1-2 levels)
  ✓ Aggregation-heavy workloads (SUM, AVG, GROUP BY)
  ✓ Strict schema with evolving constraints
  ✓ ACID transactions across many entities
  ✓ Bulk data processing

PostgreSQL recursive CTEs can handle moderate graph queries:
  WITH RECURSIVE friends AS (
    SELECT friend_id FROM friendships WHERE user_id = 1
    UNION
    SELECT f.friend_id FROM friendships f
    JOIN friends ON f.user_id = friends.friend_id
  )
  SELECT * FROM friends;
  → Works for < millions of edges, < 5-6 hops depth
  → Beyond that: dedicated graph database wins decisively
```

---

## Key Takeaways

1. **Index-free adjacency** is the key advantage of native graph storage. Traversing a relationship is O(1) pointer follow, not O(log N) index lookup.
2. **Cypher (Neo4j)** is declarative pattern matching. **Gremlin (TinkerPop)** is procedural traversal. Both are expressive; Cypher is easier to learn.
3. **Graph databases shine at depth.** At 1-2 hops, PostgreSQL is fine. At 5+ hops with millions of relationships, graph databases are orders of magnitude faster.
4. **Fraud detection, social networks, recommendations, knowledge graphs** are killer use cases. If your problem is "find patterns in connections," think graph.
5. **Graph sharding is fundamentally hard** — you can't split a highly connected graph without cutting edges, causing cross-shard traversals.
6. **ArangoDB's multi-model** approach is pragmatic — store documents AND traverse them as graphs without data duplication.

---

Next: [07-time-series-databases.md](07-time-series-databases.md) →
