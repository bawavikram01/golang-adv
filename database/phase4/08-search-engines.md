# 4.8 — Search Engines (as Databases)

> Full-text search, log analytics, e-commerce product search.  
> Search engines use inverted indexes — the opposite of traditional databases.  
> Instead of "for this row, what are its words?" they answer  
> "for this word, which rows contain it?"

---

## 1. Inverted Index — The Core Data Structure

```
Documents:
  Doc 1: "The quick brown fox"
  Doc 2: "The quick brown dog"
  Doc 3: "The lazy brown fox"

Inverted index:
  Term     → Posting List (document IDs + positions)
  ─────────────────────────────────────────────────
  "the"    → [1:0, 2:0, 3:0]        (doc:position)
  "quick"  → [1:1, 2:1]
  "brown"  → [1:2, 2:2, 3:2]
  "fox"    → [1:3, 3:3]
  "dog"    → [2:3]
  "lazy"   → [3:1]

Query "quick fox":
  "quick" → docs [1, 2]
  "fox"   → docs [1, 3]
  Intersection → doc [1] ✓

This is why search is fast:
  Instead of scanning every document for words,
  jump directly to the list of documents containing each word.
  
  Boolean operations on posting lists:
    AND → intersection of posting lists
    OR  → union of posting lists
    NOT → difference of posting lists
```

---

## 2. Text Analysis Pipeline

```
Raw text → Analyzer → Index terms

Analyzer = Character filters → Tokenizer → Token filters

Example: "The Quick BROWN fox's"

1. Character filter: HTML strip, pattern replace
   → "The Quick BROWN fox's"

2. Tokenizer: split into tokens
   Standard tokenizer → ["The", "Quick", "BROWN", "fox's"]

3. Token filters (applied in order):
   Lowercase → ["the", "quick", "brown", "fox's"]
   Possessive stemmer → ["the", "quick", "brown", "fox"]
   Stop words removal → ["quick", "brown", "fox"]
   Stemming (Porter) → ["quick", "brown", "fox"]

These are the terms stored in the inverted index.

Search query goes through the SAME analysis pipeline:
  User types "FOXES" → analyze → "fox" → matches index term "fox"

Different analyzers for different use cases:
  Standard:  general text
  Keyword:   exact match (no tokenization)
  Whitespace: split on whitespace only
  Language:  language-specific stemming and stop words
  ICU:       Unicode-aware, CJK language support
  N-gram:    partial matching ("data" → "da", "dat", "ata", "data")
  Edge n-gram: autocomplete ("data" → "d", "da", "dat", "data")
```

---

## 3. Elasticsearch

### Architecture

```
Cluster → Nodes → Indices → Shards → Segments

┌────────────────────────────────────────────────┐
│              Elasticsearch Cluster               │
│                                                  │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐      │
│  │ Node 1    │  │ Node 2    │  │ Node 3    │      │
│  │ (master)  │  │ (data)    │  │ (data)    │      │
│  │           │  │           │  │           │      │
│  │ Shard 0P  │  │ Shard 1P  │  │ Shard 2P  │      │
│  │ Shard 1R  │  │ Shard 2R  │  │ Shard 0R  │      │
│  └──────────┘  └──────────┘  └──────────┘      │
└────────────────────────────────────────────────┘

P = primary shard, R = replica shard
Each shard = independent Lucene index
Each Lucene index = multiple segments (immutable)

Segments:
  - Write: new documents go to an in-memory buffer
  - Refresh (every 1 second): buffer → new segment (searchable)
  - Merge: background process merges small segments into larger ones
  - Segments are IMMUTABLE (like SSTables)
    → Deletes = mark as deleted, removed during merge
    → Updates = delete old + insert new

Node roles:
  Master: cluster state management, shard allocation
  Data: stores shards, handles queries
  Ingest: pre-processing pipeline (transforms before indexing)
  Coordinating: routes requests, aggregates results (any node can do this)
```

### Query DSL

```json
// Full-text search:
GET /products/_search
{
  "query": {
    "match": {
      "description": "comfortable running shoes"
    }
  }
}
// Analyzes query → finds documents matching any term
// Ranked by TF-IDF / BM25 relevance score

// Phrase match (terms must be adjacent, in order):
{
  "query": {
    "match_phrase": {
      "description": "running shoes"
    }
  }
}

// Boolean queries (combine conditions):
{
  "query": {
    "bool": {
      "must": [
        { "match": { "description": "running shoes" } }
      ],
      "filter": [
        { "range": { "price": { "gte": 50, "lte": 200 } } },
        { "term": { "brand": "nike" } }
      ],
      "should": [
        { "match": { "description": "lightweight" } }  // boosts score
      ],
      "must_not": [
        { "term": { "status": "discontinued" } }
      ]
    }
  }
}
// must: required, contributes to score
// filter: required, does NOT contribute to score (cacheable!)
// should: optional, boosts score if matched
// must_not: excluded

// Fuzzy search (handles typos):
{
  "query": {
    "fuzzy": {
      "name": {
        "value": "quikc",
        "fuzziness": "AUTO"    // edit distance 1-2 based on term length
      }
    }
  }
}

// Aggregations (analytics):
{
  "size": 0,  // don't return documents, just aggregations
  "aggs": {
    "avg_price": { "avg": { "field": "price" } },
    "brands": {
      "terms": { "field": "brand.keyword", "size": 10 },
      "aggs": {
        "avg_rating": { "avg": { "field": "rating" } }
      }
    },
    "price_ranges": {
      "range": {
        "field": "price",
        "ranges": [
          { "to": 50 },
          { "from": 50, "to": 100 },
          { "from": 100 }
        ]
      }
    }
  }
}

// Autocomplete (using edge n-grams or completion suggester):
{
  "suggest": {
    "product-suggest": {
      "prefix": "lap",
      "completion": {
        "field": "suggest",
        "fuzzy": { "fuzziness": 1 }
      }
    }
  }
}
```

### Mappings (Schema)

```json
// Elasticsearch mappings = schema definition
PUT /products
{
  "mappings": {
    "properties": {
      "name": {
        "type": "text",             // full-text search (analyzed)
        "fields": {
          "keyword": { "type": "keyword" }  // exact match + aggregations
        }
      },
      "description": { "type": "text", "analyzer": "english" },
      "price": { "type": "float" },
      "brand": { "type": "keyword" },    // exact match only
      "created_at": { "type": "date" },
      "location": { "type": "geo_point" },
      "tags": { "type": "keyword" }       // array of keywords
    }
  }
}

// text vs keyword:
// text:    analyzed (tokenized, lowercased, stemmed) — for search
// keyword: not analyzed — for filtering, sorting, aggregations
// Common pattern: multi-field (text + keyword sub-field)
```

---

## 4. Lucene Internals

```
Apache Lucene: the search library underneath Elasticsearch and Solr.

Index structure:
  Segment = self-contained mini-index containing:
    - Inverted index (term → posting list)
    - Stored fields (original document values)
    - Doc values (column-oriented, for sorting/aggregation)
    - Norms (field length normalization for scoring)
    - Term vectors (per-document term positions)
    - Points (numeric/geo BKD tree index)

Segment lifecycle:
  1. Documents added to in-memory buffer
  2. Buffer flushed to new segment (immutable on disk)
  3. Multiple segments searched in parallel (union results)
  4. Background merging: small segments → larger segments
     → Reclaims deleted documents during merge
     → Merging is expensive I/O (like compaction in LSM trees)

Scoring — BM25 (default since Lucene 6):
  score(q, d) = Σ IDF(t) × (tf(t,d) × (k1 + 1)) / (tf(t,d) + k1 × (1 - b + b × |d|/avgdl))
  
  tf(t,d):  term frequency (how often term t appears in document d)
  IDF(t):   inverse document frequency (rarer terms score higher)
  |d|:      document length
  avgdl:    average document length
  k1, b:    tuning parameters (default k1=1.2, b=0.75)
  
  BM25 replaced TF-IDF because it handles:
  - Term frequency saturation (diminishing returns for repeated terms)
  - Document length normalization (short docs aren't unfairly penalized)
```

---

## 5. OpenSearch vs Elasticsearch

```
OpenSearch: Amazon's fork of Elasticsearch (since 2021).

Why the fork: Elastic changed license from Apache 2.0 to SSPL.
  Amazon forked Elasticsearch 7.10 as OpenSearch (Apache 2.0).

Differences (2024+):
  - OpenSearch: Apache 2.0 license (truly open source)
  - Elasticsearch: SSPL / Elastic License (cloud restrictions)
  - Both have diverged in features since the fork
  - OpenSearch: built-in observability, security plugin, ML
  - Elasticsearch: newer features (ESQL, universal profiling)
  - API compatibility: largely compatible, diverging over time

Choose OpenSearch: if you need Apache 2.0 license, AWS-native
Choose Elasticsearch: if you want Elastic's latest features, Elastic Cloud
```

---

## 6. Lightweight Search: Meilisearch, Typesense

```
Meilisearch:
  - Search-as-a-service oriented (simple API, fast setup)
  - Typo tolerance, faceting, filtering out of the box
  - Sub-50ms search latency
  - Not for log analytics — designed for product search, site search
  - Written in Rust, very efficient

Typesense:
  - Similar to Meilisearch, alternative to Algolia (SaaS search)
  - In-memory index (fast but needs RAM)
  - Built-in geosearch, multi-sort, curation
  - C++, open source

When to use these over Elasticsearch:
  ✓ Product search, site search (simple use case)
  ✓ Fast time-to-value (minutes to set up, not days)
  ✓ Don't need log analytics or complex aggregations
  ✗ Not for petabyte-scale data
  ✗ Not for log aggregation (use Elasticsearch/OpenSearch)
```

---

## 7. PostgreSQL Full-Text Search vs Elasticsearch

```
PostgreSQL tsvector/tsquery:
  ✓ Good enough for many applications
  ✓ No separate infrastructure
  ✓ ACID consistent with your data
  ✓ Full SQL joins with search results
  ✗ Limited relevance tuning
  ✗ No distributed search
  ✗ Slower for complex full-text queries at scale

Elasticsearch:
  ✓ Purpose-built for search (faster, more features)
  ✓ Distributed (scales to petabytes)
  ✓ Rich query DSL, aggregations, fuzzy search
  ✓ Near real-time (1-second refresh)
  ✗ Separate infrastructure to maintain
  ✗ Eventually consistent (not ACID)
  ✗ Syncing data from primary DB adds complexity

Rule of thumb:
  < 1 million searchable documents → PostgreSQL FTS is fine
  > 1 million, or complex search needs → consider Elasticsearch
  > 100 million, or log analytics → Elasticsearch/OpenSearch
```

---

## Key Takeaways

1. **Inverted index** is the foundational data structure. Terms → posting lists. Boolean operations on posting lists enable complex queries.

2. **Text analysis pipeline** determines search quality. The same analyzer must be applied at index time AND query time. Misconfigured analyzers = broken search.

3. **BM25 scoring** replaced TF-IDF. It handles term frequency saturation and document length normalization. Understanding scoring helps you tune relevance.

4. **Elasticsearch segments are immutable** (like LSM SSTables). Deletes mark, merges clean up. Heavy updates = segment churn = performance degradation.

5. **filter context vs query context**: filters are exact (yes/no, cacheable), queries contribute to score. Put non-scoring conditions in `filter` for performance.

6. **PostgreSQL FTS handles surprisingly much.** Don't add Elasticsearch unless you've outgrown pg_trgm + tsvector. The operational cost of a second system is real.

7. **Meilisearch/Typesense** are excellent for product search — sub-50ms, typo-tolerant, easy to set up. Don't use Elasticsearch for simple site search.

---

Next: [09-vector-databases.md](09-vector-databases.md) →
