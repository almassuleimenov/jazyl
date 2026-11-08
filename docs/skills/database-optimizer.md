# database-optimizer

> Источник: https://github.com/jeffallan/claude-skills/tree/main/skills/database-optimizer (MIT). Рабочая копия скилла: `.claude/skills/database-optimizer/`.

---
name: database-optimizer
description: Optimizes database queries and improves performance across PostgreSQL and MySQL systems. Use when investigating slow queries, analyzing execution plans, or optimizing database performance. Invoke for index design, query rewrites, configuration tuning, partitioning strategies, lock contention resolution.
license: MIT
metadata:
  author: https://github.com/Jeffallan
  company: https://synergetic.solutions
  version: "1.1.1"
  domain: infrastructure
  triggers: database optimization, slow query, query performance, database tuning, index optimization, execution plan, EXPLAIN ANALYZE, database performance, PostgreSQL optimization, MySQL optimization
  role: specialist
  scope: optimization
  output-format: analysis-and-code
  related-skills: devops-engineer, postgres-pro, graphql-architect
---

# Database Optimizer

Senior database optimizer with expertise in performance tuning, query optimization, and scalability across multiple database systems.

## When to Use This Skill

- Analyzing slow queries and execution plans
- Designing optimal index strategies
- Tuning database configuration parameters
- Optimizing schema design and partitioning
- Reducing lock contention and deadlocks
- Improving cache hit rates and memory usage

## Core Workflow

1. **Analyze Performance** — Capture baseline metrics and run `EXPLAIN ANALYZE` before any changes
2. **Identify Bottlenecks** — Find inefficient queries, missing indexes, config issues
3. **Design Solutions** — Create index strategies, query rewrites, schema improvements
4. **Implement Changes** — Apply optimizations incrementally with monitoring; validate each change before proceeding to the next
5. **Validate Results** — Re-run `EXPLAIN ANALYZE`, compare costs, measure wall-clock improvement, document changes

> ⚠️ Always test changes in non-production first. Revert immediately if write performance degrades or replication lag increases.

## Reference Guide

Load detailed guidance based on context:

| Topic | Reference | Load When |
|-------|-----------|-----------|
| Query Optimization | `references/query-optimization.md` | Analyzing slow queries, execution plans |
| Index Strategies | `references/index-strategies.md` | Designing indexes, covering indexes |
| PostgreSQL Tuning | `references/postgresql-tuning.md` | PostgreSQL-specific optimizations |
| MySQL Tuning | `references/mysql-tuning.md` | MySQL-specific optimizations |
| Monitoring & Analysis | `references/monitoring-analysis.md` | Performance metrics, diagnostics |

## Common Operations & Examples

### Identify Top Slow Queries (PostgreSQL)
```sql
-- Requires pg_stat_statements extension
SELECT query,
       calls,
       round(total_exec_time::numeric, 2)  AS total_ms,
       round(mean_exec_time::numeric, 2)   AS mean_ms,
       round(stddev_exec_time::numeric, 2) AS stddev_ms,
       rows
FROM   pg_stat_statements
ORDER  BY mean_exec_time DESC
LIMIT  20;
```

### Capture an Execution Plan
```sql
-- Use BUFFERS to expose cache hit vs. disk read ratio
EXPLAIN (ANALYZE, BUFFERS, FORMAT TEXT)
SELECT o.id, c.name
FROM   orders o
JOIN   customers c ON c.id = o.customer_id
WHERE  o.status = 'pending'
  AND  o.created_at > now() - interval '7 days';
```

### Reading EXPLAIN Output — Key Patterns to Find

| Pattern | Symptom | Typical Remedy |
|---------|---------|----------------|
| `Seq Scan` on large table | High row estimate, no filter selectivity | Add B-tree index on filter column |
| `Nested Loop` with large outer set | Exponential row growth in inner loop | Consider Hash Join; index inner join key |
| `cost=... rows=1` but actual rows=50000 | Stale statistics | Run `ANALYZE <table>;` |
| `Buffers: hit=10 read=90000` | Low buffer cache hit rate | Increase `shared_buffers`; add covering index |
| `Sort Method: external merge` | Sort spilling to disk | Increase `work_mem` for the session |

### Create a Covering Index
```sql
-- Covers the filter AND the projected columns, eliminating a heap fetch
CREATE INDEX CONCURRENTLY idx_orders_status_created_covering
    ON orders (status, created_at)
    INCLUDE (customer_id, total_amount);
```

### Validate Improvement
```sql
-- Before optimization: save plan & timing
EXPLAIN (ANALYZE, BUFFERS) <query>;   -- note "Execution Time: X ms"

-- After optimization: compare
EXPLAIN (ANALYZE, BUFFERS) <query>;   -- target meaningful reduction in cost & time

-- Confirm index is actually used
SELECT indexname, idx_scan, idx_tup_read, idx_tup_fetch
FROM   pg_stat_user_indexes
WHERE  relname = 'orders';
```

### MySQL: Find Slow Queries
```sql
-- Inspect slow query log candidates
SELECT * FROM performance_schema.events_statements_summary_by_digest
ORDER  BY SUM_TIMER_WAIT DESC
LIMIT  20;

-- Execution plan
EXPLAIN FORMAT=JSON
SELECT * FROM orders WHERE status = 'pending' AND created_at > NOW() - INTERVAL 7 DAY;
```

## Constraints

### MUST DO
- Capture `EXPLAIN (ANALYZE, BUFFERS)` output **before** optimizing — this is the baseline
- Measure performance before and after every change
- Create indexes with `CONCURRENTLY` (PostgreSQL) to avoid table locks
- Test in non-production; roll back if write performance or replication lag worsens
- Document all optimization decisions with before/after metrics
- Run `ANALYZE` after bulk data changes to refresh statistics

### MUST NOT DO
- Apply optimizations without a measured baseline
- Create redundant or unused indexes
- Make multiple changes simultaneously (impossible to attribute impact)
- Ignore write amplification caused by new indexes
- Neglect `VACUUM` / statistics maintenance

## Output Templates

When optimizing database performance, provide:
1. Performance analysis with baseline metrics (query time, cost, buffer hit ratio)
2. Identified bottlenecks and root causes (with EXPLAIN evidence)
3. Optimization strategy with specific changes
4. Implementation SQL / config changes
5. Validation queries to measure improvement
6. Monitoring recommendations

Maintained by [@jeffallan](https://github.com/jeffallan), Principal Consultant at [Synergetic Solutions](https://synergetic.solutions)

[Documentation](https://jeffallan.github.io/claude-skills/skills/infrastructure/database-optimizer/)

---

## Reference: index-strategies

# Index Strategies

## Index Selection Methodology

### Identify Index Candidates

```sql
-- PostgreSQL: Find queries missing indexes
SELECT query, calls, total_exec_time, mean_exec_time
FROM pg_stat_statements
WHERE mean_exec_time > 100
ORDER BY total_exec_time DESC
LIMIT 20;

-- PostgreSQL: Find sequential scans on large tables
SELECT schemaname, tablename, seq_scan, seq_tup_read,
       idx_scan, seq_tup_read / seq_scan as avg_seq_tup_read
FROM pg_stat_user_tables
WHERE seq_scan > 0
  AND seq_tup_read / seq_scan > 10000
ORDER BY seq_tup_read DESC;

-- MySQL: Check table scans
SELECT * FROM sys.statements_with_full_table_scans
WHERE db = 'your_database'
ORDER BY exec_count DESC;
```

## B-Tree Indexes (Default)

### Single Column Indexes

```sql
-- Create index for WHERE clauses
CREATE INDEX idx_users_email ON users(email);

-- Create index for JOIN conditions
CREATE INDEX idx_orders_user_id ON orders(user_id);

-- Create index for ORDER BY
CREATE INDEX idx_products_price ON products(price);

-- Unique constraint as index
CREATE UNIQUE INDEX idx_users_username ON users(username);
```

### Multi-Column Indexes

```sql
-- Order matters: most selective column first
CREATE INDEX idx_orders_status_created
ON orders(status, created_at);

-- Good for queries:
-- WHERE status = 'pending'
-- WHERE status = 'pending' AND created_at > '2024-01-01'
-- WHERE status = 'pending' ORDER BY created_at

-- NOT good for:
-- WHERE created_at > '2024-01-01' (status not specified)

-- Include commonly queried columns
CREATE INDEX idx_users_active_email_name
ON users(active, email) INCLUDE (name);
```

### Column Order Guidelines

```sql
-- Rule 1: Equality before range
CREATE INDEX idx_events_type_timestamp
ON events(type, timestamp);  -- type = 'click' AND timestamp > ...

-- Rule 2: High selectivity first
CREATE INDEX idx_orders_user_status
ON orders(user_id, status);  -- user_id is more selective than status

-- Rule 3: Match query patterns
-- Query: WHERE country = 'US' AND city = 'NYC' AND zip = '10001'
CREATE INDEX idx_locations_country_city_zip
ON locations(country, city, zip);
```

## Covering Indexes

### PostgreSQL INCLUDE Clause

```sql
-- Include non-key columns for index-only scans
CREATE INDEX idx_users_email_covering
ON users(email) INCLUDE (name, created_at);

-- Query can be satisfied entirely from index
EXPLAIN (ANALYZE, BUFFERS)
SELECT name, created_at
FROM users
WHERE email = 'user@example.com';
-- Should show "Index Only Scan"
```

### MySQL Covering Indexes

```sql
-- MySQL: Add columns to end of index
CREATE INDEX idx_orders_user_covering
ON orders(user_id, status, created_at, total);

-- Query uses covering index
EXPLAIN
SELECT status, created_at, total
FROM orders
WHERE user_id = 123;
-- Should show "Using index" in Extra column
```

## Partial Indexes

### PostgreSQL Partial Indexes

```sql
-- Index only active users
CREATE INDEX idx_users_active_email
ON users(email)
WHERE active = true;

-- Index only recent orders
CREATE INDEX idx_orders_recent
ON orders(user_id, created_at)
WHERE created_at > NOW() - INTERVAL '30 days';

-- Index only pending/processing orders (ignore completed)
CREATE INDEX idx_orders_active
ON orders(status, user_id)
WHERE status IN ('pending', 'processing');

-- Smaller index = better performance + less storage
```

### MySQL Filtered Indexes (8.0+)

```sql
-- MySQL 8.0+ supports functional indexes for similar effect
CREATE INDEX idx_users_active
ON users((CASE WHEN active = 1 THEN email END));
```

## Expression Indexes

### PostgreSQL Function Indexes

```sql
-- Index for case-insensitive search
CREATE INDEX idx_users_email_lower
ON users(LOWER(email));

-- Query must match expression
SELECT * FROM users
WHERE LOWER(email) = LOWER('User@Example.com');

-- Index for JSONB queries
CREATE INDEX idx_users_settings_theme
ON users((settings->>'theme'));

SELECT * FROM users
WHERE settings->>'theme' = 'dark';

-- Index for date truncation
CREATE INDEX idx_orders_date
ON orders(DATE(created_at));
```

### MySQL Generated Column Indexes

```sql
-- Create generated column, then index it
ALTER TABLE users
ADD COLUMN email_lower VARCHAR(255)
GENERATED ALWAYS AS (LOWER(email)) STORED;

CREATE INDEX idx_users_email_lower
ON users(email_lower);

-- Use in queries
SELECT * FROM users
WHERE email_lower = LOWER('User@Example.com');
```

## Specialized Index Types

### PostgreSQL GIN Indexes (Full-Text, Arrays, JSONB)

```sql
-- Full-text search
CREATE INDEX idx_posts_search
ON posts USING GIN(to_tsvector('english', title || ' ' || content));

SELECT * FROM posts
WHERE to_tsvector('english', title || ' ' || content)
      @@ to_tsquery('english', 'database & optimization');

-- Array search
CREATE INDEX idx_products_tags
ON products USING GIN(tags);

SELECT * FROM products
WHERE tags @> ARRAY['electronics', 'sale'];

-- JSONB search
CREATE INDEX idx_users_metadata
ON users USING GIN(metadata);

SELECT * FROM users
WHERE metadata @> '{"plan": "premium"}';
```

### PostgreSQL GiST Indexes (Geometric, Range)

```sql
-- Range types
CREATE INDEX idx_events_time_range
ON events USING GIST(time_range);

SELECT * FROM events
WHERE time_range && '[2024-01-01, 2024-01-31]'::tstzrange;

-- PostGIS geometric queries
CREATE INDEX idx_locations_coords
ON locations USING GIST(coordinates);
```

### MySQL Full-Text Indexes

```sql
-- Full-text search
CREATE FULLTEXT INDEX idx_posts_content
ON posts(title, content);

SELECT * FROM posts
WHERE MATCH(title, content)
      AGAINST('database optimization' IN NATURAL LANGUAGE MODE);

-- Boolean mode for complex searches
SELECT * FROM posts
WHERE MATCH(title, content)
      AGAINST('+database -mysql' IN BOOLEAN MODE);
```

## Index Maintenance

### PostgreSQL Maintenance

```sql
-- Update statistics for query planner
ANALYZE users;

-- Rebuild bloated index
REINDEX INDEX CONCURRENTLY idx_users_email;

-- Check index bloat
SELECT
    schemaname, tablename, indexname,
    pg_size_pretty(pg_relation_size(indexrelid)) as index_size,
    idx_scan as scans,
    idx_tup_read as tuples_read,
    idx_tup_fetch as tuples_fetched
FROM pg_stat_user_indexes
ORDER BY pg_relation_size(indexrelid) DESC;

-- Find unused indexes
SELECT
    schemaname, tablename, indexname,
    idx_scan,
    pg_size_pretty(pg_relation_size(indexrelid)) as index_size
FROM pg_stat_user_indexes
WHERE idx_scan = 0
  AND indexrelname NOT LIKE 'pg_toast%'
ORDER BY pg_relation_size(indexrelid) DESC;
```

### MySQL Maintenance

```sql
-- Update statistics
ANALYZE TABLE users;

-- Rebuild index
ALTER TABLE users DROP INDEX idx_users_email, ADD INDEX idx_users_email(email);

-- Check index usage
SELECT
    object_schema,
    object_name,
    index_name,
    count_star,
    count_read,
    count_fetch
FROM performance_schema.table_io_waits_summary_by_index_usage
WHERE object_schema = 'your_database'
ORDER BY count_star DESC;

-- Find unused indexes
SELECT
    object_schema,
    object_name,
    index_name
FROM performance_schema.table_io_waits_summary_by_index_usage
WHERE index_name IS NOT NULL
  AND count_star = 0
  AND object_schema = 'your_database';
```

## Index Anti-Patterns

| Anti-Pattern | Issue | Solution |
|-------------|-------|----------|
| Index every column | Write overhead, storage waste | Index based on query patterns |
| Redundant indexes | `(a)` + `(a,b)` | Keep only `(a,b)` |
| Wrong column order | `(created_at, user_id)` for `WHERE user_id = ?` | Put filtered columns first |
| Over-covering | Including rarely-used columns | Include only frequently accessed columns |
| Ignoring WHERE clause | Full index for 5% of data | Use partial indexes |
| Expression mismatch | Index `email`, query `LOWER(email)` | Create expression index |

## Index Design Checklist

1. **Analyze queries**: Use pg_stat_statements or slow query log
2. **Check execution plans**: Look for Seq Scan on large tables
3. **Design indexes**: Equality → Range → Include
4. **Create concurrently**: Avoid locking (PostgreSQL)
5. **Validate improvement**: Compare before/after EXPLAIN
6. **Monitor usage**: Remove unused indexes after 30 days
7. **Maintain regularly**: VACUUM, ANALYZE, REINDEX as needed

---

## Reference: monitoring-analysis

# Monitoring and Analysis

## PostgreSQL Monitoring

### Essential Extensions

```sql
-- Install performance monitoring extensions
CREATE EXTENSION IF NOT EXISTS pg_stat_statements;
CREATE EXTENSION IF NOT EXISTS pg_buffercache;
CREATE EXTENSION IF NOT EXISTS pg_trgm;  -- For similarity searches

-- Reset statistics
SELECT pg_stat_statements_reset();
SELECT pg_stat_reset();
```

### Query Performance Tracking

```sql
-- Top queries by total time
SELECT
    substring(query, 1, 100) as short_query,
    round(total_exec_time::numeric, 2) as total_time_ms,
    calls,
    round(mean_exec_time::numeric, 2) as mean_time_ms,
    round(stddev_exec_time::numeric, 2) as stddev_ms,
    round((100 * total_exec_time / sum(total_exec_time) OVER ())::numeric, 2) as pct_total
FROM pg_stat_statements
WHERE query NOT LIKE '%pg_stat_statements%'
ORDER BY total_exec_time DESC
LIMIT 20;

-- Queries with high variance
SELECT
    substring(query, 1, 100) as short_query,
    calls,
    round(mean_exec_time::numeric, 2) as mean_ms,
    round(stddev_exec_time::numeric, 2) as stddev_ms,
    round(max_exec_time::numeric, 2) as max_ms,
    round((stddev_exec_time / NULLIF(mean_exec_time, 0))::numeric, 2) as coeff_var
FROM pg_stat_statements
WHERE calls > 100
  AND stddev_exec_time > mean_exec_time * 0.5
ORDER BY stddev_exec_time DESC
LIMIT 20;

-- I/O intensive queries
SELECT
    substring(query, 1, 100) as short_query,
    calls,
    shared_blks_hit,
    shared_blks_read,
    shared_blks_written,
    round((shared_blks_read::numeric / NULLIF(calls, 0)), 2) as reads_per_call,
    round((shared_blks_hit / NULLIF(shared_blks_hit + shared_blks_read, 0)::numeric * 100), 2) as cache_hit_pct
FROM pg_stat_statements
WHERE shared_blks_read > 0
ORDER BY shared_blks_read DESC
LIMIT 20;
```

### Connection and Lock Monitoring

```sql
-- Current activity
SELECT
    pid,
    usename,
    application_name,
    client_addr,
    state,
    state_change,
    query_start,
    now() - query_start as duration,
    wait_event_type,
    wait_event,
    substring(query, 1, 100) as query
FROM pg_stat_activity
WHERE state != 'idle'
ORDER BY query_start;

-- Blocking queries
SELECT
    blocked_locks.pid AS blocked_pid,
    blocked_activity.usename AS blocked_user,
    blocking_locks.pid AS blocking_pid,
    blocking_activity.usename AS blocking_user,
    blocked_activity.query AS blocked_query,
    blocking_activity.query AS blocking_query,
    blocked_activity.application_name AS blocked_app
FROM pg_catalog.pg_locks blocked_locks
JOIN pg_catalog.pg_stat_activity blocked_activity ON blocked_activity.pid = blocked_locks.pid
JOIN pg_catalog.pg_locks blocking_locks
    ON blocking_locks.locktype = blocked_locks.locktype
    AND blocking_locks.database IS NOT DISTINCT FROM blocked_locks.database
    AND blocking_locks.relation IS NOT DISTINCT FROM blocked_locks.relation
    AND blocking_locks.page IS NOT DISTINCT FROM blocked_locks.page
    AND blocking_locks.tuple IS NOT DISTINCT FROM blocked_locks.tuple
    AND blocking_locks.virtualxid IS NOT DISTINCT FROM blocked_locks.virtualxid
    AND blocking_locks.transactionid IS NOT DISTINCT FROM blocked_locks.transactionid
    AND blocking_locks.classid IS NOT DISTINCT FROM blocked_locks.classid
    AND blocking_locks.objid IS NOT DISTINCT FROM blocked_locks.objid
    AND blocking_locks.objsubid IS NOT DISTINCT FROM blocked_locks.objsubid
    AND blocking_locks.pid != blocked_locks.pid
JOIN pg_catalog.pg_stat_activity blocking_activity ON blocking_activity.pid = blocking_locks.pid
WHERE NOT blocked_locks.granted;

-- Wait events summary
SELECT
    wait_event_type,
    wait_event,
    count(*) as waiting_connections
FROM pg_stat_activity
WHERE wait_event IS NOT NULL
GROUP BY wait_event_type, wait_event
ORDER BY waiting_connections DESC;
```

### Table and Index Statistics

```sql
-- Table bloat and dead tuples
SELECT
    schemaname,
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as total_size,
    n_live_tup,
    n_dead_tup,
    round(n_dead_tup * 100.0 / NULLIF(n_live_tup + n_dead_tup, 0), 2) as dead_pct,
    last_vacuum,
    last_autovacuum,
    last_analyze,
    last_autoanalyze
FROM pg_stat_user_tables
WHERE n_live_tup > 1000
ORDER BY n_dead_tup DESC;

-- Index usage and efficiency
SELECT
    schemaname,
    tablename,
    indexname,
    idx_scan,
    idx_tup_read,
    idx_tup_fetch,
    pg_size_pretty(pg_relation_size(indexrelid)) as size,
    CASE
        WHEN idx_scan = 0 THEN 'UNUSED'
        WHEN idx_tup_read = 0 THEN 'NEVER_READ'
        ELSE 'ACTIVE'
    END as status
FROM pg_stat_user_indexes
ORDER BY pg_relation_size(indexrelid) DESC;

-- Sequential scans on large tables
SELECT
    schemaname,
    tablename,
    seq_scan,
    seq_tup_read,
    idx_scan,
    n_live_tup,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as size
FROM pg_stat_user_tables
WHERE seq_scan > 0
  AND n_live_tup > 10000
  AND seq_tup_read / NULLIF(seq_scan, 0) > 10000
ORDER BY seq_tup_read DESC;
```

### Database Statistics

```sql
-- Database size and activity
SELECT
    datname,
    pg_size_pretty(pg_database_size(datname)) as size,
    numbackends as connections,
    xact_commit,
    xact_rollback,
    round(xact_rollback * 100.0 / NULLIF(xact_commit + xact_rollback, 0), 2) as rollback_pct,
    blks_read,
    blks_hit,
    round(blks_hit * 100.0 / NULLIF(blks_hit + blks_read, 0), 2) as cache_hit_pct
FROM pg_stat_database
WHERE datname NOT IN ('template0', 'template1', 'postgres')
ORDER BY pg_database_size(datname) DESC;

-- Checkpoint and bgwriter statistics
SELECT
    checkpoints_timed,
    checkpoints_req,
    checkpoint_write_time,
    checkpoint_sync_time,
    buffers_checkpoint,
    buffers_clean,
    buffers_backend,
    buffers_alloc,
    round(100.0 * checkpoints_req / NULLIF(checkpoints_timed + checkpoints_req, 0), 2) as req_checkpoint_pct
FROM pg_stat_bgwriter;
```

## MySQL Monitoring

### Performance Schema Queries

```sql
-- Top statements by total latency
SELECT
    DIGEST_TEXT as query,
    COUNT_STAR as exec_count,
    ROUND(AVG_TIMER_WAIT / 1000000000000, 3) as avg_sec,
    ROUND(SUM_TIMER_WAIT / 1000000000000, 3) as total_sec,
    ROUND(MAX_TIMER_WAIT / 1000000000000, 3) as max_sec,
    ROUND((SUM_TIMER_WAIT / SUM(SUM_TIMER_WAIT) OVER ()) * 100, 2) as pct_total
FROM performance_schema.events_statements_summary_by_digest
WHERE SCHEMA_NAME NOT IN ('performance_schema', 'mysql', 'sys')
ORDER BY SUM_TIMER_WAIT DESC
LIMIT 20;

-- Statements with full table scans
SELECT
    OBJECT_SCHEMA as db,
    OBJECT_NAME as tbl,
    COUNT_STAR as exec_count,
    SUM_NO_INDEX_USED as full_scans,
    SUM_NO_GOOD_INDEX_USED as bad_index
FROM performance_schema.table_io_waits_summary_by_index_usage
WHERE INDEX_NAME IS NULL
  AND OBJECT_SCHEMA NOT IN ('performance_schema', 'mysql', 'sys')
  AND COUNT_STAR > 0
ORDER BY SUM_NO_INDEX_USED DESC;

-- Table I/O statistics
SELECT
    OBJECT_SCHEMA,
    OBJECT_NAME,
    COUNT_READ,
    COUNT_WRITE,
    COUNT_FETCH,
    COUNT_INSERT,
    COUNT_UPDATE,
    COUNT_DELETE,
    ROUND(SUM_TIMER_WAIT / 1000000000000, 3) as total_latency_sec
FROM performance_schema.table_io_waits_summary_by_table
WHERE OBJECT_SCHEMA NOT IN ('performance_schema', 'mysql', 'sys')
ORDER BY SUM_TIMER_WAIT DESC
LIMIT 20;
```

### InnoDB Status Monitoring

```sql
-- InnoDB buffer pool status
SELECT
    POOL_ID,
    POOL_SIZE,
    FREE_BUFFERS,
    DATABASE_PAGES,
    OLD_DATABASE_PAGES,
    MODIFIED_DATABASE_PAGES,
    PENDING_DECOMPRESS,
    PENDING_READS,
    PENDING_FLUSH_LRU,
    PENDING_FLUSH_LIST
FROM information_schema.INNODB_BUFFER_POOL_STATS;

-- InnoDB lock waits
SELECT
    r.trx_id as waiting_trx,
    r.trx_mysql_thread_id as waiting_thread,
    r.trx_query as waiting_query,
    b.trx_id as blocking_trx,
    b.trx_mysql_thread_id as blocking_thread,
    b.trx_query as blocking_query
FROM information_schema.innodb_lock_waits w
INNER JOIN information_schema.innodb_trx b ON b.trx_id = w.blocking_trx_id
INNER JOIN information_schema.innodb_trx r ON r.trx_id = w.requesting_trx_id;

-- Long-running transactions
SELECT
    trx_id,
    trx_state,
    trx_started,
    TIMESTAMPDIFF(SECOND, trx_started, NOW()) as duration_sec,
    trx_requested_lock_id,
    trx_mysql_thread_id,
    trx_query
FROM information_schema.innodb_trx
WHERE TIMESTAMPDIFF(SECOND, trx_started, NOW()) > 60
ORDER BY trx_started;
```

### Connection and Process Monitoring

```sql
-- Current connections by state
SELECT
    command,
    state,
    COUNT(*) as connections,
    MAX(time) as max_time_sec
FROM information_schema.processlist
GROUP BY command, state
ORDER BY connections DESC;

-- Long-running queries
SELECT
    id,
    user,
    host,
    db,
    command,
    time,
    state,
    LEFT(info, 100) as query
FROM information_schema.processlist
WHERE command != 'Sleep'
  AND time > 10
ORDER BY time DESC;

-- Connection usage
SHOW STATUS LIKE 'Threads_%';
SHOW STATUS LIKE 'Max_used_connections';
SHOW VARIABLES LIKE 'max_connections';
```

### System Status Variables

```sql
-- Key buffer efficiency (MyISAM)
SHOW STATUS LIKE 'Key_%';

-- InnoDB metrics
SHOW STATUS LIKE 'Innodb_buffer_pool_%';
SHOW STATUS LIKE 'Innodb_rows_%';
SHOW STATUS LIKE 'Innodb_data_%';

-- Table locks
SHOW STATUS LIKE 'Table_locks_%';

-- Temporary tables
SHOW STATUS LIKE 'Created_tmp_%';

-- Thread cache
SHOW STATUS LIKE 'Threads_%';
SHOW STATUS LIKE 'Connections';

-- Query cache (MySQL 5.7)
SHOW STATUS LIKE 'Qcache_%';
```

## Cross-Platform Monitoring

### Resource Utilization

```sql
-- PostgreSQL: Database size growth
SELECT
    current_database() as database,
    pg_size_pretty(pg_database_size(current_database())) as size,
    (SELECT pg_size_pretty(sum(pg_total_relation_size(schemaname||'.'||tablename)))
     FROM pg_tables
     WHERE schemaname = 'public') as public_schema_size;

-- MySQL: Database size
SELECT
    table_schema as database,
    ROUND(SUM(data_length + index_length) / 1024 / 1024, 2) as size_mb
FROM information_schema.tables
WHERE table_schema NOT IN ('information_schema', 'performance_schema', 'mysql', 'sys')
GROUP BY table_schema
ORDER BY size_mb DESC;
```

### Health Check Queries

```sql
-- PostgreSQL: Overall health
SELECT
    'connections' as metric,
    count(*) as current,
    current_setting('max_connections')::int as max
FROM pg_stat_activity
UNION ALL
SELECT
    'cache_hit_ratio',
    round((sum(heap_blks_hit) / NULLIF(sum(heap_blks_hit) + sum(heap_blks_read), 0) * 100)::numeric, 2),
    95
FROM pg_statio_user_tables
UNION ALL
SELECT
    'database_size_gb',
    round((pg_database_size(current_database()) / 1024.0 / 1024.0 / 1024.0)::numeric, 2),
    NULL;

-- MySQL: Overall health
SELECT 'connections' as metric,
       (SELECT COUNT(*) FROM information_schema.processlist) as current,
       @@max_connections as max
UNION ALL
SELECT 'buffer_pool_hit_ratio',
       ROUND((1 - (
           (SELECT VARIABLE_VALUE FROM performance_schema.global_status WHERE VARIABLE_NAME = 'Innodb_buffer_pool_reads') /
           (SELECT VARIABLE_VALUE FROM performance_schema.global_status WHERE VARIABLE_NAME = 'Innodb_buffer_pool_read_requests')
       )) * 100, 2),
       95
UNION ALL
SELECT 'slow_queries',
       (SELECT VARIABLE_VALUE FROM performance_schema.global_status WHERE VARIABLE_NAME = 'Slow_queries'),
       NULL;
```

## Alert Thresholds

### PostgreSQL Alerts

```sql
-- Connection pool nearing capacity
SELECT
    count(*) as current_connections,
    current_setting('max_connections')::int as max_connections,
    CASE
        WHEN count(*) > current_setting('max_connections')::int * 0.9 THEN 'CRITICAL'
        WHEN count(*) > current_setting('max_connections')::int * 0.8 THEN 'WARNING'
        ELSE 'OK'
    END as status
FROM pg_stat_activity;

-- Cache hit ratio degradation
WITH cache_stats AS (
    SELECT
        round((sum(heap_blks_hit) / NULLIF(sum(heap_blks_hit) + sum(heap_blks_read), 0) * 100)::numeric, 2) as hit_ratio
    FROM pg_statio_user_tables
)
SELECT
    hit_ratio,
    CASE
        WHEN hit_ratio < 90 THEN 'CRITICAL'
        WHEN hit_ratio < 95 THEN 'WARNING'
        ELSE 'OK'
    END as status
FROM cache_stats;

-- Replication lag (on standby)
SELECT
    CASE
        WHEN pg_last_wal_receive_lsn() = pg_last_wal_replay_lsn() THEN 0
        ELSE EXTRACT(EPOCH FROM (now() - pg_last_xact_replay_timestamp()))
    END as lag_seconds,
    CASE
        WHEN EXTRACT(EPOCH FROM (now() - pg_last_xact_replay_timestamp())) > 60 THEN 'CRITICAL'
        WHEN EXTRACT(EPOCH FROM (now() - pg_last_xact_replay_timestamp())) > 10 THEN 'WARNING'
        ELSE 'OK'
    END as status;
```

### MySQL Alerts

```sql
-- InnoDB buffer pool efficiency
SELECT
    ROUND((1 - (
        (SELECT VARIABLE_VALUE FROM performance_schema.global_status WHERE VARIABLE_NAME = 'Innodb_buffer_pool_reads') /
        (SELECT VARIABLE_VALUE FROM performance_schema.global_status WHERE VARIABLE_NAME = 'Innodb_buffer_pool_read_requests')
    )) * 100, 2) as buffer_pool_hit_ratio,
    CASE
        WHEN (1 - (
            (SELECT VARIABLE_VALUE FROM performance_schema.global_status WHERE VARIABLE_NAME = 'Innodb_buffer_pool_reads') /
            (SELECT VARIABLE_VALUE FROM performance_schema.global_status WHERE VARIABLE_NAME = 'Innodb_buffer_pool_read_requests')
        )) * 100 < 90 THEN 'CRITICAL'
        WHEN (1 - (
            (SELECT VARIABLE_VALUE FROM performance_schema.global_status WHERE VARIABLE_NAME = 'Innodb_buffer_pool_reads') /
            (SELECT VARIABLE_VALUE FROM performance_schema.global_status WHERE VARIABLE_NAME = 'Innodb_buffer_pool_read_requests')
        )) * 100 < 95 THEN 'WARNING'
        ELSE 'OK'
    END as status;

-- Replication lag (on replica)
SELECT
    Seconds_Behind_Master as lag_seconds,
    CASE
        WHEN Slave_IO_Running = 'No' OR Slave_SQL_Running = 'No' THEN 'CRITICAL - Replication stopped'
        WHEN Seconds_Behind_Master > 300 THEN 'CRITICAL'
        WHEN Seconds_Behind_Master > 60 THEN 'WARNING'
        ELSE 'OK'
    END as status
FROM (SHOW SLAVE STATUS) s;
```

## Monitoring Best Practices

1. **Establish baselines** - Record normal performance metrics
2. **Track trends** - Monitor daily/weekly patterns
3. **Set thresholds** - Define warning and critical levels
4. **Automate alerts** - Use monitoring tools (Prometheus, Grafana, Datadog)
5. **Regular reviews** - Weekly performance analysis meetings
6. **Document changes** - Track configuration and schema modifications
7. **Capacity planning** - Monitor growth and forecast needs
8. **Test queries** - Validate optimizations in staging first

---

## Reference: mysql-tuning

# MySQL Tuning

## InnoDB Memory Configuration

### Buffer Pool

```sql
-- Recommended: 70-80% of system RAM for dedicated MySQL server
-- For 16GB RAM server:
SET GLOBAL innodb_buffer_pool_size = 12884901888;  -- 12GB

-- Check buffer pool usage
SHOW GLOBAL STATUS LIKE 'Innodb_buffer_pool_%';

-- Buffer pool hit ratio (target: >99%)
SELECT
    (1 - (Innodb_buffer_pool_reads / Innodb_buffer_pool_read_requests)) * 100 as hit_ratio
FROM (
    SELECT
        VARIABLE_VALUE as Innodb_buffer_pool_reads
    FROM performance_schema.global_status
    WHERE VARIABLE_NAME = 'Innodb_buffer_pool_reads'
) reads,
(
    SELECT
        VARIABLE_VALUE as Innodb_buffer_pool_read_requests
    FROM performance_schema.global_status
    WHERE VARIABLE_NAME = 'Innodb_buffer_pool_read_requests'
) requests;

-- Buffer pool instances (for multi-core systems)
-- Recommended: 1 instance per 1GB, max 64
SET GLOBAL innodb_buffer_pool_instances = 8;
```

### Sort and Join Buffers

```sql
-- Sort buffer per connection
SET GLOBAL sort_buffer_size = 2097152;  -- 2MB

-- Join buffer for full joins
SET GLOBAL join_buffer_size = 2097152;  -- 2MB

-- Temporary table size
SET GLOBAL tmp_table_size = 67108864;  -- 64MB
SET GLOBAL max_heap_table_size = 67108864;  -- 64MB

-- Monitor temp table usage
SHOW GLOBAL STATUS LIKE 'Created_tmp%';
```

## Query Cache (Deprecated in 8.0)

```sql
-- MySQL 5.7 and earlier
-- Note: Removed in MySQL 8.0
SET GLOBAL query_cache_type = 1;
SET GLOBAL query_cache_size = 67108864;  -- 64MB

-- Check query cache effectiveness
SHOW STATUS LIKE 'Qcache%';

-- Query cache hit ratio
SELECT
    Qcache_hits / (Qcache_hits + Com_select) * 100 as cache_hit_ratio
FROM (
    SELECT VARIABLE_VALUE as Qcache_hits
    FROM performance_schema.global_status
    WHERE VARIABLE_NAME = 'Qcache_hits'
) hits,
(
    SELECT VARIABLE_VALUE as Com_select
    FROM performance_schema.global_status
    WHERE VARIABLE_NAME = 'Com_select'
) selects;
```

## InnoDB Performance Settings

### Log Files and Flushing

```sql
-- InnoDB log file size (larger = better write performance)
-- Recommended: 1-2GB for write-heavy workloads
SET GLOBAL innodb_log_file_size = 1073741824;  -- 1GB

-- Log buffer size
SET GLOBAL innodb_log_buffer_size = 16777216;  -- 16MB

-- Flush method (O_DIRECT for dedicated server, avoids double buffering)
-- Set in my.cnf
innodb_flush_method = O_DIRECT

-- Flush log at transaction commit
-- 1 = full ACID (default, safest)
-- 2 = write to OS cache, flush every second
-- 0 = write and flush every second (fastest, risk data loss)
SET GLOBAL innodb_flush_log_at_trx_commit = 1;

-- For replication slaves or analytics (trade safety for speed)
SET GLOBAL innodb_flush_log_at_trx_commit = 2;
```

### I/O Configuration

```sql
-- Read I/O threads
SET GLOBAL innodb_read_io_threads = 8;

-- Write I/O threads
SET GLOBAL innodb_write_io_threads = 8;

-- I/O capacity (IOPS your storage can handle)
-- For SSD: 5000-20000
SET GLOBAL innodb_io_capacity = 10000;
SET GLOBAL innodb_io_capacity_max = 20000;

-- Flush method for optimal I/O
-- my.cnf:
innodb_flush_method = O_DIRECT
innodb_flush_neighbors = 0  -- Disable for SSD
```

### Thread Configuration

```sql
-- Max connections
SET GLOBAL max_connections = 200;

-- Thread cache (reuse threads)
SET GLOBAL thread_cache_size = 100;

-- Check thread cache effectiveness
SHOW STATUS LIKE 'Threads_%';
SHOW STATUS LIKE 'Connections';

-- Thread cache hit ratio (target: >90%)
SELECT
    (1 - (Threads_created / Connections)) * 100 as thread_cache_hit_ratio
FROM (
    SELECT VARIABLE_VALUE as Threads_created
    FROM performance_schema.global_status
    WHERE VARIABLE_NAME = 'Threads_created'
) created,
(
    SELECT VARIABLE_VALUE as Connections
    FROM performance_schema.global_status
    WHERE VARIABLE_NAME = 'Connections'
) conns;
```

## Query Optimization

### Slow Query Log

```sql
-- Enable slow query logging
SET GLOBAL slow_query_log = 'ON';
SET GLOBAL long_query_time = 1.0;  -- Log queries > 1 second
SET GLOBAL log_queries_not_using_indexes = 'ON';

-- Slow query log file location
SET GLOBAL slow_query_log_file = '/var/log/mysql/slow-query.log';

-- Analyze slow query log with pt-query-digest
-- $ pt-query-digest /var/log/mysql/slow-query.log

-- Check slow query status
SHOW GLOBAL STATUS LIKE 'Slow_queries';
```

### Performance Schema

```sql
-- Enable performance schema (my.cnf)
performance_schema = ON

-- Top queries by total execution time
SELECT
    DIGEST_TEXT,
    COUNT_STAR as exec_count,
    ROUND(AVG_TIMER_WAIT / 1000000000000, 3) as avg_time_sec,
    ROUND(SUM_TIMER_WAIT / 1000000000000, 3) as total_time_sec,
    ROUND((SUM_TIMER_WAIT / SUM(SUM_TIMER_WAIT) OVER ()) * 100, 2) as pct
FROM performance_schema.events_statements_summary_by_digest
ORDER BY SUM_TIMER_WAIT DESC
LIMIT 10;

-- Full table scans
SELECT * FROM sys.statements_with_full_table_scans
ORDER BY exec_count DESC
LIMIT 10;

-- Tables with high I/O
SELECT
    object_schema,
    object_name,
    count_read,
    count_write,
    count_fetch,
    SUM_TIMER_WAIT / 1000000000000 as total_latency_sec
FROM performance_schema.table_io_waits_summary_by_table
WHERE object_schema NOT IN ('mysql', 'performance_schema', 'sys')
ORDER BY SUM_TIMER_WAIT DESC
LIMIT 10;
```

## Index Optimization

### Index Statistics

```sql
-- Update index statistics
ANALYZE TABLE users;

-- Check index cardinality
SHOW INDEX FROM users;

-- Find duplicate/redundant indexes
SELECT
    a.table_schema,
    a.table_name,
    a.index_name as index1,
    a.column_name,
    b.index_name as index2
FROM information_schema.statistics a
JOIN information_schema.statistics b
    ON a.table_schema = b.table_schema
    AND a.table_name = b.table_name
    AND a.seq_in_index = b.seq_in_index
    AND a.column_name = b.column_name
    AND a.index_name != b.index_name
WHERE a.table_schema NOT IN ('mysql', 'information_schema', 'performance_schema', 'sys')
ORDER BY a.table_schema, a.table_name, a.index_name;

-- Find unused indexes
SELECT
    object_schema,
    object_name,
    index_name
FROM performance_schema.table_io_waits_summary_by_index_usage
WHERE index_name IS NOT NULL
  AND count_star = 0
  AND object_schema NOT IN ('mysql', 'performance_schema', 'sys')
ORDER BY object_schema, object_name;
```

### Covering Indexes

```sql
-- Create covering index
CREATE INDEX idx_users_email_name_created
ON users(email, name, created_at);

-- Query can use covering index
EXPLAIN
SELECT name, created_at FROM users WHERE email = 'user@example.com';
-- Look for "Using index" in Extra column

-- Force index usage for testing
SELECT name FROM users FORCE INDEX (idx_users_email_name_created)
WHERE email = 'user@example.com';
```

## Partitioning

### Range Partitioning

```sql
-- Create partitioned table
CREATE TABLE events (
    id BIGINT NOT NULL AUTO_INCREMENT,
    event_type VARCHAR(50),
    created_at DATETIME NOT NULL,
    data JSON,
    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (YEAR(created_at)) (
    PARTITION p2023 VALUES LESS THAN (2024),
    PARTITION p2024 VALUES LESS THAN (2025),
    PARTITION p2025 VALUES LESS THAN (2026),
    PARTITION pmax VALUES LESS THAN MAXVALUE
);

-- Query with partition pruning
EXPLAIN PARTITIONS
SELECT * FROM events
WHERE created_at >= '2024-01-01' AND created_at < '2024-02-01';
-- Should show "partitions: p2024"

-- Add new partition
ALTER TABLE events
ADD PARTITION (PARTITION p2026 VALUES LESS THAN (2027));

-- Drop old partition (fast delete)
ALTER TABLE events DROP PARTITION p2023;
```

### List Partitioning

```sql
-- Partition by discrete values
CREATE TABLE orders (
    id BIGINT NOT NULL AUTO_INCREMENT,
    user_id BIGINT,
    status VARCHAR(20),
    PRIMARY KEY (id, status)
) PARTITION BY LIST COLUMNS(status) (
    PARTITION p_pending VALUES IN ('pending', 'processing'),
    PARTITION p_completed VALUES IN ('completed', 'shipped'),
    PARTITION p_cancelled VALUES IN ('cancelled', 'refunded')
);
```

## Replication Optimization

### Binary Log Settings

```sql
-- Binary log format
SET GLOBAL binlog_format = 'ROW';  -- ROW, STATEMENT, or MIXED

-- Binary log cache size
SET GLOBAL binlog_cache_size = 1048576;  -- 1MB per transaction

-- Sync binary log (durability vs performance)
SET GLOBAL sync_binlog = 1;  -- Safest, sync after each commit
-- sync_binlog = 0  -- Fastest, let OS handle flushing

-- Expire binary logs after N days
SET GLOBAL binlog_expire_logs_seconds = 604800;  -- 7 days
```

### Replication Lag Monitoring

```sql
-- On replica: Check replication lag
SHOW SLAVE STATUS\G

-- Parse seconds behind master
SELECT
    IF(Slave_IO_Running = 'Yes' AND Slave_SQL_Running = 'Yes',
       Seconds_Behind_Master,
       NULL) as replication_lag_seconds
FROM (SHOW SLAVE STATUS) s;

-- Parallel replication (MySQL 8.0+)
SET GLOBAL slave_parallel_workers = 4;
SET GLOBAL slave_parallel_type = 'LOGICAL_CLOCK';
```

## Table Optimization

### Table Maintenance

```sql
-- Optimize table (rebuilds, reclaims space)
OPTIMIZE TABLE users;

-- Check table for errors
CHECK TABLE users;

-- Repair table if corrupted
REPAIR TABLE users;

-- Analyze table statistics
ANALYZE TABLE users;

-- Check fragmentation
SELECT
    table_schema,
    table_name,
    ROUND(data_length / 1024 / 1024, 2) as data_mb,
    ROUND(data_free / 1024 / 1024, 2) as free_mb,
    ROUND(data_free / data_length * 100, 2) as fragmentation_pct
FROM information_schema.tables
WHERE table_schema NOT IN ('mysql', 'information_schema', 'performance_schema', 'sys')
  AND data_free > 0
ORDER BY fragmentation_pct DESC;
```

### Table Compression

```sql
-- InnoDB compression (requires ROW_FORMAT=COMPRESSED)
CREATE TABLE compressed_logs (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    message TEXT,
    created_at DATETIME
) ROW_FORMAT=COMPRESSED KEY_BLOCK_SIZE=8;

-- Check compression ratio
SELECT
    table_schema,
    table_name,
    ROUND(data_length / 1024 / 1024, 2) as data_mb,
    ROUND(index_length / 1024 / 1024, 2) as index_mb,
    create_options
FROM information_schema.tables
WHERE row_format = 'Compressed';
```

## Configuration File Example

```ini
# my.cnf - Production optimized for 16GB RAM server

[mysqld]
# InnoDB Settings
innodb_buffer_pool_size = 12G
innodb_buffer_pool_instances = 8
innodb_log_file_size = 1G
innodb_log_buffer_size = 16M
innodb_flush_log_at_trx_commit = 1
innodb_flush_method = O_DIRECT
innodb_flush_neighbors = 0

# I/O Settings
innodb_read_io_threads = 8
innodb_write_io_threads = 8
innodb_io_capacity = 10000
innodb_io_capacity_max = 20000

# Connection Settings
max_connections = 200
thread_cache_size = 100

# Query Cache (MySQL 5.7)
# query_cache_type = 1
# query_cache_size = 64M

# Temporary Tables
tmp_table_size = 64M
max_heap_table_size = 64M

# Slow Query Log
slow_query_log = ON
long_query_time = 1
log_queries_not_using_indexes = ON

# Binary Log
binlog_format = ROW
sync_binlog = 1
binlog_expire_logs_seconds = 604800

# Performance Schema
performance_schema = ON

# Character Set
character_set_server = utf8mb4
collation_server = utf8mb4_unicode_ci
```

---

## Reference: postgresql-tuning

# PostgreSQL Tuning

## Memory Configuration

### Shared Buffers

```sql
-- Recommended: 25% of system RAM (up to 40% for dedicated DB server)
-- For 16GB RAM server:
ALTER SYSTEM SET shared_buffers = '4GB';

-- Check current setting
SHOW shared_buffers;

-- Monitor buffer hit ratio (target: >99%)
SELECT
    sum(heap_blks_read) as heap_read,
    sum(heap_blks_hit) as heap_hit,
    round(sum(heap_blks_hit) / nullif(sum(heap_blks_hit) + sum(heap_blks_read), 0) * 100, 2) as cache_hit_ratio
FROM pg_statio_user_tables;
```

### Work Memory

```sql
-- Per-operation memory for sorting/hashing
-- Recommended: (Total RAM * 0.25) / max_connections
-- For 16GB RAM, 100 connections: ~40MB
ALTER SYSTEM SET work_mem = '40MB';

-- Monitor sorts
SELECT
    query,
    calls,
    total_exec_time,
    mean_exec_time,
    min_exec_time,
    max_exec_time
FROM pg_stat_statements
WHERE query LIKE '%ORDER BY%' OR query LIKE '%GROUP BY%'
ORDER BY total_exec_time DESC
LIMIT 10;

-- Set per-session for large operations
SET work_mem = '256MB';
SELECT ... ORDER BY ... LIMIT 1000;
RESET work_mem;
```

### Maintenance Work Memory

```sql
-- For VACUUM, CREATE INDEX, ALTER TABLE
-- Recommended: 1-2GB for production systems
ALTER SYSTEM SET maintenance_work_mem = '2GB';

-- Autovacuum workers use proportional amount
ALTER SYSTEM SET autovacuum_work_mem = '512MB';
```

### Effective Cache Size

```sql
-- Planner hint for available OS cache
-- Recommended: 50-75% of total RAM
-- For 16GB RAM:
ALTER SYSTEM SET effective_cache_size = '12GB';
```

## Query Planner Settings

### Statistics Target

```sql
-- Default is 100, increase for better estimates on complex queries
ALTER SYSTEM SET default_statistics_target = 200;

-- Per-column statistics for specific columns
ALTER TABLE users ALTER COLUMN email SET STATISTICS 500;

-- Force statistics update
ANALYZE users;

-- Check statistics quality
SELECT
    schemaname, tablename, attname,
    n_distinct, correlation
FROM pg_stats
WHERE tablename = 'users';
```

### Parallel Query Configuration

```sql
-- Enable parallel queries
ALTER SYSTEM SET max_parallel_workers_per_gather = 4;
ALTER SYSTEM SET max_parallel_workers = 8;
ALTER SYSTEM SET parallel_setup_cost = 100;
ALTER SYSTEM SET parallel_tuple_cost = 0.01;

-- Minimum rows to consider parallel execution
ALTER SYSTEM SET min_parallel_table_scan_size = '8MB';
ALTER SYSTEM SET min_parallel_index_scan_size = '512kB';

-- Check if query uses parallel execution
EXPLAIN (ANALYZE, BUFFERS)
SELECT COUNT(*) FROM large_table WHERE condition = 'value';
-- Look for "Parallel Seq Scan" or "Gather" nodes
```

### Join and Scan Methods

```sql
-- Enable all join methods (usually all enabled by default)
ALTER SYSTEM SET enable_hashjoin = on;
ALTER SYSTEM SET enable_mergejoin = on;
ALTER SYSTEM SET enable_nestloop = on;

-- Cost parameters (adjust based on hardware)
ALTER SYSTEM SET random_page_cost = 1.1;  -- For SSD (default 4.0 is for HDD)
ALTER SYSTEM SET seq_page_cost = 1.0;

-- Disable methods for testing (don't do in production)
SET enable_seqscan = off;  -- Force index usage for testing
```

## Write Performance Optimization

### WAL Configuration

```sql
-- WAL write strategy
ALTER SYSTEM SET wal_buffers = '16MB';
ALTER SYSTEM SET wal_writer_delay = '200ms';

-- Checkpoint configuration
ALTER SYSTEM SET checkpoint_completion_target = 0.9;
ALTER SYSTEM SET max_wal_size = '2GB';
ALTER SYSTEM SET min_wal_size = '1GB';

-- Monitor checkpoints
SELECT
    checkpoints_timed,
    checkpoints_req,
    checkpoint_write_time,
    checkpoint_sync_time,
    buffers_checkpoint,
    buffers_clean,
    buffers_backend
FROM pg_stat_bgwriter;

-- Too many requested checkpoints = increase max_wal_size
```

### Commit Delays

```sql
-- Group commits (trade latency for throughput)
ALTER SYSTEM SET commit_delay = 10000;  -- 10ms
ALTER SYSTEM SET commit_siblings = 5;

-- Asynchronous commit (trade durability for speed)
-- Use cautiously - risk losing recent commits on crash
ALTER SYSTEM SET synchronous_commit = 'off';

-- Or per-transaction
BEGIN;
SET LOCAL synchronous_commit = 'off';
INSERT INTO logs (...) VALUES (...);
COMMIT;
```

## VACUUM and Autovacuum

### Autovacuum Configuration

```sql
-- Enable autovacuum (should always be on)
ALTER SYSTEM SET autovacuum = on;

-- Autovacuum worker settings
ALTER SYSTEM SET autovacuum_max_workers = 4;
ALTER SYSTEM SET autovacuum_naptime = '30s';

-- Thresholds for triggering autovacuum
ALTER SYSTEM SET autovacuum_vacuum_scale_factor = 0.1;  -- 10% dead tuples
ALTER SYSTEM SET autovacuum_vacuum_threshold = 50;

-- Analyze thresholds
ALTER SYSTEM SET autovacuum_analyze_scale_factor = 0.05;  -- 5% changed
ALTER SYSTEM SET autovacuum_analyze_threshold = 50;

-- Per-table autovacuum settings for high-churn tables
ALTER TABLE busy_table SET (
    autovacuum_vacuum_scale_factor = 0.01,  -- More aggressive
    autovacuum_vacuum_cost_delay = 2,       -- Faster vacuum
    autovacuum_vacuum_cost_limit = 1000
);
```

### Manual Vacuum Operations

```sql
-- Full vacuum (locks table, reclaims space)
VACUUM FULL users;  -- Use sparingly, requires exclusive lock

-- Regular vacuum (non-locking)
VACUUM (ANALYZE, VERBOSE) users;

-- Check table bloat
SELECT
    schemaname, tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as total_size,
    pg_size_pretty(pg_relation_size(schemaname||'.'||tablename)) as table_size,
    n_dead_tup,
    n_live_tup,
    round(n_dead_tup * 100.0 / NULLIF(n_live_tup + n_dead_tup, 0), 2) as dead_pct
FROM pg_stat_user_tables
WHERE n_live_tup > 0
ORDER BY n_dead_tup DESC;

-- Monitor autovacuum activity
SELECT
    schemaname, relname,
    last_vacuum, last_autovacuum,
    last_analyze, last_autoanalyze,
    vacuum_count, autovacuum_count,
    analyze_count, autoanalyze_count
FROM pg_stat_user_tables
ORDER BY last_autovacuum DESC NULLS LAST;
```

## Connection Pooling

### Configuration

```sql
-- Max connections (keep reasonable to manage memory)
ALTER SYSTEM SET max_connections = 200;

-- Reserved connections for superuser
ALTER SYSTEM SET superuser_reserved_connections = 3;

-- Connection lifecycle
ALTER SYSTEM SET idle_in_transaction_session_timeout = '5min';
ALTER SYSTEM SET statement_timeout = '30s';  -- Per-query timeout

-- Monitor connections
SELECT
    state,
    count(*),
    max(now() - state_change) as max_idle_time
FROM pg_stat_activity
WHERE state IS NOT NULL
GROUP BY state;

-- Find long-running queries
SELECT
    pid,
    now() - pg_stat_activity.query_start AS duration,
    query,
    state
FROM pg_stat_activity
WHERE (now() - pg_stat_activity.query_start) > interval '5 minutes'
  AND state != 'idle';
```

## Lock Management

### Lock Monitoring

```sql
-- Check current locks
SELECT
    locktype,
    relation::regclass,
    mode,
    granted,
    pid,
    pg_blocking_pids(pid) as blocked_by
FROM pg_locks
WHERE NOT granted
ORDER BY relation;

-- Find blocking queries
SELECT
    blocked_locks.pid AS blocked_pid,
    blocked_activity.usename AS blocked_user,
    blocking_locks.pid AS blocking_pid,
    blocking_activity.usename AS blocking_user,
    blocked_activity.query AS blocked_statement,
    blocking_activity.query AS blocking_statement
FROM pg_catalog.pg_locks blocked_locks
JOIN pg_catalog.pg_stat_activity blocked_activity ON blocked_activity.pid = blocked_locks.pid
JOIN pg_catalog.pg_locks blocking_locks ON blocking_locks.locktype = blocked_locks.locktype
    AND blocking_locks.relation = blocked_locks.relation
    AND blocking_locks.pid != blocked_locks.pid
JOIN pg_catalog.pg_stat_activity blocking_activity ON blocking_activity.pid = blocking_locks.pid
WHERE NOT blocked_locks.granted;

-- Deadlock configuration
ALTER SYSTEM SET deadlock_timeout = '1s';
ALTER SYSTEM SET log_lock_waits = on;
```

## Partitioning

### Range Partitioning

```sql
-- Create partitioned table
CREATE TABLE events (
    id BIGSERIAL,
    event_type VARCHAR(50),
    created_at TIMESTAMP NOT NULL,
    data JSONB
) PARTITION BY RANGE (created_at);

-- Create partitions
CREATE TABLE events_2024_01 PARTITION OF events
    FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');

CREATE TABLE events_2024_02 PARTITION OF events
    FOR VALUES FROM ('2024-02-01') TO ('2024-03-01');

-- Create indexes on partitions
CREATE INDEX idx_events_2024_01_type ON events_2024_01(event_type);
CREATE INDEX idx_events_2024_02_type ON events_2024_02(event_type);

-- Query uses partition pruning
EXPLAIN (ANALYZE)
SELECT * FROM events
WHERE created_at >= '2024-01-15' AND created_at < '2024-01-20';
-- Should show "Partitions pruned: X"
```

## Performance Monitoring

### Key Metrics Queries

```sql
-- pg_stat_statements (install extension first)
CREATE EXTENSION IF NOT EXISTS pg_stat_statements;

-- Top slow queries
SELECT
    round(total_exec_time::numeric, 2) as total_time,
    calls,
    round(mean_exec_time::numeric, 2) as mean_time,
    round((100 * total_exec_time / sum(total_exec_time) OVER ())::numeric, 2) as pct,
    query
FROM pg_stat_statements
ORDER BY total_exec_time DESC
LIMIT 10;

-- Cache hit ratio by table
SELECT
    schemaname,
    tablename,
    heap_blks_hit,
    heap_blks_read,
    round(100.0 * heap_blks_hit / NULLIF(heap_blks_hit + heap_blks_read, 0), 2) as cache_hit_pct
FROM pg_statio_user_tables
WHERE heap_blks_hit + heap_blks_read > 0
ORDER BY heap_blks_read DESC;

-- Index usage statistics
SELECT
    schemaname,
    tablename,
    indexname,
    idx_scan,
    idx_tup_read,
    idx_tup_fetch,
    pg_size_pretty(pg_relation_size(indexrelid)) as size
FROM pg_stat_user_indexes
ORDER BY idx_scan DESC;
```

## Configuration File Example

```ini
# postgresql.conf - Production optimized for 16GB RAM server

# Memory
shared_buffers = 4GB
effective_cache_size = 12GB
work_mem = 40MB
maintenance_work_mem = 2GB

# WAL
wal_buffers = 16MB
checkpoint_completion_target = 0.9
max_wal_size = 2GB

# Query Planner
default_statistics_target = 200
random_page_cost = 1.1  # SSD
effective_io_concurrency = 200  # SSD

# Parallel Queries
max_parallel_workers_per_gather = 4
max_parallel_workers = 8

# Connections
max_connections = 200

# Logging
log_min_duration_statement = 1000  # Log queries > 1s
log_line_prefix = '%t [%p]: [%l-1] user=%u,db=%d,app=%a,client=%h '
log_checkpoints = on
log_lock_waits = on
```

---

## Reference: query-optimization

# Query Optimization

## Execution Plan Analysis

### PostgreSQL EXPLAIN ANALYZE

```sql
-- Get actual execution statistics
EXPLAIN (ANALYZE, BUFFERS, VERBOSE, TIMING)
SELECT u.id, u.name, COUNT(o.id) as order_count
FROM users u
LEFT JOIN orders o ON u.id = o.user_id
WHERE u.created_at > NOW() - INTERVAL '30 days'
GROUP BY u.id, u.name
HAVING COUNT(o.id) > 5;

-- Key metrics to examine:
-- 1. Actual time vs Planning time
-- 2. Rows estimate vs Actual rows (cardinality)
-- 3. Buffers (shared hits vs reads)
-- 4. Sequential Scans vs Index Scans
-- 5. Join methods (Nested Loop, Hash Join, Merge Join)
```

### MySQL EXPLAIN

```sql
-- Basic execution plan
EXPLAIN SELECT * FROM orders
WHERE user_id = 123 AND status = 'pending';

-- JSON format for detailed analysis
EXPLAIN FORMAT=JSON
SELECT u.name, o.total
FROM users u
INNER JOIN orders o ON u.id = o.user_id
WHERE o.created_at > '2024-01-01';

-- Analyze actual execution (MySQL 8.0+)
EXPLAIN ANALYZE
SELECT * FROM products
WHERE category_id = 5
ORDER BY price DESC
LIMIT 10;
```

## Query Rewriting Patterns

### Eliminate Subqueries

```sql
-- BEFORE (Slow - executes subquery for each row)
SELECT *
FROM orders o
WHERE total > (
    SELECT AVG(total)
    FROM orders
    WHERE user_id = o.user_id
);

-- AFTER (Fast - single join with window function)
WITH user_averages AS (
    SELECT user_id, AVG(total) as avg_total
    FROM orders
    GROUP BY user_id
)
SELECT o.*
FROM orders o
INNER JOIN user_averages ua ON o.user_id = ua.user_id
WHERE o.total > ua.avg_total;
```

### Optimize JOIN Order

```sql
-- BEFORE (Cartesian product then filter)
SELECT p.name, c.name, s.stock
FROM products p, categories c, stock s
WHERE p.category_id = c.id
  AND p.id = s.product_id
  AND c.active = true;

-- AFTER (Filter first, then join)
SELECT p.name, c.name, s.stock
FROM categories c
INNER JOIN products p ON p.category_id = c.id
INNER JOIN stock s ON s.product_id = p.id
WHERE c.active = true;
```

### Use EXISTS Instead of IN

```sql
-- BEFORE (Slow - materializes entire subquery)
SELECT * FROM users
WHERE id IN (
    SELECT DISTINCT user_id
    FROM orders
    WHERE total > 1000
);

-- AFTER (Fast - short-circuits on first match)
SELECT * FROM users u
WHERE EXISTS (
    SELECT 1 FROM orders o
    WHERE o.user_id = u.id
    AND o.total > 1000
);
```

### Optimize DISTINCT

```sql
-- BEFORE (Sorts entire result set)
SELECT DISTINCT u.email
FROM users u
INNER JOIN orders o ON u.id = o.user_id
WHERE o.status = 'completed';

-- AFTER (Uses index for uniqueness)
SELECT u.email
FROM users u
WHERE EXISTS (
    SELECT 1 FROM orders o
    WHERE o.user_id = u.id
    AND o.status = 'completed'
);
```

## CTE Optimization

### Materialized vs Inline CTEs

```sql
-- PostgreSQL: Force materialization for reuse
WITH expensive_calculation AS MATERIALIZED (
    SELECT user_id,
           SUM(total) as lifetime_value,
           COUNT(*) as order_count
    FROM orders
    WHERE created_at > NOW() - INTERVAL '1 year'
    GROUP BY user_id
)
SELECT *
FROM expensive_calculation
WHERE lifetime_value > 10000
   OR order_count > 50;

-- Force inline for single-use CTEs
WITH recent_users AS NOT MATERIALIZED (
    SELECT id FROM users
    WHERE created_at > NOW() - INTERVAL '7 days'
)
SELECT * FROM recent_users;
```

## Window Function Optimization

```sql
-- BEFORE (Multiple subqueries)
SELECT
    o.id,
    o.total,
    (SELECT MAX(total) FROM orders WHERE user_id = o.user_id) as max_total,
    (SELECT AVG(total) FROM orders WHERE user_id = o.user_id) as avg_total
FROM orders o;

-- AFTER (Single window function scan)
SELECT
    id,
    total,
    MAX(total) OVER (PARTITION BY user_id) as max_total,
    AVG(total) OVER (PARTITION BY user_id) as avg_total
FROM orders;
```

## Aggregation Strategies

### Partial Aggregation

```sql
-- For large cardinality groups, pre-aggregate
WITH daily_stats AS (
    SELECT
        DATE(created_at) as day,
        user_id,
        COUNT(*) as daily_orders,
        SUM(total) as daily_total
    FROM orders
    WHERE created_at > NOW() - INTERVAL '90 days'
    GROUP BY DATE(created_at), user_id
)
SELECT
    user_id,
    SUM(daily_orders) as total_orders,
    AVG(daily_total) as avg_daily_total
FROM daily_stats
GROUP BY user_id;
```

## Pagination Optimization

```sql
-- BEFORE (Slow on large offsets)
SELECT * FROM products
ORDER BY created_at DESC
LIMIT 20 OFFSET 10000;

-- AFTER (Keyset pagination - cursor-based)
SELECT * FROM products
WHERE created_at < '2024-01-01 12:00:00'
   OR (created_at = '2024-01-01 12:00:00' AND id < 12345)
ORDER BY created_at DESC, id DESC
LIMIT 20;

-- Create index for keyset pagination
CREATE INDEX idx_products_pagination
ON products (created_at DESC, id DESC);
```

## Query Pattern Red Flags

| Pattern | Issue | Solution |
|---------|-------|----------|
| `SELECT *` | Fetches unnecessary columns | Select only needed columns |
| `OR` conditions | Prevents index usage | Use UNION or separate queries |
| `LIKE '%term%'` | Full table scan | Use full-text search or trigram indexes |
| `WHERE DATE(column) = ...` | Function prevents index usage | Use range: `column >= '2024-01-01' AND column < '2024-01-02'` |
| Large `IN` lists | Inefficient for >100 items | Use temporary table or JOIN |
| Implicit type conversion | Prevents index usage | Match column data types exactly |

## Performance Validation

```sql
-- PostgreSQL: Compare query performance
EXPLAIN (ANALYZE, BUFFERS)
-- your query here

-- Check buffer cache hits
SELECT
    sum(heap_blks_read) as heap_read,
    sum(heap_blks_hit) as heap_hit,
    sum(heap_blks_hit) / (sum(heap_blks_hit) + sum(heap_blks_read)) as ratio
FROM pg_statio_user_tables;

-- MySQL: Check handler statistics
SHOW STATUS LIKE 'Handler%';
FLUSH STATUS;
-- run your query
SHOW STATUS LIKE 'Handler%';
```
