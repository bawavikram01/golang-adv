# 4.7 — Time-Series Databases

> Metrics, IoT sensors, financial ticks, logs — data stamped with time.  
> Time-series workloads are dominated by writes that never update,  
> queries that aggregate over time ranges, and data that expires.  
> General-purpose databases struggle. Purpose-built ones dominate.

---

## 1. What Makes Time-Series Special

```
Time-series data characteristics:
  1. Write-heavy:     append-only, massive ingest rates (millions/sec)
  2. Time-ordered:    data arrives roughly in order
  3. Rarely updated:  sensor readings don't change after writing
  4. Range-queried:   "show me CPU usage for the last 24 hours"
  5. Downsampled:     old data is aggregated (5-min avg → 1-hour avg → 1-day avg)
  6. TTL:             data expires (keep 30 days raw, 1 year downsampled)
  7. High cardinality: millions of unique series (one per device × metric)

Why general-purpose DBs struggle:
  PostgreSQL: B-tree index on timestamp → write amplification, bloat
  MongoDB: document overhead per data point, index maintenance
  
  Time-series DBs optimize for:
  - Columnar compression (adjacent timestamps compress well)
  - Time-partitioned storage (drop old data = drop a file)
  - Specialized aggregation (rollups, continuous aggregates)
  - Append-only write path (no MVCC overhead)
```

---

## 2. TimescaleDB — PostgreSQL + Time-Series

```sql
-- TimescaleDB: PostgreSQL extension, not a separate database.
-- Full SQL, joins, transactions — PLUS time-series optimizations.

-- Install:
CREATE EXTENSION timescaledb;

-- Convert a regular table to a hypertable:
CREATE TABLE metrics (
    time        TIMESTAMPTZ NOT NULL,
    device_id   TEXT NOT NULL,
    cpu         DOUBLE PRECISION,
    memory      DOUBLE PRECISION,
    disk_io     DOUBLE PRECISION
);

SELECT create_hypertable('metrics', 'time');
-- Now 'metrics' is transparently partitioned by time (default: 7 days per chunk)

-- Optionally partition by device_id too:
SELECT create_hypertable('metrics', 'time',
    partitioning_column => 'device_id',
    number_partitions => 4);

-- Insert (same SQL, but optimized):
INSERT INTO metrics VALUES (NOW(), 'server-1', 72.5, 85.3, 120.0);

-- Time-series queries:
SELECT time_bucket('1 hour', time) AS hour,
       device_id,
       AVG(cpu) AS avg_cpu,
       MAX(cpu) AS max_cpu,
       MIN(cpu) AS min_cpu
FROM metrics
WHERE time > NOW() - INTERVAL '24 hours'
GROUP BY hour, device_id
ORDER BY hour DESC;

-- Continuous aggregates (materialized, auto-refreshing):
CREATE MATERIALIZED VIEW hourly_metrics
WITH (timescaledb.continuous) AS
SELECT time_bucket('1 hour', time) AS hour,
       device_id,
       AVG(cpu) AS avg_cpu,
       MAX(memory) AS max_memory
FROM metrics
GROUP BY hour, device_id;

-- Auto-refresh policy:
SELECT add_continuous_aggregate_policy('hourly_metrics',
    start_offset => INTERVAL '3 hours',
    end_offset => INTERVAL '1 hour',
    schedule_interval => INTERVAL '1 hour');

-- Data retention (automatically drop old data):
SELECT add_retention_policy('metrics', INTERVAL '30 days');
-- Drops entire chunks (not individual rows) → instant, no vacuum needed

-- Compression (columnar, huge space savings):
ALTER TABLE metrics SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'device_id',
    timescaledb.compress_orderby = 'time DESC'
);
SELECT add_compression_policy('metrics', INTERVAL '7 days');
-- Compressed chunks: 90-95% space reduction, faster range scans

-- Still full PostgreSQL:
SELECT m.*, d.location
FROM metrics m
JOIN devices d ON m.device_id = d.id
WHERE m.time > NOW() - INTERVAL '1 hour';
-- This is impossible in most dedicated time-series DBs!
```

---

## 3. InfluxDB

```
InfluxDB: purpose-built time-series database.

Data model (line protocol):
  measurement,tag_key=tag_value field_key=field_value timestamp
  
  cpu,host=server01,region=us-east usage_idle=72.5,usage_user=23.1 1435362189575692182
  
  Measurement: like a table name ("cpu")
  Tags: indexed metadata (host, region) — for filtering/grouping
  Fields: the actual values (usage_idle) — NOT indexed
  Timestamp: nanosecond precision

Flux query language (InfluxDB 2.x):
  from(bucket: "metrics")
  |> range(start: -24h)
  |> filter(fn: (r) => r._measurement == "cpu" and r.host == "server01")
  |> aggregateWindow(every: 1h, fn: mean)
  |> yield()

InfluxQL (SQL-like, InfluxDB 1.x):
  SELECT MEAN("usage_idle") FROM "cpu"
  WHERE "host" = 'server01' AND time > now() - 24h
  GROUP BY time(1h)

Architecture (InfluxDB 3.x / IOx):
  - Rewritten in Rust (was Go)
  - Apache Arrow + Parquet (columnar)
  - DataFusion query engine
  - Object storage (S3) for persistence
  
Retention policies:
  CREATE RETENTION POLICY "one_month" ON "mydb"
    DURATION 30d REPLICATION 1 DEFAULT;

InfluxDB vs TimescaleDB:
  InfluxDB: simpler for pure metrics, Flux is powerful for transforms
  TimescaleDB: full SQL + JOINs, runs on PostgreSQL (reuse existing skills)
```

---

## 4. Prometheus — Pull-Based Monitoring

```
Prometheus: metrics collection + time-series DB + alerting.
NOT a general-purpose TSDB — designed for monitoring and alerting.

Architecture:
  Prometheus server ──scrape──→ Application /metrics endpoint
       │                        (every 15-30 seconds)
       ├── TSDB (local storage)
       ├── PromQL (query engine)
       └── Alertmanager (notifications)

Data model:
  metric_name{label1="value1", label2="value2"} float64_value timestamp
  
  http_requests_total{method="GET", handler="/api/users", status="200"} 1234
  
  Metric types:
    Counter:   monotonically increasing (requests, errors)
    Gauge:     up and down (temperature, memory usage)
    Histogram: bucketed distribution (request latency)
    Summary:   quantiles (p50, p99)

PromQL:
  # Rate of requests per second over last 5 minutes:
  rate(http_requests_total{job="api"}[5m])
  
  # 99th percentile latency:
  histogram_quantile(0.99, rate(http_request_duration_seconds_bucket[5m]))
  
  # Top 5 services by error rate:
  topk(5, sum by (service)(rate(http_requests_total{status=~"5.."}[5m])))

Limitations:
  - Local storage only (not distributed)
  - Not for long-term storage (30-day default retention)
  - For long-term: Thanos, Cortex, Mimir (add distributed storage on top)
  - Pull model doesn't work for batch/short-lived jobs (use pushgateway)
  - Float64 values only (no strings, no complex types)
```

---

## 5. ClickHouse — The Analytics Powerhouse

```
ClickHouse: columnar OLAP database. Not purely a TSDB,
but EXCELLENT for time-series analytics.

Why ClickHouse is fast:
  1. Columnar storage: only reads columns needed by query
  2. Compression: ~10x on columnar data (LZ4, ZSTD, delta, gorilla)
  3. Vectorized execution: processes columns in batches, SIMD
  4. Sparse indexing: primary key is a sparse index (not B-tree)
  5. Parallel processing: all cores for every query

CREATE TABLE metrics (
    timestamp DateTime,
    device_id String,
    metric_name String,
    value Float64
) ENGINE = MergeTree()
ORDER BY (device_id, metric_name, timestamp)
PARTITION BY toYYYYMM(timestamp)
TTL timestamp + INTERVAL 90 DAY;

-- ORDER BY is the sparse primary key (not unique, not a constraint)
-- PARTITION BY enables efficient data retention (drop partition)
-- TTL auto-deletes expired data

-- Queries are FAST:
SELECT device_id,
       toStartOfHour(timestamp) AS hour,
       avg(value) AS avg_value,
       quantile(0.99)(value) AS p99
FROM metrics
WHERE metric_name = 'cpu_usage'
  AND timestamp > now() - INTERVAL 7 DAY
GROUP BY device_id, hour
ORDER BY hour;
-- Scans billions of rows per second on modest hardware

-- Materialized views (real-time aggregation):
CREATE MATERIALIZED VIEW hourly_avg
ENGINE = AggregatingMergeTree()
ORDER BY (device_id, metric_name, hour)
AS SELECT
    device_id, metric_name,
    toStartOfHour(timestamp) AS hour,
    avgState(value) AS avg_value
FROM metrics
GROUP BY device_id, metric_name, hour;

ClickHouse vs TimescaleDB:
  ClickHouse: faster for pure analytics, no JOINs complexity, massive scale
  TimescaleDB: full SQL, ACID, JOINs, PostgreSQL ecosystem
  
ClickHouse vs InfluxDB:
  ClickHouse: more flexible data model, SQL, better for ad-hoc analytics
  InfluxDB: simpler for metrics collection, better write API
```

---

## 6. QuestDB

```
QuestDB: high-performance time-series DB written in zero-GC Java.

Key innovations:
  - Columnar storage with memory-mapped files
  - SIMD-accelerated SQL execution
  - Ingestion: 1.4M rows/sec per core (via ILP protocol)
  - PostgreSQL wire protocol (connect with any PG client)
  - InfluxDB line protocol support (drop-in replacement)

SELECT timestamp, avg(temperature), max(temperature)
FROM sensors
WHERE device = 'thermostat-1'
SAMPLE BY 1h                    -- QuestDB's time-bucket syntax
ALIGN TO CALENDAR;

Good for: IoT, financial tick data, real-time dashboards
```

---

## 7. Data Retention and Downsampling

```
Raw data retention strategies:

Tiered retention:
  Raw data (1-second granularity):   keep 7 days
  5-minute aggregates:               keep 30 days
  1-hour aggregates:                 keep 1 year
  1-day aggregates:                  keep forever

Implementation:
  TimescaleDB: continuous aggregates + retention policies
  InfluxDB: continuous queries + retention policies
  Prometheus: recording rules + Thanos long-term storage
  ClickHouse: materialized views + TTL per partition

Drop vs delete:
  Time-partitioned storage lets you DROP old partitions.
  Dropping a partition = deleting a file = instant, O(1).
  Deleting individual rows = scan + mark + compact = slow, O(N).
  → Always partition by time for time-series data!
```

---

## Key Takeaways

1. **TimescaleDB is the safe choice** — full PostgreSQL + time-series optimizations. If you already know PostgreSQL, you already know TimescaleDB.

2. **Continuous aggregates** are the killer feature of time-series DBs. Pre-compute hourly/daily rollups automatically as data arrives.

3. **Data retention = drop partitions, not delete rows.** Time-based partitioning makes expiring old data instant.

4. **ClickHouse is insanely fast for analytics** — billions of rows per second. Use it when you need analytical queries on time-series, not just simple aggregation.

5. **Prometheus is for monitoring, not general TSDB.** It's pull-based, local-only, and float64-only. Pair with Thanos/Mimir for long-term storage.

6. **Compression ratios of 10-20x** are common for time-series data because adjacent timestamps and values compress extremely well with delta/gorilla encoding.

7. **Choose by your SQL needs**: TimescaleDB for full SQL + JOINs, InfluxDB for simpler metric pipelines, ClickHouse for heavy analytics, Prometheus for monitoring + alerting.

---

Next: [08-search-engines.md](08-search-engines.md) →
