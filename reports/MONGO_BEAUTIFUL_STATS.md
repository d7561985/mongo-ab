# MongoDB Performance Report

## Generated: 2025-08-10 20:48:55

## Replica Set Status

**Replica Set Name**: rs0

| Node | State | Health |
|------|-------|--------|
| 63.xxx.xxx.2ff:50000 | PRIMARY | 1 |
| 18.xxx.xxx.e98:50000 | SECONDARY | 1 |
| 3.xxx.xxx.c6e:50000 | SECONDARY | 1 |

## Database Statistics

**Database**: production_test

- **Data Size**: 52.52 GB
- **Storage Size**: 21.19 GB
- **Index Size**: 22.06 GB
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
- **Document Count**: 195113187
- **Data Size**: 52.49 GB
- **Storage Size**: 21.17 GB
- **Avg Document Size**: 0 bytes
- **Indexes**:
  - _id_: map[_id:1]
  - unique_hash_1: map[unique_hash:1]
  - account_id_1: map[account_id:1]

## Performance Metrics

### Operation Counters (Cumulative)
- **Inserts**: 195300128
- **Queries**: 195089977
- **Updates**: 195117138
- **Deletes**: 159946

### Resource Usage
- **Cache Usage**: 12.67 GB / 16.00 GB (79.2%)
- **Connections**: 119

## Configuration

- **Cache Size**: 16 GB
- **Replica Set Settings**:
  - Election Timeout: 10000 ms
  - Heartbeat Interval: 2000 ms

## Live Performance Metrics

### Real-time Operations (5-second sample)

| Time | Insert | Query | Update | Delete | Total Ops/sec | Cache Used | Dirty | Network In | Network Out | Connections |
|------|--------|-------|--------|--------|--------------|------------|-------|------------|-------------|-------------|
| 20:48:58 | 2568 | 2570 | 2555 | 0 | **7693** | 79.4% (12.7GB/16.0GB) | 0.0% | 4.6 MB/s | 7.1 MB/s | 119 |
| 20:48:59 | 2376 | 2389 | 2386 | 0 | **7151** | 79.4% (12.7GB/16.0GB) | 0.0% | 4.3 MB/s | 6.6 MB/s | 119 |
| 20:49:00 | 2275 | 2274 | 2257 | 0 | **6806** | 79.5% (12.7GB/16.0GB) | 0.0% | 4.2 MB/s | 6.4 MB/s | 119 |
| 20:49:01 | 2379 | 2394 | 2413 | 0 | **7186** | 79.6% (12.7GB/16.0GB) | 0.0% | 4.3 MB/s | 6.5 MB/s | 119 |

### Key Metrics Analysis

* **Average Operations**: 7209 ops/sec
* **Average Cache Usage**: 79.4%
* **Average Network In**: 4.4 MB/s
* **Average Network Out**: 6.6 MB/s
* **Virtual Memory**: 16.1 GB
* **Resident Memory**: 14.0 GB
* **Active Clients**: 0 readers, 3 writers

