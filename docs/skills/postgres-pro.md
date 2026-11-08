# postgres-pro

> Источник: https://github.com/jeffallan/claude-skills/tree/main/skills/postgres-pro (MIT). Рабочая копия скилла: `.claude/skills/postgres-pro/`.

---
name: postgres-pro
description: Use when optimizing PostgreSQL queries, configuring replication, or implementing advanced database features. Invoke for EXPLAIN analysis, JSONB operations, extension usage, VACUUM tuning, performance monitoring.
license: MIT
metadata:
  author: https://github.com/Jeffallan
  company: https://synergetic.solutions
  version: "1.1.0"
  domain: infrastructure
  triggers: PostgreSQL, Postgres, EXPLAIN ANALYZE, pg_stat, JSONB, streaming replication, logical replication, VACUUM, PostGIS, pgvector
  role: specialist
  scope: implementation
  output-format: code
  related-skills: database-optimizer, devops-engineer, sre-engineer
---

# PostgreSQL Pro

Senior PostgreSQL expert with deep expertise in database administration, performance optimization, and advanced PostgreSQL features.

## When to Use This Skill

- Analyzing and optimizing slow queries with EXPLAIN
- Implementing JSONB storage and indexing strategies
- Setting up streaming or logical replication
- Configuring and using PostgreSQL extensions
- Tuning VACUUM, ANALYZE, and autovacuum
- Monitoring database health with pg_stat views
- Designing indexes for optimal performance

## Core Workflow

1. **Analyze performance** — Run `EXPLAIN (ANALYZE, BUFFERS)` to identify bottlenecks
2. **Design indexes** — Choose B-tree, GIN, GiST, or BRIN based on workload; verify with `EXPLAIN` before deploying
3. **Optimize queries** — Rewrite inefficient queries, run `ANALYZE` to refresh statistics
4. **Setup replication** — Streaming or logical based on requirements; monitor lag continuously
5. **Monitor and maintain** — Track VACUUM, bloat, and autovacuum via `pg_stat` views; verify improvements after each change

### End-to-End Example: Slow Query → Fix → Verification

```sql
-- Step 1: Identify slow queries
SELECT query, mean_exec_time, calls
FROM pg_stat_statements
ORDER BY mean_exec_time DESC
LIMIT 10;

-- Step 2: Analyze a specific slow query
EXPLAIN (ANALYZE, BUFFERS, FORMAT TEXT)
SELECT * FROM orders WHERE customer_id = 42 AND status = 'pending';
-- Look for: Seq Scan (bad on large tables), high Buffers hit, nested loops on large sets

-- Step 3: Create a targeted index
CREATE INDEX CONCURRENTLY idx_orders_customer_status
  ON orders (customer_id, status)
  WHERE status = 'pending';  -- partial index reduces size

-- Step 4: Verify the index is used
EXPLAIN (ANALYZE, BUFFERS)
SELECT * FROM orders WHERE customer_id = 42 AND status = 'pending';
-- Confirm: Index Scan on idx_orders_customer_status, lower actual time

-- Step 5: Update statistics if needed after bulk changes
ANALYZE orders;
```

## Reference Guide

Load detailed guidance based on context:

| Topic | Reference | Load When |
|-------|-----------|-----------|
| Performance | `references/performance.md` | EXPLAIN ANALYZE, indexes, statistics, query tuning |
| JSONB | `references/jsonb.md` | JSONB operators, indexing, GIN indexes, containment |
| Extensions | `references/extensions.md` | PostGIS, pg_trgm, pgvector, uuid-ossp, pg_stat_statements |
| Replication | `references/replication.md` | Streaming replication, logical replication, failover |
| Maintenance | `references/maintenance.md` | VACUUM, ANALYZE, pg_stat views, monitoring, bloat |

## Common Patterns

### JSONB — GIN Index and Query

```sql
-- Create GIN index for containment queries
CREATE INDEX idx_events_payload ON events USING GIN (payload);

-- Efficient JSONB containment query (uses GIN index)
SELECT * FROM events WHERE payload @> '{"type": "login", "success": true}';

-- Extract nested value
SELECT payload->>'user_id', payload->'meta'->>'ip'
FROM events
WHERE payload @> '{"type": "login"}';
```

### VACUUM and Bloat Monitoring

```sql
-- Check tables with high dead tuple counts
SELECT relname, n_dead_tup, n_live_tup,
       round(n_dead_tup::numeric / NULLIF(n_live_tup + n_dead_tup, 0) * 100, 2) AS dead_pct,
       last_autovacuum
FROM pg_stat_user_tables
ORDER BY n_dead_tup DESC
LIMIT 20;

-- Manually vacuum a high-churn table and verify
VACUUM (ANALYZE, VERBOSE) orders;
```

### Replication Lag Monitoring

```sql
-- On primary: check standby lag
SELECT client_addr, state, sent_lsn, write_lsn, flush_lsn, replay_lsn,
       (sent_lsn - replay_lsn) AS replication_lag_bytes
FROM pg_stat_replication;
```

## Constraints

### MUST DO
- Use `EXPLAIN (ANALYZE, BUFFERS)` for query optimization
- Verify indexes are actually used with `EXPLAIN` before and after creation
- Use `CREATE INDEX CONCURRENTLY` to avoid table locks in production
- Run `ANALYZE` after bulk data changes to refresh statistics
- Monitor autovacuum; tune `autovacuum_vacuum_scale_factor` for high-churn tables
- Use connection pooling (pgBouncer, pgPool)
- Monitor replication lag via `pg_stat_replication`
- Use prepared statements to prevent SQL injection
- Use `uuid` type for UUIDs, not `text`

### MUST NOT DO
- Disable autovacuum globally
- Create indexes without first analyzing query patterns
- Use `SELECT *` in production queries
- Ignore replication lag alerts
- Skip VACUUM on high-churn tables
- Store large BLOBs in the database (use object storage)
- Deploy index changes without verifying the planner uses them

## Output Templates

When implementing PostgreSQL solutions, provide:
1. Query with `EXPLAIN (ANALYZE, BUFFERS)` output and interpretation
2. Index definitions with rationale and pre/post verification
3. Configuration changes with before/after values
4. Monitoring queries for ongoing health checks
5. Brief explanation of performance impact

## Knowledge Reference

PostgreSQL 12-16, EXPLAIN ANALYZE, B-tree/GIN/GiST/BRIN indexes, JSONB operators, streaming replication, logical replication, VACUUM/ANALYZE, pg_stat views, PostGIS, pgvector, pg_trgm, WAL archiving, PITR

Maintained by [@jeffallan](https://github.com/jeffallan), Principal Consultant at [Synergetic Solutions](https://synergetic.solutions)

[Documentation](https://jeffallan.github.io/claude-skills/skills/infrastructure/postgres-pro/)

---

## Reference: extensions

# PostgreSQL Extensions

## Extension Management

```sql
-- List available extensions
SELECT * FROM pg_available_extensions ORDER BY name;

-- List installed extensions
SELECT * FROM pg_extension;

-- Install extension
CREATE EXTENSION IF NOT EXISTS pg_stat_statements;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Drop extension
DROP EXTENSION pg_stat_statements;

-- Update extension
ALTER EXTENSION pg_stat_statements UPDATE TO '1.10';
```

## pg_stat_statements (Query Performance)

```sql
-- Install and configure
CREATE EXTENSION pg_stat_statements;

-- postgresql.conf:
-- shared_preload_libraries = 'pg_stat_statements'
-- pg_stat_statements.max = 10000
-- pg_stat_statements.track = all

-- Top 10 slowest queries by mean time
SELECT
  query,
  calls,
  total_exec_time,
  mean_exec_time,
  max_exec_time,
  stddev_exec_time,
  rows
FROM pg_stat_statements
ORDER BY mean_exec_time DESC
LIMIT 10;

-- Most frequently called queries
SELECT
  query,
  calls,
  total_exec_time,
  mean_exec_time
FROM pg_stat_statements
ORDER BY calls DESC
LIMIT 10;

-- Most time-consuming queries (total time)
SELECT
  query,
  calls,
  total_exec_time / 1000 as total_seconds,
  mean_exec_time,
  (total_exec_time / sum(total_exec_time) OVER ()) * 100 as percentage
FROM pg_stat_statements
ORDER BY total_exec_time DESC
LIMIT 10;

-- Reset statistics
SELECT pg_stat_statements_reset();
```

## uuid-ossp (UUID Generation)

```sql
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Generate UUIDs
SELECT uuid_generate_v1();    -- Time-based + MAC address
SELECT uuid_generate_v4();    -- Random (most common)

-- Use in tables
CREATE TABLE users (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  email TEXT NOT NULL,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Insert
INSERT INTO users (email) VALUES ('user@example.com')
RETURNING id;
```

## pg_trgm (Fuzzy String Matching)

```sql
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Similarity search
SELECT
  email,
  similarity(email, 'john@example.com') as sim
FROM users
WHERE similarity(email, 'john@example.com') > 0.3
ORDER BY sim DESC;

-- LIKE optimization with trigram index
CREATE INDEX idx_users_email_trgm ON users USING GIN(email gin_trgm_ops);

-- Now these queries use index:
SELECT * FROM users WHERE email ILIKE '%john%';
SELECT * FROM users WHERE email % 'jon@example.com';  -- Similar to

-- Trigram operators
SELECT 'hello' % 'helo';              -- True (similar)
SELECT similarity('hello', 'helo');   -- 0.5
SELECT word_similarity('hello', 'hello world');  -- 1.0

-- Set similarity threshold
SET pg_trgm.similarity_threshold = 0.5;
SELECT * FROM users WHERE email % 'searchtext';
```

## PostGIS (Spatial and Geographic)

```sql
CREATE EXTENSION IF NOT EXISTS postgis;

-- Create spatial table
CREATE TABLE locations (
  id SERIAL PRIMARY KEY,
  name TEXT NOT NULL,
  geom GEOMETRY(Point, 4326)  -- WGS84 lat/lng
);

-- Add spatial index
CREATE INDEX idx_locations_geom ON locations USING GIST(geom);

-- Insert point (longitude, latitude)
INSERT INTO locations (name, geom)
VALUES ('NYC', ST_SetSRID(ST_MakePoint(-74.0060, 40.7128), 4326));

-- Distance queries (in meters)
SELECT
  name,
  ST_Distance(
    geom::geography,
    ST_SetSRID(ST_MakePoint(-73.9857, 40.7484), 4326)::geography
  ) as distance_meters
FROM locations
ORDER BY distance_meters
LIMIT 10;

-- Within radius (1km = 1000m)
SELECT * FROM locations
WHERE ST_DWithin(
  geom::geography,
  ST_SetSRID(ST_MakePoint(-74.0060, 40.7128), 4326)::geography,
  1000
);

-- Bounding box query (very fast with GIST index)
SELECT * FROM locations
WHERE geom && ST_MakeEnvelope(-74.1, 40.6, -73.9, 40.8, 4326);

-- Contains query
SELECT * FROM zones
WHERE ST_Contains(geom, ST_SetSRID(ST_MakePoint(-74.0060, 40.7128), 4326));

-- Area calculation
SELECT
  name,
  ST_Area(geom::geography) / 1000000 as area_km2
FROM zones;

-- GeoJSON export
SELECT
  name,
  ST_AsGeoJSON(geom) as geojson
FROM locations;
```

## pgvector (Vector Similarity Search)

```sql
CREATE EXTENSION IF NOT EXISTS vector;

-- Create table with vector column
CREATE TABLE embeddings (
  id SERIAL PRIMARY KEY,
  content TEXT,
  embedding vector(1536)  -- OpenAI embeddings are 1536 dimensions
);

-- Add vector index (HNSW for better performance)
CREATE INDEX ON embeddings USING hnsw (embedding vector_cosine_ops);
-- or IVFFlat for memory efficiency:
-- CREATE INDEX ON embeddings USING ivfflat (embedding vector_cosine_ops);

-- Insert vectors
INSERT INTO embeddings (content, embedding)
VALUES ('Hello world', '[0.1, 0.2, 0.3, ...]');

-- Similarity search (cosine distance)
SELECT
  content,
  1 - (embedding <=> '[0.1, 0.2, ...]') as similarity
FROM embeddings
ORDER BY embedding <=> '[0.1, 0.2, ...]'
LIMIT 10;

-- Distance operators
-- <-> L2 distance (Euclidean)
-- <#> negative inner product
-- <=> cosine distance (most common for embeddings)

-- Set index parameters for better recall
SET hnsw.ef_search = 100;  -- Higher = better recall, slower query

-- Bulk insert optimization
SET maintenance_work_mem = '2GB';
```

## pgcrypto (Encryption and Hashing)

```sql
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Hash passwords (use bcrypt)
INSERT INTO users (email, password_hash)
VALUES ('user@example.com', crypt('<password>', gen_salt('bf', 10)));

-- Verify password
SELECT * FROM users
WHERE email = 'user@example.com'
  AND password_hash = crypt('<password>', password_hash);

-- Generate random values
SELECT gen_random_uuid();
SELECT gen_random_bytes(32);

-- Encrypt/decrypt data
SELECT
  pgp_sym_encrypt('sensitive data', 'encryption-key'),
  pgp_sym_decrypt(encrypted_column, 'encryption-key')
FROM table_name;

-- Digest functions
SELECT digest('data', 'sha256');
SELECT encode(digest('data', 'sha256'), 'hex');
```

## postgres_fdw (Foreign Data Wrapper)

```sql
CREATE EXTENSION IF NOT EXISTS postgres_fdw;

-- Create foreign server
CREATE SERVER remote_db
FOREIGN DATA WRAPPER postgres_fdw
OPTIONS (host 'remote-host', port '5432', dbname 'remote_db');

-- User mapping
CREATE USER MAPPING FOR current_user
SERVER remote_db
OPTIONS (user 'remote_user', password '<remote-db-password>');

-- Import foreign schema
IMPORT FOREIGN SCHEMA public
FROM SERVER remote_db
INTO remote_schema;

-- Or create specific foreign table
CREATE FOREIGN TABLE remote_users (
  id INTEGER,
  email TEXT,
  created_at TIMESTAMPTZ
)
SERVER remote_db
OPTIONS (schema_name 'public', table_name 'users');

-- Query remote table (transparent)
SELECT * FROM remote_users WHERE created_at > NOW() - INTERVAL '1 day';

-- Join local and remote tables
SELECT
  l.id,
  l.name,
  r.email
FROM local_table l
JOIN remote_users r ON l.user_id = r.id;
```

## pg_repack (Online Table Reorganization)

```sql
CREATE EXTENSION IF NOT EXISTS pg_repack;

-- Repack table (removes bloat, rebuilds indexes)
-- Run via command line, not SQL:
-- pg_repack -d mydb -t users

-- Check bloat before repack
SELECT
  schemaname,
  tablename,
  pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as size,
  pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename) -
                 pg_relation_size(schemaname||'.'||tablename)) as index_size
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;

-- Repack entire database
-- pg_repack -d mydb

-- Repack with custom order
-- pg_repack -d mydb -t users -o "created_at DESC"
```

## timescaledb (Time-Series Data)

```sql
CREATE EXTENSION IF NOT EXISTS timescaledb;

-- Create hypertable (must have time column)
CREATE TABLE metrics (
  time TIMESTAMPTZ NOT NULL,
  device_id INTEGER,
  temperature DOUBLE PRECISION,
  humidity DOUBLE PRECISION
);

-- Convert to hypertable
SELECT create_hypertable('metrics', 'time');

-- Set chunk interval (default 7 days)
SELECT set_chunk_time_interval('metrics', INTERVAL '1 day');

-- Add compression
ALTER TABLE metrics SET (
  timescaledb.compress,
  timescaledb.compress_segmentby = 'device_id',
  timescaledb.compress_orderby = 'time DESC'
);

-- Automatic compression policy (compress chunks older than 7 days)
SELECT add_compression_policy('metrics', INTERVAL '7 days');

-- Retention policy (drop chunks older than 30 days)
SELECT add_retention_policy('metrics', INTERVAL '30 days');

-- Continuous aggregates (materialized views for time-series)
CREATE MATERIALIZED VIEW metrics_hourly
WITH (timescaledb.continuous) AS
SELECT
  time_bucket('1 hour', time) AS bucket,
  device_id,
  AVG(temperature) as avg_temp,
  MAX(temperature) as max_temp,
  MIN(temperature) as min_temp
FROM metrics
GROUP BY bucket, device_id;

-- Refresh policy
SELECT add_continuous_aggregate_policy('metrics_hourly',
  start_offset => INTERVAL '3 hours',
  end_offset => INTERVAL '1 hour',
  schedule_interval => INTERVAL '1 hour'
);
```

## Extension Recommendations by Use Case

**Query Performance Monitoring:**
- `pg_stat_statements` (essential)
- `pg_stat_kcache` (cache hit statistics)

**Text Search:**
- `pg_trgm` (fuzzy matching, LIKE optimization)
- Built-in full-text search (no extension needed)

**Spatial Data:**
- `postgis` (comprehensive spatial features)

**Vector Embeddings / AI:**
- `pgvector` (for semantic search, RAG applications)

**Time-Series:**
- `timescaledb` (automatic partitioning, compression)

**Data Security:**
- `pgcrypto` (hashing, encryption)
- `pg_audit` (audit logging)

**UUID Support:**
- `uuid-ossp` (UUID generation)

**Cross-Database Queries:**
- `postgres_fdw` (query remote PostgreSQL)
- `file_fdw` (query CSV files)

**Table Maintenance:**
- `pg_repack` (online bloat removal)

---

## Reference: jsonb

# JSONB Operations

## JSONB vs JSON

```sql
-- Use JSONB (binary, indexed, faster)
CREATE TABLE documents (
  id SERIAL PRIMARY KEY,
  data JSONB NOT NULL
);

-- NOT json (text storage, no indexing)
-- Only use json if you need to preserve exact formatting/whitespace
```

## JSONB Operators

### Retrieval Operators

```sql
-- -> returns JSONB
SELECT data -> 'user' FROM documents;                    -- {"id": 123, "name": "Alice"}
SELECT data -> 'user' -> 'name' FROM documents;          -- "Alice" (still JSONB)

-- ->> returns text
SELECT data ->> 'status' FROM documents;                 -- active (text)
SELECT data -> 'user' ->> 'name' FROM documents;         -- Alice (text)

-- #> for nested paths (JSONB)
SELECT data #> '{user,address,city}' FROM documents;    -- "NYC" (JSONB)

-- #>> for nested paths (text)
SELECT data #>> '{user,address,city}' FROM documents;   -- NYC (text)

-- Array access
SELECT data -> 'tags' -> 0 FROM documents;               -- First tag
SELECT jsonb_array_elements(data -> 'tags') FROM documents;  -- Expand array
```

### Containment Operators

```sql
-- @> contains (most useful for indexing)
SELECT * FROM documents WHERE data @> '{"status": "active"}';
SELECT * FROM documents WHERE data @> '{"tags": ["postgresql"]}';
SELECT * FROM documents WHERE data -> 'user' @> '{"role": "admin"}';

-- <@ is contained by
SELECT * FROM documents WHERE '{"status": "active"}' <@ data;

-- ? key exists
SELECT * FROM documents WHERE data ? 'email';
SELECT * FROM documents WHERE data -> 'user' ? 'email';

-- ?| any key exists
SELECT * FROM documents WHERE data ?| ARRAY['email', 'phone'];

-- ?& all keys exist
SELECT * FROM documents WHERE data ?& ARRAY['email', 'phone'];
```

### Modification Operators

```sql
-- || concatenate/merge (shallow)
UPDATE documents SET data = data || '{"updated_at": "2024-01-01"}'::jsonb;

-- - remove key
UPDATE documents SET data = data - 'temp_field';

-- #- remove nested path
UPDATE documents SET data = data #- '{user,temp_field}';

-- jsonb_set for deep updates
UPDATE documents
SET data = jsonb_set(data, '{user,email}', '"new@example.com"'::jsonb)
WHERE id = 123;

-- jsonb_insert
UPDATE documents
SET data = jsonb_insert(data, '{tags,0}', '"new-tag"'::jsonb)
WHERE id = 123;
```

## JSONB Indexing

### GIN Index (Default for containment)

```sql
-- Standard GIN index (for @>, ?, ?&, ?| operators)
CREATE INDEX idx_documents_data ON documents USING GIN(data);

-- Queries that benefit:
SELECT * FROM documents WHERE data @> '{"status": "active"}';
SELECT * FROM documents WHERE data ? 'email';
SELECT * FROM documents WHERE data ?& ARRAY['email', 'phone'];
```

### GIN Index on Specific Path

```sql
-- Index specific path for better performance
CREATE INDEX idx_documents_status ON documents USING GIN((data -> 'status'));
CREATE INDEX idx_documents_user ON documents USING GIN((data -> 'user'));

-- Smaller index, faster queries on specific paths
SELECT * FROM documents WHERE data -> 'status' @> '"active"';
```

### GIN Index with jsonb_path_ops

```sql
-- Smaller, faster index for @> queries only
CREATE INDEX idx_documents_path_ops ON documents USING GIN(data jsonb_path_ops);

-- Good for: WHERE data @> '{"key": "value"}'
-- Bad for: WHERE data ? 'key' (not supported)
-- ~20% smaller than default GIN, faster for containment
```

### B-tree Index on Extracted Values

```sql
-- Index extracted value (most selective)
CREATE INDEX idx_documents_status_btree ON documents((data ->> 'status'));
CREATE INDEX idx_documents_user_id ON documents((CAST(data -> 'user' ->> 'id' AS INTEGER)));

-- Enables efficient equality and range queries
SELECT * FROM documents WHERE data ->> 'status' = 'active';
SELECT * FROM documents WHERE CAST(data -> 'user' ->> 'id' AS INTEGER) > 1000;
```

### Expression Index for Nested Values

```sql
-- Index deep nested value
CREATE INDEX idx_documents_user_email ON documents((data #>> '{user,email}'));

-- Enables:
SELECT * FROM documents WHERE data #>> '{user,email}' = 'user@example.com';
```

## Query Patterns

### Filtering

```sql
-- Exact match
SELECT * FROM documents WHERE data @> '{"status": "active"}';

-- Multiple conditions
SELECT * FROM documents
WHERE data @> '{"status": "active", "verified": true}';

-- Nested conditions
SELECT * FROM documents
WHERE data -> 'user' @> '{"role": "admin"}';

-- Array containment
SELECT * FROM documents
WHERE data -> 'tags' @> '["postgresql"]';

-- Text search in JSONB value
SELECT * FROM documents
WHERE data ->> 'title' ILIKE '%postgres%';
```

### Aggregation

```sql
-- Extract and aggregate
SELECT
  data ->> 'status' as status,
  COUNT(*) as count,
  AVG(CAST(data ->> 'score' AS FLOAT)) as avg_score
FROM documents
GROUP BY data ->> 'status';

-- Array aggregation
SELECT
  jsonb_agg(data -> 'user') as users
FROM documents
WHERE data @> '{"status": "active"}';

-- Object aggregation
SELECT
  jsonb_object_agg(id, data -> 'user') as user_map
FROM documents
WHERE data ? 'user';
```

### Array Operations

```sql
-- Expand array to rows
SELECT
  id,
  jsonb_array_elements(data -> 'tags') as tag
FROM documents;

-- Expand array to text
SELECT
  id,
  jsonb_array_elements_text(data -> 'tags') as tag
FROM documents;

-- Array length
SELECT * FROM documents
WHERE jsonb_array_length(data -> 'tags') > 5;

-- Filter array elements
SELECT
  id,
  jsonb_path_query_array(data, '$.tags[*] ? (@ like_regex "^post.*" flag "i")') as postgres_tags
FROM documents;
```

## JSONB Functions

```sql
-- Build JSONB
SELECT jsonb_build_object('id', 123, 'name', 'Alice', 'active', true);
SELECT jsonb_build_array(1, 2, 'three', true);

-- Object keys
SELECT jsonb_object_keys(data) FROM documents;

-- Pretty print
SELECT jsonb_pretty(data) FROM documents;

-- Type checking
SELECT jsonb_typeof(data -> 'score');  -- number, string, array, object, boolean, null

-- Strip nulls
SELECT jsonb_strip_nulls(data) FROM documents;
```

## JSONB Path Queries (Postgres 12+)

```sql
-- jsonb_path_query for flexible queries
SELECT jsonb_path_query(data, '$.user.address.city') FROM documents;

-- With filters
SELECT jsonb_path_query(data, '$.items[*] ? (@.price > 100)') FROM documents;

-- Exists check
SELECT * FROM documents
WHERE jsonb_path_exists(data, '$.tags[*] ? (@ == "postgresql")');

-- Array result
SELECT jsonb_path_query_array(data, '$.items[*].name') FROM documents;
```

## Performance Best Practices

### DO

```sql
-- Use specific path indexes for hot paths
CREATE INDEX idx_docs_status ON documents((data ->> 'status'));

-- Use GIN index with path ops for containment-only queries
CREATE INDEX idx_docs_pathops ON documents USING GIN(data jsonb_path_ops);

-- Extract frequently queried values to columns
ALTER TABLE documents ADD COLUMN status TEXT GENERATED ALWAYS AS (data ->> 'status') STORED;
CREATE INDEX idx_docs_status_col ON documents(status);

-- Use @> for indexed queries
WHERE data @> '{"status": "active"}'  -- Fast with GIN index
```

### DON'T

```sql
-- Don't use ->> with @> (mixing types)
WHERE data @> '{"score": "100"}'  -- Wrong, comparing string
WHERE CAST(data ->> 'score' AS INTEGER) = 100  -- Better

-- Don't query without indexes
SELECT * FROM documents WHERE data -> 'nested' -> 'deep' ->> 'value' = 'x';
-- Add index: CREATE INDEX ON documents((data #>> '{nested,deep,value}'));

-- Don't store huge arrays in JSONB
-- If you have 10k+ elements, use a separate table

-- Don't use JSONB for high-update columns
-- Extract to regular column if updated frequently
```

## Schema Validation (Postgres 15+)

```sql
-- Using CHECK constraints
ALTER TABLE documents
ADD CONSTRAINT check_data_schema
CHECK (
  jsonb_typeof(data) = 'object' AND
  data ? 'id' AND
  data ? 'status' AND
  data ->> 'status' IN ('active', 'pending', 'archived')
);
```

## Migration Patterns

```sql
-- Add JSONB column
ALTER TABLE users ADD COLUMN metadata JSONB DEFAULT '{}'::jsonb;

-- Migrate existing columns to JSONB
UPDATE users SET metadata = jsonb_build_object(
  'preferences', preferences,
  'settings', settings,
  'flags', flags
);

-- Drop old columns after validation
ALTER TABLE users DROP COLUMN preferences, DROP COLUMN settings, DROP COLUMN flags;
```

---

## Reference: maintenance

# Database Maintenance

## VACUUM Fundamentals

### Why VACUUM is Critical

PostgreSQL uses MVCC (Multi-Version Concurrency Control):
- Updates/deletes don't remove old rows immediately
- Old rows marked as "dead tuples"
- VACUUM reclaims space from dead tuples
- Without VACUUM: table bloat, degraded performance, transaction ID wraparound

### VACUUM Variants

```sql
-- Standard VACUUM (non-blocking, reclaims space for reuse)
VACUUM users;
VACUUM;  -- All tables

-- VACUUM FULL (locks table, rewrites entire table, reclaims disk space)
VACUUM FULL users;
-- Use pg_repack instead for production (non-blocking alternative)

-- VACUUM VERBOSE (shows details)
VACUUM VERBOSE users;

-- VACUUM ANALYZE (vacuum + update statistics)
VACUUM ANALYZE users;
```

### VACUUM Monitoring

```sql
-- Check when tables were last vacuumed
SELECT
  schemaname,
  relname,
  last_vacuum,
  last_autovacuum,
  n_dead_tup,
  n_live_tup,
  round(100.0 * n_dead_tup / NULLIF(n_live_tup + n_dead_tup, 0), 2) as dead_pct
FROM pg_stat_user_tables
ORDER BY n_dead_tup DESC;

-- Check vacuum progress (PG 9.6+)
SELECT
  pid,
  datname,
  relid::regclass,
  phase,
  heap_blks_total,
  heap_blks_scanned,
  heap_blks_vacuumed,
  round(100.0 * heap_blks_scanned / NULLIF(heap_blks_total, 0), 2) as pct_complete
FROM pg_stat_progress_vacuum;
```

## Autovacuum Configuration

```sql
-- Global settings (postgresql.conf)
autovacuum = on
autovacuum_max_workers = 3
autovacuum_naptime = 60s  -- Check interval

-- Vacuum thresholds
autovacuum_vacuum_threshold = 50
autovacuum_vacuum_scale_factor = 0.2
-- Triggers when: dead_tuples > threshold + (scale_factor * total_tuples)
-- Default: 50 + (0.2 * 1000000) = 200,050 dead tuples for 1M row table

-- Analyze thresholds
autovacuum_analyze_threshold = 50
autovacuum_analyze_scale_factor = 0.1

-- Performance settings
autovacuum_vacuum_cost_delay = 2ms  -- Lower = faster, more I/O impact
autovacuum_vacuum_cost_limit = 200
```

### Per-Table Autovacuum Tuning

```sql
-- High-churn table: vacuum more aggressively
ALTER TABLE orders SET (
  autovacuum_vacuum_scale_factor = 0.05,  -- 5% instead of 20%
  autovacuum_vacuum_threshold = 1000,
  autovacuum_analyze_scale_factor = 0.02
);

-- Large, stable table: vacuum less often
ALTER TABLE archive_logs SET (
  autovacuum_vacuum_scale_factor = 0.5,
  autovacuum_vacuum_threshold = 5000
);

-- Very high-churn table: disable cost delays
ALTER TABLE sessions SET (
  autovacuum_vacuum_cost_delay = 0
);

-- View table settings
SELECT
  relname,
  reloptions
FROM pg_class
WHERE relname = 'orders';
```

## ANALYZE (Statistics)

```sql
-- Update statistics for query planner
ANALYZE users;
ANALYZE;  -- All tables

-- Check statistics freshness
SELECT
  schemaname,
  relname,
  last_analyze,
  last_autoanalyze,
  n_mod_since_analyze
FROM pg_stat_user_tables
ORDER BY n_mod_since_analyze DESC;

-- Increase statistics target for high-cardinality columns
ALTER TABLE users ALTER COLUMN email SET STATISTICS 1000;
-- Default is 100, range is 0-10000
-- Higher = better estimates, slower ANALYZE

-- View column statistics
SELECT
  tablename,
  attname,
  n_distinct,      -- Estimated unique values
  correlation,     -- Physical vs logical ordering (-1 to 1)
  null_frac        -- Percentage of nulls
FROM pg_stats
WHERE tablename = 'users';
```

## Bloat Detection and Removal

### Detect Table Bloat

```sql
-- Approximate bloat calculation
SELECT
  schemaname,
  relname,
  pg_size_pretty(pg_total_relation_size(format('%I.%I', schemaname, relname))) as total_size,
  pg_size_pretty(pg_relation_size(format('%I.%I', schemaname, relname))) as table_size,
  round(100 * pg_relation_size(format('%I.%I', schemaname, relname))::numeric /
        NULLIF(pg_total_relation_size(format('%I.%I', schemaname, relname)), 0), 2) as table_pct,
  n_dead_tup,
  round(100.0 * n_dead_tup / NULLIF(n_live_tup + n_dead_tup, 0), 2) as dead_pct
FROM pg_stat_user_tables
WHERE pg_total_relation_size(format('%I.%I', schemaname, relname)) > 10485760  -- > 10MB
ORDER BY pg_total_relation_size(format('%I.%I', schemaname, relname)) DESC;
```

### Detect Index Bloat

```sql
-- Unused indexes
SELECT
  schemaname,
  relname,
  indexrelname,
  idx_scan,
  pg_size_pretty(pg_relation_size(indexrelid)) as index_size
FROM pg_stat_user_indexes
WHERE idx_scan = 0
  AND indexrelname NOT LIKE '%pkey'
ORDER BY pg_relation_size(indexrelid) DESC;

-- Index size vs table size
SELECT
  schemaname,
  relname,
  pg_size_pretty(pg_relation_size(format('%I.%I', schemaname, relname))) as table_size,
  pg_size_pretty(pg_total_relation_size(format('%I.%I', schemaname, relname)) -
                 pg_relation_size(format('%I.%I', schemaname, relname))) as indexes_size,
  round(100.0 * (pg_total_relation_size(format('%I.%I', schemaname, relname)) -
                 pg_relation_size(format('%I.%I', schemaname, relname)))::numeric /
        NULLIF(pg_relation_size(format('%I.%I', schemaname, relname)), 0), 2) as index_ratio_pct
FROM pg_stat_user_tables
WHERE pg_total_relation_size(format('%I.%I', schemaname, relname)) > 10485760
ORDER BY pg_total_relation_size(format('%I.%I', schemaname, relname)) DESC;
```

### Remove Bloat

```sql
-- Option 1: VACUUM FULL (locks table)
VACUUM FULL users;

-- Option 2: pg_repack (online, no locks)
-- Command line: pg_repack -d mydb -t users

-- Option 3: REINDEX (for index bloat)
REINDEX TABLE users;
REINDEX INDEX CONCURRENTLY idx_users_email;  -- Non-blocking (PG 12+)

-- Option 4: CLUSTER (rewrite table in index order, locks table)
CLUSTER users USING users_pkey;
```

## pg_stat Monitoring Views

### pg_stat_activity (Current Queries)

```sql
-- Active queries
SELECT
  pid,
  usename,
  application_name,
  client_addr,
  state,
  query_start,
  state_change,
  query
FROM pg_stat_activity
WHERE state = 'active'
  AND query NOT LIKE '%pg_stat_activity%'
ORDER BY query_start;

-- Long-running queries
SELECT
  pid,
  now() - query_start as duration,
  state,
  query
FROM pg_stat_activity
WHERE state = 'active'
  AND (now() - query_start) > interval '5 minutes'
ORDER BY duration DESC;

-- Kill long-running query
SELECT pg_cancel_backend(pid);  -- Graceful
SELECT pg_terminate_backend(pid);  -- Forceful

-- Idle transactions (bad, hold locks)
SELECT
  pid,
  usename,
  state,
  now() - state_change as idle_duration,
  query
FROM pg_stat_activity
WHERE state = 'idle in transaction'
  AND (now() - state_change) > interval '1 minute';
```

### pg_stat_database (Database-wide Stats)

```sql
SELECT
  datname,
  numbackends,  -- Active connections
  xact_commit,
  xact_rollback,
  round(100.0 * xact_rollback / NULLIF(xact_commit + xact_rollback, 0), 2) as rollback_pct,
  blks_read,
  blks_hit,
  round(100.0 * blks_hit / NULLIF(blks_hit + blks_read, 0), 2) as cache_hit_ratio,
  tup_returned,
  tup_fetched,
  tup_inserted,
  tup_updated,
  tup_deleted
FROM pg_stat_database
WHERE datname = current_database();
```

### pg_stat_user_tables (Table Stats)

```sql
SELECT
  schemaname,
  relname,
  seq_scan,        -- Sequential scans (high = may need index)
  seq_tup_read,
  idx_scan,        -- Index scans
  idx_tup_fetch,
  n_tup_ins,
  n_tup_upd,
  n_tup_del,
  n_tup_hot_upd,   -- HOT updates (good, in-page updates)
  n_live_tup,
  n_dead_tup,
  last_vacuum,
  last_autovacuum,
  last_analyze,
  last_autoanalyze
FROM pg_stat_user_tables
ORDER BY seq_scan DESC;  -- Tables with most sequential scans
```

### pg_stat_user_indexes (Index Usage)

```sql
-- Index usage efficiency
SELECT
  schemaname,
  relname,
  indexrelname,
  idx_scan,
  idx_tup_read,
  idx_tup_fetch,
  pg_size_pretty(pg_relation_size(indexrelid)) as index_size
FROM pg_stat_user_indexes
ORDER BY idx_scan;  -- Low idx_scan = potentially unused index

-- Index hit ratio
SELECT
  schemaname,
  relname,
  indexrelname,
  idx_scan,
  idx_tup_read,
  idx_tup_fetch,
  CASE WHEN idx_tup_read > 0
    THEN round(100.0 * idx_tup_fetch / idx_tup_read, 2)
    ELSE 0
  END as hit_ratio
FROM pg_stat_user_indexes
WHERE idx_scan > 0
ORDER BY hit_ratio;
```

### pg_statio_user_tables (I/O Stats)

```sql
SELECT
  schemaname,
  relname,
  heap_blks_read,   -- Disk reads
  heap_blks_hit,    -- Cache hits
  round(100.0 * heap_blks_hit / NULLIF(heap_blks_hit + heap_blks_read, 0), 2) as cache_hit_ratio,
  idx_blks_read,
  idx_blks_hit,
  toast_blks_read,
  toast_blks_hit
FROM pg_statio_user_tables
WHERE heap_blks_read + heap_blks_hit > 0
ORDER BY heap_blks_read DESC;
```

## Lock Monitoring

```sql
-- Current locks
SELECT
  l.pid,
  a.usename,
  a.query,
  l.mode,
  l.locktype,
  l.granted,
  l.relation::regclass
FROM pg_locks l
JOIN pg_stat_activity a ON l.pid = a.pid
WHERE NOT l.granted
ORDER BY l.pid;

-- Blocking queries
SELECT
  blocked_locks.pid AS blocked_pid,
  blocked_activity.usename AS blocked_user,
  blocking_locks.pid AS blocking_pid,
  blocking_activity.usename AS blocking_user,
  blocked_activity.query AS blocked_statement,
  blocking_activity.query AS blocking_statement
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
```

## Transaction ID Wraparound

```sql
-- Check distance to wraparound (should be < 1 billion)
SELECT
  datname,
  age(datfrozenxid) as xid_age,
  2147483647 - age(datfrozenxid) as xids_remaining
FROM pg_database
ORDER BY age(datfrozenxid) DESC;

-- Per-table wraparound status
SELECT
  schemaname,
  relname,
  age(relfrozenxid) as xid_age,
  pg_size_pretty(pg_total_relation_size(schemaname||'.'||relname)) as size
FROM pg_stat_user_tables
ORDER BY age(relfrozenxid) DESC
LIMIT 20;

-- Prevent wraparound: VACUUM FREEZE
VACUUM FREEZE;  -- Current database
VACUUM FREEZE users;  -- Specific table
```

## Maintenance Checklist

**Daily:**
- Monitor autovacuum activity
- Check for long-running queries
- Verify replication lag (if applicable)
- Check cache hit ratio

**Weekly:**
- Review slow queries from pg_stat_statements
- Check for table/index bloat
- Review unused indexes
- Monitor disk space usage

**Monthly:**
- Review autovacuum settings
- Reindex heavily updated indexes
- Update statistics on large tables
- Review database growth trends

**Quarterly:**
- Test backup restoration
- Review and optimize slow queries
- Capacity planning
- PostgreSQL version updates

## Helpful Maintenance Queries

```sql
-- Database size
SELECT
  pg_database.datname,
  pg_size_pretty(pg_database_size(pg_database.datname)) as size
FROM pg_database
ORDER BY pg_database_size(pg_database.datname) DESC;

-- Largest tables
SELECT
  schemaname,
  tablename,
  pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as total_size,
  pg_size_pretty(pg_relation_size(schemaname||'.'||tablename)) as table_size,
  pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename) -
                 pg_relation_size(schemaname||'.'||tablename)) as index_size
FROM pg_tables
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC
LIMIT 20;

-- Connection count by state
SELECT
  state,
  count(*) as count
FROM pg_stat_activity
GROUP BY state
ORDER BY count DESC;

-- Reset statistics (after performance testing)
SELECT pg_stat_reset();
SELECT pg_stat_statements_reset();
```

---

## Reference: performance

# Performance Optimization

## EXPLAIN ANALYZE Fundamentals

```sql
-- Basic EXPLAIN ANALYZE
EXPLAIN (ANALYZE, BUFFERS, VERBOSE)
SELECT u.id, u.name, COUNT(o.id) as order_count
FROM users u
LEFT JOIN orders o ON u.id = o.user_id
WHERE u.created_at > '2024-01-01'
GROUP BY u.id, u.name;

-- Key metrics to watch:
-- Planning Time: Time spent creating query plan
-- Execution Time: Actual query execution time
-- Shared Hit Blocks: Data found in cache (good)
-- Shared Read Blocks: Data read from disk (slow)
-- Rows: Estimated vs actual row counts
```

## Reading EXPLAIN Output

```
Seq Scan on users  (cost=0.00..1234.56 rows=10000 width=32)
                    ^^^^^^^^^^^^^^^^^^^^  ^^^^^^     ^^^^^^^^
                    startup..total cost   estimate   row width

Actual time: 0.123..45.678 rows=9876 loops=1
             ^^^^^^^^^^^^^^^  ^^^^^^^^  ^^^^^^^
             first..last row  actual    iterations
```

**Node types (fastest to slowest):**
- Index Only Scan - Best, data from index only
- Index Scan - Good, uses index + heap lookup
- Bitmap Index Scan - Good for multiple conditions
- Seq Scan - Table scan, OK for small tables
- Seq Scan on large table - Problem, needs index

## Index Strategies

### B-tree Indexes (Default)

```sql
-- Single column index
CREATE INDEX idx_users_email ON users(email);

-- Multi-column index (order matters!)
CREATE INDEX idx_orders_user_date ON orders(user_id, created_at DESC);
-- Good for: WHERE user_id = X ORDER BY created_at DESC
-- Good for: WHERE user_id = X AND created_at > Y
-- Bad for: WHERE created_at > Y (doesn't use index)

-- Partial index (smaller, faster)
CREATE INDEX idx_active_users ON users(email) WHERE active = true;

-- Expression index
CREATE INDEX idx_users_lower_email ON users(LOWER(email));
-- Enables: WHERE LOWER(email) = 'user@example.com'

-- Covering index (includes extra columns)
CREATE INDEX idx_orders_covering ON orders(user_id) INCLUDE (total, created_at);
-- Enables Index Only Scan
```

### GIN Indexes (JSONB, arrays, full-text)

```sql
-- JSONB containment
CREATE INDEX idx_data_gin ON documents USING GIN(data);
-- Enables: WHERE data @> '{"status": "active"}'

-- JSONB specific paths
CREATE INDEX idx_data_status ON documents USING GIN((data -> 'status'));

-- Array operations
CREATE INDEX idx_tags_gin ON posts USING GIN(tags);
-- Enables: WHERE tags @> ARRAY['postgresql', 'performance']

-- Full-text search
CREATE INDEX idx_content_fts ON articles USING GIN(to_tsvector('english', content));
-- Enables: WHERE to_tsvector('english', content) @@ to_tsquery('postgresql & performance')
```

### GiST Indexes (Spatial, ranges, nearest neighbor)

```sql
-- PostGIS spatial index
CREATE INDEX idx_locations_geom ON locations USING GIST(geom);
-- Enables: WHERE ST_DWithin(geom, point, 1000)

-- Range types
CREATE INDEX idx_bookings_range ON bookings USING GIST(during);
-- Enables: WHERE during && '[2024-01-01, 2024-01-31]'::daterange

-- Nearest neighbor (KNN)
CREATE INDEX idx_locations_gist ON locations USING GIST(coordinates);
-- Enables: ORDER BY coordinates <-> point('0,0') LIMIT 10
```

### BRIN Indexes (Large, naturally ordered tables)

```sql
-- Time-series data (insert-only, sorted by time)
CREATE INDEX idx_metrics_time_brin ON metrics USING BRIN(timestamp);
-- Very small index, good for WHERE timestamp > NOW() - INTERVAL '1 day'

-- Works well with:
-- - Log tables
-- - Time-series metrics
-- - Append-only tables with natural order
```

## Statistics and Planner

```sql
-- Update statistics (do after bulk changes)
ANALYZE users;
ANALYZE;  -- All tables

-- Check statistics freshness
SELECT schemaname, relname, last_analyze, last_autoanalyze
FROM pg_stat_user_tables
WHERE schemaname = 'public';

-- Increase statistics target for high-cardinality columns
ALTER TABLE users ALTER COLUMN email SET STATISTICS 1000;
-- Default is 100, increase for better selectivity estimates

-- View column statistics
SELECT * FROM pg_stats WHERE tablename = 'users' AND attname = 'email';
```

## Query Optimization Patterns

### Problem: Sequential scan on large table

```sql
-- Bad: Full table scan
SELECT * FROM orders WHERE user_id = 123;
-- Solution: Add index
CREATE INDEX idx_orders_user ON orders(user_id);
```

### Problem: Index not used

```sql
-- Bad: Function prevents index usage
SELECT * FROM users WHERE LOWER(email) = 'user@example.com';
-- Solution: Expression index
CREATE INDEX idx_users_email_lower ON users(LOWER(email));

-- Bad: Implicit type conversion
SELECT * FROM users WHERE id = '123';  -- id is integer
-- Solution: Use correct type
SELECT * FROM users WHERE id = 123;
```

### Problem: Large JOIN inefficiency

```sql
-- Bad: Nested loop on large tables
EXPLAIN ANALYZE
SELECT * FROM orders o JOIN users u ON o.user_id = u.id;

-- Solutions:
-- 1. Ensure indexes exist on join columns
CREATE INDEX idx_orders_user ON orders(user_id);
-- 2. Update statistics
ANALYZE orders, users;
-- 3. Increase work_mem if hash join would be better
SET work_mem = '256MB';
```

### Problem: COUNT(*) slow

```sql
-- Bad: Full table scan
SELECT COUNT(*) FROM orders WHERE status = 'pending';

-- Solutions:
-- 1. Partial index
CREATE INDEX idx_orders_pending ON orders(id) WHERE status = 'pending';

-- 2. Approximate count for large tables
SELECT reltuples::bigint FROM pg_class WHERE relname = 'orders';

-- 3. Materialized count for reports
CREATE MATERIALIZED VIEW order_counts AS
SELECT status, COUNT(*) FROM orders GROUP BY status;
CREATE UNIQUE INDEX ON order_counts(status);
REFRESH MATERIALIZED VIEW CONCURRENTLY order_counts;
```

## Connection Pooling

```sql
-- Check active connections
SELECT count(*) FROM pg_stat_activity WHERE state = 'active';

-- Connection limit reached? Use pgBouncer
-- pgbouncer.ini:
-- [databases]
-- mydb = host=localhost port=5432 dbname=mydb
-- [pgbouncer]
-- pool_mode = transaction
-- max_client_conn = 1000
-- default_pool_size = 25
```

## Configuration Tuning

```sql
-- Memory settings (for 16GB RAM server)
shared_buffers = 4GB           -- 25% of RAM
effective_cache_size = 12GB    -- 75% of RAM
work_mem = 64MB                -- Per operation
maintenance_work_mem = 1GB     -- For VACUUM, CREATE INDEX

-- Checkpoint tuning
checkpoint_completion_target = 0.9
wal_buffers = 16MB
checkpoint_timeout = 10min

-- Query planner
random_page_cost = 1.1         -- Lower for SSD (default 4.0)
effective_io_concurrency = 200 -- Higher for SSD

-- Parallelism (Postgres 10+)
max_parallel_workers_per_gather = 4
max_parallel_workers = 8
```

## Performance Monitoring

```sql
-- Slow queries (requires pg_stat_statements)
SELECT
  query,
  calls,
  mean_exec_time,
  max_exec_time,
  stddev_exec_time
FROM pg_stat_statements
ORDER BY mean_exec_time DESC
LIMIT 20;

-- Cache hit ratio (should be > 99%)
SELECT
  sum(blks_hit) * 100.0 / sum(blks_hit + blks_read) as cache_hit_ratio
FROM pg_stat_database;

-- Index usage
SELECT
  schemaname,
  relname,
  indexrelname,
  idx_scan,
  idx_tup_read,
  idx_tup_fetch
FROM pg_stat_user_indexes
WHERE idx_scan = 0
  AND indexrelname NOT LIKE '%pkey';  -- Unused indexes
```

---

## Reference: replication

# PostgreSQL Replication

## Streaming Replication (Physical)

### Primary Server Setup

```sql
-- postgresql.conf
wal_level = replica
max_wal_senders = 10
max_replication_slots = 10
wal_keep_size = 1GB  # Or 1024MB for older versions
hot_standby = on
archive_mode = on
archive_command = 'cp %p /var/lib/postgresql/wal_archive/%f'

-- pg_hba.conf (allow replication connections)
host replication replicator 10.0.0.0/24 scram-sha-256
```

```sql
-- Create replication user
CREATE ROLE replicator WITH REPLICATION LOGIN PASSWORD '<replication-password>';

-- Create replication slot (prevents WAL deletion)
SELECT * FROM pg_create_physical_replication_slot('replica_1');
```

### Standby Server Setup

```bash
# Stop PostgreSQL on standby
systemctl stop postgresql

# Remove data directory
rm -rf /var/lib/postgresql/14/main/*

# Base backup from primary
pg_basebackup -h primary-host -D /var/lib/postgresql/14/main \
  -U replicator -P -v -R -X stream -S replica_1

# -R creates standby.signal and recovery config
# -X stream: stream WAL during backup
# -S replica_1: use replication slot
```

```sql
-- standby.signal file created by pg_basebackup -R
-- recovery parameters in postgresql.auto.conf:
primary_conninfo = 'host=primary-host port=5432 user=replicator password=<replication-password>'
primary_slot_name = 'replica_1'
```

### Monitoring Replication

```sql
-- On primary: Check replication status
SELECT
  client_addr,
  state,
  sync_state,
  sent_lsn,
  write_lsn,
  flush_lsn,
  replay_lsn,
  pg_wal_lsn_diff(sent_lsn, replay_lsn) as lag_bytes
FROM pg_stat_replication;

-- On standby: Check replay lag
SELECT
  now() - pg_last_xact_replay_timestamp() AS replication_lag;

-- Check replication slots
SELECT
  slot_name,
  slot_type,
  active,
  restart_lsn,
  pg_wal_lsn_diff(pg_current_wal_lsn(), restart_lsn) as retained_bytes
FROM pg_replication_slots;
```

### Synchronous Replication

```sql
-- postgresql.conf on primary
synchronous_commit = on
synchronous_standby_names = 'FIRST 1 (replica_1, replica_2)'
# Waits for 1 standby to confirm before commit

# Options:
# FIRST n (names): Wait for n standbys
# ANY n (names): Wait for any n standbys
# name: Wait for specific standby

-- Query to check sync status
SELECT
  application_name,
  sync_state,
  state
FROM pg_stat_replication;
-- sync_state: sync (synchronous), async, potential
```

## Logical Replication (Row-level)

### Publisher Setup

```sql
-- postgresql.conf
wal_level = logical
max_replication_slots = 10
max_wal_senders = 10

-- Create publication (all tables)
CREATE PUBLICATION my_publication FOR ALL TABLES;

-- Or specific tables
CREATE PUBLICATION my_publication FOR TABLE users, orders;

-- Or tables matching pattern (PG15+)
CREATE PUBLICATION my_publication FOR TABLES IN SCHEMA public;

-- With row filters (PG15+)
CREATE PUBLICATION active_users FOR TABLE users WHERE (active = true);

-- View publications
SELECT * FROM pg_publication;
SELECT * FROM pg_publication_tables;
```

### Subscriber Setup

```sql
-- Create subscription (creates replication slot on publisher)
CREATE SUBSCRIPTION my_subscription
CONNECTION 'host=publisher-host port=5432 dbname=mydb user=replicator password=<replication-password>'
PUBLICATION my_publication;

-- Subscription options
CREATE SUBSCRIPTION my_subscription
CONNECTION 'host=publisher-host dbname=mydb user=replicator'
PUBLICATION my_publication
WITH (
  copy_data = true,           -- Initial data copy
  create_slot = true,          -- Create replication slot
  enabled = true,              -- Start immediately
  slot_name = 'my_sub_slot',
  synchronous_commit = 'off'   -- Performance vs durability
);

-- View subscriptions
SELECT * FROM pg_subscription;
SELECT * FROM pg_stat_subscription;

-- Manage subscription
ALTER SUBSCRIPTION my_subscription DISABLE;
ALTER SUBSCRIPTION my_subscription ENABLE;
ALTER SUBSCRIPTION my_subscription REFRESH PUBLICATION;
DROP SUBSCRIPTION my_subscription;
```

### Logical Replication Monitoring

```sql
-- On publisher: Check replication slots
SELECT
  slot_name,
  plugin,
  slot_type,
  active,
  pg_wal_lsn_diff(pg_current_wal_lsn(), confirmed_flush_lsn) as lag_bytes
FROM pg_replication_slots
WHERE slot_type = 'logical';

-- On subscriber: Check subscription status
SELECT
  subname,
  pid,
  received_lsn,
  latest_end_lsn,
  last_msg_send_time,
  last_msg_receipt_time,
  latest_end_time
FROM pg_stat_subscription;
```

## Cascading Replication

```
Primary -> Standby1 -> Standby2
```

```sql
-- On Standby1 (acts as relay)
-- postgresql.conf
hot_standby = on
max_wal_senders = 10
wal_keep_size = 1GB

-- Standby2 connects to Standby1
-- Same setup as regular standby, but primary_conninfo points to Standby1
primary_conninfo = 'host=standby1-host user=replicator...'
```

## Delayed Replication (Delayed Standby)

```sql
-- On standby: postgresql.conf
recovery_min_apply_delay = '4h'

-- Useful for:
-- - Protection against accidental data deletion
-- - Rolling back to specific point in time
-- - Can promote delayed standby to recover dropped table

-- Check delay
SELECT now() - pg_last_xact_replay_timestamp() AS current_delay;
```

## Failover and Promotion

### Manual Failover

```bash
# On standby server
# Promote standby to primary
pg_ctl promote -D /var/lib/postgresql/14/main

# Or use SQL
SELECT pg_promote();

# Verify promotion
SELECT pg_is_in_recovery();  -- Should return false
```

### Automatic Failover with pg_auto_failover

```bash
# Install pg_auto_failover
apt-get install pg-auto-failover

# Setup monitor node
pg_autoctl create monitor --hostname monitor-host --pgdata /var/lib/monitor

# Setup primary
pg_autoctl create postgres \
  --hostname primary-host \
  --pgdata /var/lib/postgresql/14/main \
  --monitor postgres://monitor-host/pg_auto_failover

# Setup standby
pg_autoctl create postgres \
  --hostname standby-host \
  --pgdata /var/lib/postgresql/14/main \
  --monitor postgres://monitor-host/pg_auto_failover

# Check status
pg_autoctl show state
```

### Patroni (Production HA Solution)

```yaml
# patroni.yml
scope: postgres-cluster
name: node1

restapi:
  listen: 0.0.0.0:8008
  connect_address: node1:8008

etcd:
  hosts: etcd1:2379,etcd2:2379,etcd3:2379

bootstrap:
  dcs:
    ttl: 30
    loop_wait: 10
    retry_timeout: 10
    maximum_lag_on_failover: 1048576
    postgresql:
      use_pg_rewind: true
      parameters:
        max_connections: 100
        max_wal_senders: 10
        wal_level: replica

postgresql:
  listen: 0.0.0.0:5432
  connect_address: node1:5432
  data_dir: /var/lib/postgresql/14/main
  authentication:
    replication:
      username: replicator
      password: <replication-password>
    superuser:
      username: postgres
      password: <superuser-password>
```

## Connection Pooling for HA

### PgBouncer Configuration

```ini
# pgbouncer.ini
[databases]
mydb = host=primary-host port=5432 dbname=mydb

[pgbouncer]
listen_addr = *
listen_port = 6432
auth_type = scram-sha-256
auth_file = /etc/pgbouncer/userlist.txt
pool_mode = transaction
max_client_conn = 1000
default_pool_size = 25
reserve_pool_size = 5
```

### HAProxy for Load Balancing

```
# haproxy.cfg
frontend postgres_frontend
    bind *:5432
    mode tcp
    default_backend postgres_backend

backend postgres_backend
    mode tcp
    option tcp-check
    tcp-check expect string is_master:true

    server primary primary-host:5432 check
    server standby1 standby1-host:5432 check backup
    server standby2 standby2-host:5432 check backup
```

## Backup and Point-in-Time Recovery (PITR)

### WAL Archiving Setup

```sql
-- postgresql.conf
wal_level = replica
archive_mode = on
archive_command = 'test ! -f /backup/wal/%f && cp %p /backup/wal/%f'
archive_timeout = 300  # Force archive every 5 minutes

-- Or use pg_archivecleanup
archive_command = 'pgbackrest --stanza=main archive-push %p'
```

### Base Backup with pg_basebackup

```bash
# Full backup
pg_basebackup -h localhost -U postgres \
  -D /backup/base/$(date +%Y%m%d) \
  -Ft -z -P -X fetch

# -Ft: tar format
# -z: gzip compression
# -P: progress
# -X fetch: include WAL files
```

### Point-in-Time Recovery

```bash
# Stop PostgreSQL
systemctl stop postgresql

# Restore base backup
rm -rf /var/lib/postgresql/14/main/*
tar -xzf /backup/base/20241201/base.tar.gz -C /var/lib/postgresql/14/main

# Create recovery.signal
touch /var/lib/postgresql/14/main/recovery.signal

# Configure recovery
# postgresql.conf or postgresql.auto.conf:
restore_command = 'cp /backup/wal/%f %p'
recovery_target_time = '2024-12-01 14:30:00'
# Or: recovery_target_xid, recovery_target_name, recovery_target_lsn

# Start PostgreSQL (will recover to target)
systemctl start postgresql

# After recovery, check
SELECT pg_is_in_recovery();  # Should be false after recovery completes
```

## Monitoring Best Practices

```sql
-- Create monitoring view
CREATE VIEW replication_status AS
SELECT
  client_addr,
  application_name,
  state,
  sync_state,
  pg_wal_lsn_diff(pg_current_wal_lsn(), replay_lsn) / 1024 / 1024 AS lag_mb,
  (pg_wal_lsn_diff(pg_current_wal_lsn(), replay_lsn)::float /
   (1024 * 1024 * 16))::int AS estimated_wal_segments_behind
FROM pg_stat_replication;

-- Alert if lag > 100MB
SELECT * FROM replication_status WHERE lag_mb > 100;

-- Check replication slot disk usage
SELECT
  slot_name,
  pg_size_pretty(
    pg_wal_lsn_diff(pg_current_wal_lsn(), restart_lsn)
  ) as retained_wal
FROM pg_replication_slots;
```

## Troubleshooting

```sql
-- Replication broken?
-- 1. Check pg_stat_replication on primary
SELECT * FROM pg_stat_replication;

-- 2. Check logs on standby
-- tail -f /var/log/postgresql/postgresql-14-main.log

-- 3. Check replication slot exists
SELECT * FROM pg_replication_slots WHERE slot_name = 'replica_1';

-- 4. Recreate slot if missing
SELECT pg_create_physical_replication_slot('replica_1');

-- 5. Check WAL files available
-- ls -lh /var/lib/postgresql/14/main/pg_wal/

-- Standby too far behind?
-- Option 1: Increase wal_keep_size
-- Option 2: Use replication slots
-- Option 3: Re-run pg_basebackup
```
