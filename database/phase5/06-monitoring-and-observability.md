# 5.6 — Monitoring & Observability

> If you can't see it, you can't fix it.  
> A well-monitored database tells you about problems BEFORE users notice.

---

## 1. Key Metrics Dashboard

```
The 10 metrics that must be on every database dashboard:

1. Query throughput (QPS):        queries per second (total, reads, writes)
2. Query latency (p50/p95/p99):   how long queries take
3. Active connections:             current connection count vs max_connections
4. Cache hit ratio:                shared_buffers hit rate (target: >99%)
5. Replication lag:                bytes or seconds behind primary
6. Disk I/O:                       IOPS, throughput, I/O wait %
7. CPU usage:                      user, system, iowait
8. Dead tuples / bloat:            vacuum effectiveness
9. Lock waits:                     blocked queries, deadlocks
10. Transaction rate:              commits/sec, rollbacks/sec
```

---

## 2. PostgreSQL Monitoring Queries

```sql
-- Connection status:
SELECT state, count(*)
FROM pg_stat_activity
GROUP BY state;
-- Watch for: too many 'active' or 'idle in transaction'

-- Cache hit ratio (should be > 99%):
SELECT
    ROUND(SUM(heap_blks_hit)::NUMERIC / NULLIF(SUM(heap_blks_hit + heap_blks_read), 0) * 100, 2) AS table_cache_hit_pct,
    ROUND(SUM(idx_blks_hit)::NUMERIC / NULLIF(SUM(idx_blks_hit + idx_blks_read), 0) * 100, 2) AS index_cache_hit_pct
FROM pg_statio_user_tables;

-- Replication lag:
SELECT client_addr, state,
       pg_wal_lsn_diff(pg_current_wal_lsn(), replay_lsn) AS replay_lag_bytes,
       now() - replay_lag AS replay_lag_time
FROM pg_stat_replication;

-- Table bloat (dead tuples):
SELECT relname, n_live_tup, n_dead_tup,
       ROUND(n_dead_tup::NUMERIC / NULLIF(n_live_tup + n_dead_tup, 0) * 100, 2) AS dead_pct,
       last_autovacuum
FROM pg_stat_user_tables
WHERE n_dead_tup > 10000
ORDER BY n_dead_tup DESC;

-- Lock waits (who is blocking whom):
SELECT blocked.pid AS blocked_pid,
       blocked_activity.query AS blocked_query,
       blocking.pid AS blocking_pid,
       blocking_activity.query AS blocking_query,
       now() - blocked_activity.query_start AS wait_duration
FROM pg_locks blocked
JOIN pg_locks blocking ON blocking.locktype = blocked.locktype
    AND blocking.relation IS NOT DISTINCT FROM blocked.relation
    AND blocking.pid != blocked.pid
JOIN pg_stat_activity blocked_activity ON blocked.pid = blocked_activity.pid
JOIN pg_stat_activity blocking_activity ON blocking.pid = blocking_activity.pid
WHERE NOT blocked.granted
ORDER BY wait_duration DESC;

-- Transaction ID age (wraparound risk):
SELECT datname, age(datfrozenxid) AS xid_age,
       ROUND(age(datfrozenxid)::NUMERIC / 2000000000 * 100, 2) AS pct_to_wraparound
FROM pg_database ORDER BY xid_age DESC;
-- Alert if pct_to_wraparound > 50%

-- Checkpoint frequency (if too frequent, increase max_wal_size):
SELECT * FROM pg_stat_bgwriter;
-- checkpoints_timed: scheduled checkpoints (normal)
-- checkpoints_req: forced checkpoints (too frequent = bad)
```

---

## 3. Prometheus + Grafana Stack

```yaml
# postgres_exporter: exposes PG metrics as Prometheus endpoints
# docker-compose.yml:
services:
  postgres-exporter:
    image: prometheuscommunity/postgres-exporter
    environment:
      DATA_SOURCE_NAME: "postgresql://monitor:password@postgres:5432/mydb?sslmode=disable"
    ports:
      - "9187:9187"

  prometheus:
    image: prom/prometheus
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml

  grafana:
    image: grafana/grafana
    ports:
      - "3000:3000"

# prometheus.yml:
scrape_configs:
  - job_name: 'postgres'
    static_configs:
      - targets: ['postgres-exporter:9187']
    scrape_interval: 15s

# Key Grafana dashboards:
# - PostgreSQL Overview (ID: 9628)
# - PostgreSQL Database (ID: 12273)
# - Node Exporter Full (ID: 1860) — for OS metrics

# Alert rules (prometheus):
groups:
  - name: postgres
    rules:
      - alert: HighReplicationLag
        expr: pg_replication_lag_seconds > 30
        for: 5m
        labels: { severity: critical }

      - alert: LowCacheHitRatio
        expr: pg_stat_database_blks_hit / (pg_stat_database_blks_hit + pg_stat_database_blks_read) < 0.99
        for: 15m
        labels: { severity: warning }

      - alert: TooManyConnections
        expr: pg_stat_activity_count > 0.8 * pg_settings_max_connections
        for: 5m
        labels: { severity: warning }

      - alert: HighDeadTupleRatio
        expr: pg_stat_user_tables_n_dead_tup / (pg_stat_user_tables_n_live_tup + 1) > 0.1
        for: 30m
        labels: { severity: warning }
```

---

## 4. Alerting Strategy

```
Tiered alerting (don't wake someone up for a warning):

P1 — Critical (page on-call):
  - Database down / primary unreachable
  - Replication lag > 5 minutes
  - Disk usage > 90%
  - XID wraparound age > 1.5 billion
  - Connection count = max_connections
  - Zero successful checkpoints in 30 minutes

P2 — Warning (Slack notification):
  - Cache hit ratio < 99%
  - Replication lag > 30 seconds
  - Disk usage > 75%
  - Dead tuple ratio > 10% on any table
  - Long-running transactions > 1 hour
  - Unusual query latency spike (p99 > 2x baseline)

P3 — Info (dashboard only):
  - Slow queries logged
  - Autovacuum runs
  - Connection pool saturation approaching
  - Backup completion status
```

---

## Key Takeaways

1. **Cache hit ratio > 99%** is the first health indicator. Below 99% = you need more `shared_buffers` or your working set doesn't fit in RAM.
2. **Prometheus + Grafana + postgres_exporter** is the standard open-source monitoring stack. Set it up on day one, not after the first outage.
3. **Monitor XID wraparound age.** Alert at 50% of the 2-billion limit. This is the silent killer that can freeze your entire database.
4. **Alert on replication lag, not just replication status.** A replica that's "connected" but 5 minutes behind is worse than one that's disconnected (at least you KNOW it's down).
5. **Tiered alerts prevent alert fatigue.** Only page humans for database-down and imminent data loss. Everything else goes to Slack or dashboards.

---

Next: [07-schema-migrations.md](07-schema-migrations.md) →
