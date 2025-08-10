# MongoDB Performance Report

## Generated: 2025-08-10 19:37:51

## Replica Set Status

**Replica Set Name**: rs0

| Node | State | Health |
|------|-------|--------|
| 63.xxx.xxx.2ff:50000 | PRIMARY | 1 |
| 18.xxx.xxx.e98:50000 | SECONDARY | 1 |
| 3.xxx.xxx.c6e:50000 | SECONDARY | 1 |

## Database Statistics

**Database**: production_test

- **Data Size**: 49.99 GB
- **Storage Size**: 20.14 GB
- **Index Size**: 21.16 GB
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
- **Document Count**: 185717197
- **Data Size**: 49.96 GB
- **Storage Size**: 20.13 GB
- **Avg Document Size**: 0 bytes
- **Indexes**:
  - _id_: map[_id:1]
  - unique_hash_1: map[unique_hash:1]
  - account_id_1: map[account_id:1]

## Performance Metrics

### Operation Counters (Cumulative)
- **Inserts**: 185904022
- **Queries**: 185693833
- **Updates**: 185719526
- **Deletes**: 159940

### Resource Usage
- **Cache Usage**: 12.78 GB / 16.00 GB (79.9%)
- **Connections**: 120

## Configuration

- **Cache Size**: 16 GB
- **Replica Set Settings**:
  - Election Timeout: 10000 ms
  - Heartbeat Interval: 2000 ms
