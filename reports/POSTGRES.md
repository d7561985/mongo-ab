# Postgres

config:
```
max_connections = 600
shared_buffers = 12GB
temp_buffers = 256MB
wal_level = replica
checkpoint_timeout = 15min # range 30s-1d
max_wal_size = 100GB
min_wal_size = 1GB
checkpoint_completion_target = 0.9
wal_keep_segments = 0
seq_page_cost = 1.0 # measured on an arbitrary scale
random_page_cost = 1.3 # we use io1, NVME
effective_cache_size = 36GB
default_statistics_target = 200
```

usage via sql:
```sql
SELECT *, pg_size_pretty(total_bytes) AS total
     , pg_size_pretty(index_bytes) AS index
     , pg_size_pretty(toast_bytes) AS toast
     , pg_size_pretty(table_bytes) AS table
FROM (
    SELECT *, total_bytes-index_bytes-coalesce(toast_bytes,0) AS table_bytes FROM (
    SELECT c.oid,nspname AS table_schema, relname AS table_name
        , c.reltuples AS row_estimate
        , pg_size_pretty(c.oid) AS total_bytes
        , pg_indexes_size(c.oid) AS index_bytes
        , pg_total_relation_size(reltoastrelid) AS toast_bytes
    FROM pg_class c
    LEFT JOIN pg_namespace n ON n.oid = c.relnamespace
    WHERE relkind = 'r'
    ) a
    ) a;
```

```json
[
  {
    "oid": 16398,
    "table_schema": "public",
    "table_name": "balance",
    "row_estimate": 100000,
    "total_bytes": 9969664,
    "index_bytes": 2768896,
    "toast_bytes": null,
    "table_bytes": 7200768,
    "total": "9736 kB",
    "index": "2704 kB",
    "toast": null,
    "table": "7032 kB"
  },
  {
    "oid": 16403,
    "table_schema": "public",
    "table_name": "journal",
    "row_estimate": 33506400,
    "total_bytes": 8184045568,
    "index_bytes": 1336508416,
    "toast_bytes": 8192,
    "table_bytes": 6847528960,
    "total": "7805 MB",
    "index": "1275 MB",
    "toast": "8192 bytes",
    "table": "6530 MB"
  }
]
```

SO: 8184045568 / 33506400 = `244.253204403` average