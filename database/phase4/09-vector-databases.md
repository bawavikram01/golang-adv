# 4.9 — Vector Databases

> The AI era brought a new data type: EMBEDDINGS.  
> High-dimensional vectors representing the MEANING of text, images, audio.  
> Vector databases find "similar" items in milliseconds across billions of vectors.  
> This is the bridge between databases and machine learning.

---

## 1. What Are Embeddings?

```
An embedding is a fixed-size vector of floating-point numbers
that represents the SEMANTIC MEANING of data.

Text embedding (e.g., OpenAI text-embedding-3-small, 1536 dimensions):
  "The cat sat on the mat"  → [0.021, -0.034, 0.078, ..., 0.012]  (1536 floats)
  "A kitten rested on a rug" → [0.019, -0.031, 0.082, ..., 0.015]  (similar vector!)
  "Stock market crashed"     → [-0.071, 0.042, -0.003, ..., -0.088] (very different)

Meaningfully similar items have vectors that are CLOSE in high-dimensional space.

Embeddings exist for:
  Text:   OpenAI, Cohere, sentence-transformers, BGE
  Images: CLIP, ResNet, ViT
  Audio:  Whisper embeddings, CLAP
  Code:   CodeBERT, StarCoder embeddings
  Multi-modal: CLIP (text + image in same space)

The task: given a query vector, find the K NEAREST vectors in the database.
This is called K-Nearest Neighbors (KNN) / similarity search.
```

---

## 2. Distance Metrics

```
How to measure "closeness" between vectors:

Cosine similarity:
  cos(A, B) = (A · B) / (||A|| × ||B||)
  Range: -1 (opposite) to 1 (identical direction)
  Ignores magnitude, only compares direction.
  Best for: text embeddings (most common choice)

Euclidean distance (L2):
  d(A, B) = √(Σ (Aᵢ - Bᵢ)²)
  Range: 0 (identical) to ∞
  Considers both direction and magnitude.
  Best for: when magnitude matters (image features, spatial data)

Dot product (inner product):
  A · B = Σ (Aᵢ × Bᵢ)
  Range: -∞ to ∞
  Combines similarity and magnitude.
  Best for: recommendation systems, normalized embeddings

  If vectors are normalized (unit length):
    cosine similarity = dot product = 1 - (L2²/2)
    → All three metrics give equivalent rankings!
    → Most embedding models output normalized vectors.
```

---

## 3. ANN Algorithms — Finding Needles in Haystacks

```
Exact KNN: compare query to ALL vectors → O(N × D) per query.
  At 1 billion vectors × 1536 dimensions: IMPOSSIBLY SLOW.

Approximate Nearest Neighbors (ANN): trade accuracy for speed.
  Find ~95-99% of true nearest neighbors in milliseconds.

Major ANN algorithms:

HNSW (Hierarchical Navigable Small World):
  The most popular algorithm. Used by pgvector, Pinecone, Qdrant, Weaviate.
  
  Multi-layer graph where each node connects to nearby neighbors:
  
  Layer 3: ●─────────────────────●  (sparse, long-range connections)
  Layer 2: ●────●────────●───────●  (medium connections)
  Layer 1: ●──●──●──●──●──●──●──●  (dense, short-range connections)
  Layer 0: ●●●●●●●●●●●●●●●●●●●●●  (all points, dense graph)
  
  Search: start at top layer, greedily move toward query.
  Drop to lower layer, repeat with finer resolution.
  → O(log N) hops, each hop checks a few neighbors.
  
  Pros: fast search, high recall (accuracy), no training needed
  Cons: HIGH memory (graph in RAM), slow index build

IVF (Inverted File Index):
  Cluster vectors using k-means → N clusters (centroids).
  At query time: find nearest clusters → search only those clusters.
  
  Build: k-means on all vectors → assign each vector to nearest centroid.
  Search: find nprobe nearest centroids → search vectors in those clusters.
  
  Pros: lower memory than HNSW, fast build
  Cons: lower recall unless nprobe is high, requires training (k-means)

PQ (Product Quantization):
  Compress 1536-dim vector into ~128 bytes by:
  1. Split vector into subvectors (e.g., 1536 → 192 groups of 8)
  2. Cluster each subvector space (256 centroids per group)
  3. Store only the centroid IDs (1 byte each) → 192 bytes total
  
  Search: approximate distance using quantized representation.
  → 10-50x memory reduction with ~95% recall.
  
  Often combined: IVF-PQ = IVF clustering + PQ compression.

ScaNN (Google):
  Combines quantization with anisotropic loss function.
  State-of-the-art recall-speed trade-off.
```

---

## 4. pgvector — Vector Search in PostgreSQL

```sql
-- pgvector: PostgreSQL extension for vector similarity search
-- Install:
CREATE EXTENSION vector;

-- Table with embeddings:
CREATE TABLE documents (
    id SERIAL PRIMARY KEY,
    content TEXT,
    embedding vector(1536)    -- 1536-dimensional vector
);

-- Insert:
INSERT INTO documents (content, embedding)
VALUES ('PostgreSQL is great', '[0.021, -0.034, ...]');

-- Exact nearest neighbors (brute force):
SELECT id, content, embedding <=> '[0.019, ...]' AS distance
FROM documents
ORDER BY embedding <=> '[0.019, ...]'    -- <=> is cosine distance
LIMIT 10;

-- Distance operators:
-- <->  L2 (Euclidean) distance
-- <#>  negative inner product (for ORDER BY, since PG sorts ascending)
-- <=>  cosine distance

-- HNSW index (recommended):
CREATE INDEX ON documents
USING hnsw (embedding vector_cosine_ops)
WITH (m = 16, ef_construction = 200);
-- m: connections per node (higher = better recall, more memory)
-- ef_construction: search depth during build (higher = better index)

-- IVFFlat index (less memory, lower recall):
CREATE INDEX ON documents
USING ivfflat (embedding vector_cosine_ops)
WITH (lists = 100);   -- number of clusters
-- Query: SET ivfflat.probes = 10;  -- search 10 clusters

-- Hybrid search (vector + metadata filter):
SELECT id, content
FROM documents
WHERE category = 'technical'                  -- metadata filter
ORDER BY embedding <=> '[0.019, ...]'         -- vector similarity
LIMIT 10;
-- Pre-filter: apply WHERE first, then vector search on remaining
-- This works well when filters are selective

-- pgvector strengths:
-- ✓ Full PostgreSQL SQL (JOINs, transactions, ACID)
-- ✓ No separate infrastructure
-- ✓ Hybrid search (text + vector + metadata in one query)
-- ✓ Good for < 10 million vectors

-- pgvector limitations:
-- ✗ Slower than purpose-built vector DBs at scale
-- ✗ HNSW index must fit in memory
-- ✗ No built-in sharding for vectors
```

---

## 5. Purpose-Built Vector Databases

### Pinecone

```
Fully managed, proprietary. The "easy button" for vector search.
  - Serverless or pod-based deployment
  - Auto-scaling, no infrastructure management
  - Metadata filtering + vector search combined
  - Namespaces for data isolation
  - Sparse-dense hybrid search

index.upsert(vectors=[
    {"id": "doc1", "values": [0.1, 0.2, ...], "metadata": {"category": "tech"}},
])
results = index.query(vector=[0.1, 0.2, ...], top_k=10, filter={"category": "tech"})
```

### Milvus

```
Open-source, distributed vector database.
  - Supports billions of vectors
  - Multiple index types: HNSW, IVF-PQ, DiskANN, ScaNN
  - Hybrid search (sparse + dense vectors)
  - GPU-accelerated indexing and search
  - Cloud offering: Zilliz Cloud
  - Schema-based with typed fields
```

### Weaviate

```
Open-source vector database with built-in ML model integration.
  - Auto-vectorization: send text/images → Weaviate calls embedding model
  - GraphQL API
  - Hybrid search: BM25 + vector combined
  - Multi-tenancy built-in
  - Modules: text2vec-openai, img2vec-neural, etc.
```

### Qdrant

```
Open-source, Rust-based. Fast and memory-efficient.
  - HNSW with quantization (scalar, product, binary)
  - Efficient filtering (pre-filter optimization)
  - Payload (metadata) stored alongside vectors
  - gRPC and REST APIs
  - Distributed mode with sharding + replication
```

### Chroma

```
Open-source, lightweight. Designed for AI application prototyping.
  - Runs in-process (Python) or as a server
  - Simple API: add, query, update, delete
  - Auto-embedding with sentence-transformers
  - Good for: quick prototyping, small-medium datasets
  - Not for: production billions-scale deployment

import chromadb
collection = client.create_collection("docs")
collection.add(documents=["hello world"], ids=["1"])
results = collection.query(query_texts=["hi"], n_results=5)
```

---

## 6. Hybrid Search — Vector + Keyword

```
Pure vector search misses exact matches.
  Query: "error code ERR_CONNECTION_REFUSED"
  Vector search finds semantically similar errors
  BUT might miss the exact error code string!

Pure keyword search misses semantic meaning.
  Query: "how to fix slow database queries"
  Keyword search looks for exact words
  BUT misses documents that say "performance optimization for SQL"

Hybrid search: combine both.

Reciprocal Rank Fusion (RRF):
  1. Run keyword search → get ranking
  2. Run vector search → get ranking
  3. Combine: score = Σ 1/(k + rank_i) for each result
     k = smoothing constant (usually 60)
  
  Document appears in keyword rank 3 AND vector rank 5:
  RRF_score = 1/(60+3) + 1/(60+5) = 0.0159 + 0.0154 = 0.0313

Weighted scoring:
  final_score = α × keyword_score + (1-α) × vector_score
  Tune α per use case (0.3 keyword + 0.7 vector is common for semantic search)

Implementations:
  pgvector + pg_trgm: vector search + trigram text search in one query
  Elasticsearch 8.x: dense_vector field + text search in one query
  Weaviate: built-in hybrid mode
  Pinecone: sparse-dense vectors
```

---

## 7. RAG Architecture (Retrieval-Augmented Generation)

```
The primary use case for vector databases in 2024+:

User Question
     │
     ▼
Embed question → query vector
     │
     ▼
Vector DB: find top-K similar documents
     │
     ▼
Construct prompt: "Given these documents: [context], answer: [question]"
     │
     ▼
Send to LLM (GPT-4, Claude, etc.)
     │
     ▼
LLM generates answer grounded in retrieved documents

Chunking strategies:
  - Fixed size: split every 512 tokens (simple, loses context)
  - Semantic: split at paragraph/section boundaries
  - Recursive: try large chunks, split smaller if too big
  - Sliding window: overlapping chunks (redundant but catches boundaries)

Typical chunk size: 256-1024 tokens with 10-20% overlap

Embedding model choice matters enormously:
  Model                    Dimensions  Quality  Speed
  OpenAI text-embedding-3-large  3072   Excellent  Cloud API
  OpenAI text-embedding-3-small  1536   Very good  Cloud API
  Cohere embed-v3           1024       Very good  Cloud API
  BGE-large-en-v1.5         1024       Very good  Local
  all-MiniLM-L6-v2           384       Good       Fast local
```

---

## Key Takeaways

1. **Embeddings capture semantic meaning** as high-dimensional vectors. Similar meaning = nearby vectors. This is fundamentally different from keyword search.

2. **HNSW is the dominant ANN algorithm** — used by pgvector, Pinecone, Qdrant, Weaviate. It gives ~95-99% recall with O(log N) search time. Trade-off: high memory usage.

3. **pgvector is the pragmatic starting point.** If you're already on PostgreSQL, add vector search without a new database. It handles millions of vectors. Go purpose-built when you hit tens of millions.

4. **Hybrid search (vector + keyword) beats either alone.** Exact matches need keywords; semantic matches need vectors. Combine with RRF or weighted scoring.

5. **Chunking strategy matters more than vector DB choice** for RAG applications. Bad chunks = bad retrieval = bad LLM answers.

6. **Distance metrics are equivalent for normalized vectors.** Most embedding models normalize output, so cosine, dot product, and L2 give the same ranking.

7. **The vector DB landscape is volatile.** pgvector, Pinecone, and Qdrant are the most production-proven as of 2024. For prototyping, Chroma is the fastest to start.

---

Next: [10-streaming-and-queues.md](10-streaming-and-queues.md) →
