# 5.2 — High Availability & Replication

> Downtime costs money. Every minute of database outage can cost  
> $5,000-$300,000+ depending on the business.  
> HA is not optional in production. It's day-one infrastructure.

---

## 1. HA Concepts

```
Availability targets:
  99.9%   (three nines)  = 8.76 hours downtime/year  (43 min/month)
  99.99%  (four nines)   = 52.6 minutes downtime/year (4.3 min/month)
  99.999% (five nines)   = 5.26 minutes downtime/year (26 sec/month)

RTO (Recovery Time Objective):
  Maximum acceptable downtime after a failure.
  "We must be back up within 30 seconds."

RPO (Recovery Point Objective):
  Maximum acceptable data loss measured in time.
  "We can afford to lose at most 1 second of transactions."
  
  RPO = 0: synchronous replication (no data loss)
  RPO > 0: asynchronous replication (some data loss possible)

Failure types:
  Process crash:     database process dies (restart in seconds)
  Server failure:    hardware/OS failure (failover to standby)
  Disk failure:      storage corruption (RAID, replicas)
  Network partition: node is alive but unreachable
  Datacenter outage: entire DC goes down (cross-DC failover)
  Region outage:     entire cloud region (cross-region setup)
```

---

## 2. PostgreSQL HA with Patroni

```
Patroni: the industry-standard HA solution for PostgreSQL.
Uses a DCS (Distributed Configuration Store) for leader election.

Architecture:
  ┌──────────────────────────────────────────────────┐
  │              DCS (etcd / Consul / ZooKeeper)      │
  │  Stores: who is leader, cluster config, state     │
  └──────────┬────────────┬─────────────┬────────────┘
             │            │             │
  ┌──────────▼──┐  ┌──────▼──────┐  ┌──▼──────────┐
  │ Patroni +   │  │ Patroni +    │  │ Patroni +    │
  │ PostgreSQL  │  │ PostgreSQL   │  │ PostgreSQL   │
  │ (Leader)    │  │ (Replica)    │  │ (Replica)    │
  │ R/W         │  │ R/O          │  │ R/O          │
  └──────────┬──┘  └──────┬──────┘  └──┬──────────┘
             │            │             │
  ┌──────────▼────────────▼─────────────▼────────────┐
  │         HAProxy / PgBouncer / Application          │
  │  Routes writes → leader, reads → any replica       │
  └──────────────────────────────────────────────────┘

How Patroni works:
  1. Each Patroni instance manages its local PostgreSQL
  2. Leader holds a LOCK in the DCS (etcd key with TTL)
  3. Leader renews the lock every ttl/3 seconds
  4. If leader fails to renew (crash, network issue):
     → Lock expires → other Patroni instances compete for lock
     → Winner promotes its PostgreSQL to primary
     → Losers reconfigure their PostgreSQL to follow new primary
  5. Automatic failover in <30 seconds typically
  
Patroni configuration (patroni.yml):
  scope: mydb-cluster
  name: node1
  
  etcd3:
    hosts: etcd1:2379,etcd2:2379,etcd3:2379
  
  bootstrap:
    dcs:
      ttl: 30
      loop_wait: 10
      retry_timeout: 10
      maximum_lag_on_failover: 1048576  # 1MB — don't promote if too far behind
      synchronous_mode: true            # sync replication for RPO=0
    postgresql:
      parameters:
        wal_level: replica
        max_wal_senders: 10
        max_replication_slots: 10
        hot_standby: 'on'
  
  postgresql:
    listen: 0.0.0.0:5432
    data_dir: /var/lib/postgresql/data
    authentication:
      replication:
        username: replicator
        password: secret

Patroni operations:
  patronictl list              # show cluster state
  patronictl switchover        # planned failover (zero data loss)
  patronictl failover          # emergency failover
  patronictl reinit node2      # rebuild a replica from scratch
  patronictl edit-config       # change cluster-wide postgresql.conf
```

---

## 3. MySQL HA Solutions

```
MySQL InnoDB Cluster:
  Group Replication (Paxos) + MySQL Router + MySQL Shell
  Single-primary or multi-primary mode
  Automatic failover with MySQL Router

Orchestrator:
  Open-source topology manager for MySQL replication
  Detects master failure, promotes best replica
  Understands complex replication topologies
  
  orchestrator-client -c topology -i master:3306
  orchestrator-client -c graceful-master-takeover -i master:3306

ProxySQL:
  MySQL protocol-aware proxy
  Query routing: writes → master, reads → replicas
  Connection pooling and multiplexing
  Query caching, query rules, automatic failover detection

Galera Cluster (MariaDB/Percona):
  Synchronous multi-master using certification-based replication
  All nodes accept writes
  ✓ No replication lag
  ✗ Higher write latency (all nodes certify)
  ✗ InnoDB only, no LOCK TABLES
```

---

## 4. Load Balancing and Read Routing

```
Connection routing strategies:

HAProxy (Layer 4 TCP):
  frontend pg_write
    bind *:5432
    default_backend pg_primary
  
  frontend pg_read
    bind *:5433
    default_backend pg_replicas
  
  backend pg_primary
    option httpchk GET /primary     # Patroni REST API health check
    http-check expect status 200
    server node1 10.0.0.1:5432 check port 8008
    server node2 10.0.0.2:5432 check port 8008
    server node3 10.0.0.3:5432 check port 8008
  
  backend pg_replicas
    balance roundrobin
    option httpchk GET /replica
    server node1 10.0.0.1:5432 check port 8008
    server node2 10.0.0.2:5432 check port 8008
    server node3 10.0.0.3:5432 check port 8008

Application-level routing:
  # Many ORMs and connection libraries support read/write splitting:
  # Django: database routers
  # Rails: multiple database support
  # Spring: AbstractRoutingDataSource
  # libpq: target_session_attrs=read-write (connect to primary only)
  
  # PostgreSQL libpq multi-host:
  postgresql://node1,node2,node3/mydb?target_session_attrs=read-write
  # Tries each host, connects to the one that's primary

Handling replication lag:
  Problem: write to primary, immediately read from replica → stale data!
  
  Solutions:
  1. Read-your-writes: route reads to primary for N seconds after a write
  2. Causal consistency: track LSN, wait for replica to catch up
     SELECT pg_current_wal_lsn();           -- after write on primary
     SELECT pg_last_wal_replay_lsn();        -- check on replica
  3. Synchronous replication: zero lag, but higher write latency
  4. Application-aware: critical reads from primary, eventual reads from replicas
```

---

## 5. Zero-Downtime Failover Checklist

```
Before failover:
  ✓ Replicas are caught up (replication lag < threshold)
  ✓ Connection pool can drain (graceful close, not kill)
  ✓ Application handles connection errors with retry logic
  ✓ DNS TTL is low enough (or use IP-based routing)

During failover:
  1. Detect failure (health check fails, DCS lock expires)
  2. Promote replica to primary (promote trigger / pg_promote())
  3. Reconfigure other replicas to follow new primary
  4. Update routing (HAProxy backend switch / DNS update)
  5. Application reconnects (connection pool refreshes)

After failover:
  ✓ Verify new primary is accepting writes
  ✓ Verify replicas are streaming from new primary
  ✓ Handle old primary (prevent split-brain: fence it!)
  ✓ Rebuild old primary as a replica (pg_rewind or pg_basebackup)
  ✓ Post-mortem: why did failover happen?

Total downtime target: <30 seconds (Patroni typically achieves 10-30s)
```

---

## Key Takeaways

1. **Patroni + etcd + HAProxy** is the production-standard PostgreSQL HA stack. Automatic failover in <30 seconds.
2. **Synchronous replication = RPO 0** (zero data loss) but higher latency. Asynchronous = faster but potential data loss on failover.
3. **Connection routing** must handle the read-your-writes problem. Route critical reads to primary, eventual reads to replicas.
4. **Applications must handle reconnection gracefully.** Retry logic, circuit breakers, and connection pool refresh are essential.
5. **Fencing the old primary** prevents split-brain. Use STONITH, DCS locks, or `recovery.conf` to guarantee only one writer.

---

Next: [03-sharding-and-partitioning.md](03-sharding-and-partitioning.md) →
