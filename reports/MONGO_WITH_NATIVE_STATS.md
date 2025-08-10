# MongoDB Performance Report

## Generated: 2025-08-10 20:43:44

## Replica Set Status

**Replica Set Name**: rs0

| Node | State | Health |
|------|-------|--------|
| 63.xxx.xxx.2ff:50000 | PRIMARY | 1 |
| 18.xxx.xxx.e98:50000 | SECONDARY | 1 |
| 3.xxx.xxx.c6e:50000 | SECONDARY | 1 |

## Database Statistics

**Database**: production_test

- **Data Size**: 52.33 GB
- **Storage Size**: 21.11 GB
- **Index Size**: 22.03 GB
- **Collections**: 2
- **Indexes**: 5

## Collections

### accounts
- **Document Count**: 180372
- **Data Size**: 0.00 GB
- **Storage Size**: 0.00 GB
- **Avg Document Size**: 0 bytes
- **Indexes**:
  - _id_: map[_id:1]
  - user_external_id_1: map[user_external_id:1]

### transactions
- **Document Count**: 194423751
- **Data Size**: 52.30 GB
- **Storage Size**: 21.10 GB
- **Avg Document Size**: 0 bytes
- **Indexes**:
  - _id_: map[_id:1]
  - unique_hash_1: map[unique_hash:1]
  - account_id_1: map[account_id:1]

## Performance Metrics

### Operation Counters (Cumulative)
- **Inserts**: 194610693
- **Queries**: 194400495
- **Updates**: 194427485
- **Deletes**: 159944

### Resource Usage
- **Cache Usage**: 12.78 GB / 16.00 GB (79.9%)
- **Connections**: 119

## Configuration

- **Cache Size**: 16 GB
- **Replica Set Settings**:
  - Election Timeout: 10000 ms
  - Heartbeat Interval: 2000 ms

## Live Performance Metrics (mongostat)

```
timestamp           insert query update delete getmore command dirty used flushes vsize   res qrw arw net_in net_out conn
20:43:46              2370  2381   2353      0       0     0|0   0.0% 79.7%       0 16.1G   14.1G 0|0 1|3 4MB    6MB      119
20:43:47              2475  2472   2481      0       0     0|0   0.0% 79.9%       0 16.1G   14.1G 0|0 2|3 4MB    6MB      119
20:43:48              2388  2396   2383      0       0     0|0   0.0% 79.9%       0 16.1G   14.1G 0|0 1|1 4MB    6MB      119
20:43:49              2731  2736   2826      1       0     0|0   0.0% 80.0%       0 16.1G   14.1G 0|0 1|3 4MB    7MB      119
```
