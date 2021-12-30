# Reports

## AWS

## Test1

### topology:
* mongod:  x3 r5d.xlarge: 32 GiB of memory, 4 vCPUs, 1 x 150 NVMe SSD, 64-bit platform
* mongos:  x1 t3.small: t3.small, 2 GiB of Memory, 2 vCPUs, EBS only, 64-bit platform
* config:  x1 x1 t3.small: t3.small, 2 GiB of Memory, 2 vCPUs, EBS only, 64-bit platform 

### config mongod:
link to SSD
storage.dbPath: /data/mongodb

storage.directoryPerDB: true
storage.journal.enabled: true
storage.journal.commitIntervalMs: 100
storage.wiredTiger.cacheSizeGB: 10
storage.wiredTiger.directoryForIndexes: true

### indexes:
```bash
[direct: mongos] db> db.balance.getIndexes()
[
  { v: 2, key: { _id: 1 }, name: '_id_' },
  { v: 2, key: { _id: 'hashed' }, name: '_id_hashed' }
]
```
```bash
[direct: mongos] db> db.journal.getIndexes()
[
  { v: 2, key: { _id: 1 }, name: '_id_' },
  { v: 2, key: { accountId: 'hashed' }, name: 'accountId_hashed' },
  { v: 2, key: { accountId: 1 }, name: 'accountId_1' }
]
```

#### go-driver
```go
	clientOpts := options.Client().ApplyURI(addr).
		SetWriteConcern(writeconcern.New(
			writeconcern.WTimeout(timeout),
			writeconcern.J(false),
		)).SetRetryWrites(false).
		SetCompressors([]string{"zlib"}).
		SetZlibLevel(9)
```

#### outcome 
collection count:
```bash
[direct: mongos] db> db.journal.estimatedDocumentCount()
314124917
[direct: mongos] db> db.balance.count({})
100000
```

#### stats
in progress

```bash
insert query update delete getmore command flushes mapped vsize  res faults qrw arw net_in net_out conn                time
  1706    *0     *0     *0       0  5110|0       0     0B 1.77G 126M      0 0|0 0|0  2.46m   1.46m  212 Dec 26 11:59:20.703
  1680    *0     *0     *0       0  5071|0       0     0B 1.77G 127M      0 0|0 0|0  2.45m   1.44m  212 Dec 26 11:59:21.703
  1618    *0     *0     *0       0  4873|0       0     0B 1.77G 126M      0 0|0 0|0  2.35m   1.39m  212 Dec 26 11:59:24.731
  1705    *0     *0     *0       0  5134|0       0     0B 1.77G 126M      0 0|0 0|0  2.48m   1.47m  212 Dec 26 11:59:25.715
  1682    *0     *0     *0       0  5050|0       0     0B 1.77G 125M      0 0|0 0|0  2.44m   1.44m  212 Dec 26 11:59:26.699
```

#### disc usage:
```bash
[ec2-user@ip-172-31-22-180 ~]$ sudo du -s /data
34241868/data

> mongosh
use db;
use db.journal.stats()
  size: Long("41557982228"),
  count: 103999035,
  avgObjSize: 399,
  storageSize: Long("19791413248"),
  freeStorageSize: 417792,
  capped: false,

41557982228 > 38,70388700440526

[ec2-user@ip-172-31-16-144 ~]$ sudo du -s /data
34819912/data
> mongosh
  ns: 'db.journal',
  size: Long("42367986503"),
  count: 106025831,
  avgObjSize: 399,
  storageSize: Long("20179197952"),
  freeStorageSize: 2568192,
  capped: false,
  
[ec2-user@ip-172-31-30-188 ~]$ sudo du -s /data
34230000/data

> mongosh
  ns: 'db.journal',
  size: Long("41598551721"),
  count: 104100051,
  avgObjSize: 399,
  storageSize: Long("19811229696"),
  freeStorageSize: 1265664,
  capped: false,
```

#### sh.status
```bash
[direct: mongos] db> sh.status()
shardingVersion
{
  _id: 1,
  minCompatibleVersion: 5,
  currentVersion: 6,
  clusterId: ObjectId("61c787049821922b5906f59c")
}
---
shards
[
  {
    _id: 'rs1',
    host: 'rs1/3.70.203.181:50000',
    state: 1,
    topologyTime: Timestamp({ t: 1640466188, i: 3 })
  },
  {
    _id: 'rs2',
    host: 'rs2/18.159.60.246:50000',
    state: 1,
    topologyTime: Timestamp({ t: 1640466188, i: 8 })
  },
  {
    _id: 'rs3',
    host: 'rs3/3.125.18.58:50000',
    state: 1,
    topologyTime: Timestamp({ t: 1640466188, i: 13 })
  }
]
---
active mongoses
[ { '5.0.5': 1 } ]
---
autosplit
{ 'Currently enabled': 'yes' }
---
balancer
{
  'Currently enabled': 'yes',
  'Currently running': 'no',
  'Failed balancer rounds in last 5 attempts': 0,
  'Migration Results for the last 24 hours': { '682': 'Success' }
}
---
databases
[
  {
    database: { _id: 'config', primary: 'config', partitioned: true },
    collections: {
      'config.system.sessions': {
        shardKey: { _id: 1 },
        unique: false,
        balancing: true,
        chunkMetadata: [
          { shard: 'rs1', nChunks: 342 },
          { shard: 'rs2', nChunks: 341 },
          { shard: 'rs3', nChunks: 341 }
        ],
        chunks: [
          'too many chunks to print, use verbose if you want to force print'
        ],
        tags: []
      }
    }
  },
  {
    database: {
      _id: 'db',
      primary: 'rs3',
      partitioned: true,
      version: {
        uuid: UUID("0e6eecbe-b746-43ba-9505-a7e5a518f0cd"),
        timestamp: Timestamp({ t: 1640466188, i: 15 }),
        lastMod: 1
      }
    },
    collections: {
      'db.balance': {
        shardKey: { _id: 'hashed' },
        unique: false,
        balancing: true,
        chunkMetadata: [
          { shard: 'rs1', nChunks: 8191 },
          { shard: 'rs2', nChunks: 8191 },
          { shard: 'rs3', nChunks: 8191 }
        ],
        chunks: [
          'too many chunks to print, use verbose if you want to force print'
        ],
        tags: []
      },
      'db.journal': {
        shardKey: { accountId: 'hashed' },
        unique: false,
        balancing: true,
        chunkMetadata: [
          { shard: 'rs1', nChunks: 8191 },
          { shard: 'rs2', nChunks: 8191 },
          { shard: 'rs3', nChunks: 8191 }
        ],
        chunks: [
          'too many chunks to print, use verbose if you want to force print'
        ],
        tags: []
      }
    }
  }
]
```

## Test2

Update from test1:
* mongos configuration up to `c5.large`
* config mongod `storage.wiredTiger.cacheSizeGB up to 25

### topology:
* mongod:  x3 r5d.xlarge: 32 GiB of memory, 4 vCPUs, 1 x 150 NVMe SSD, 64-bit platform
* mongos:  x1 c5.large: c6i.large, 4 GiB of Memory, 2 vCPUs, EBS only, 64-bit platform
* config:  x1 x1 t3.small: t3.small, 2 GiB of Memory, 2 vCPUs, EBS only, 64-bit platform 

### config mongod:
link to SSD
storage.dbPath: /data/mongodb

storage.directoryPerDB: true
storage.journal.enabled: true
storage.journal.commitIntervalMs: 100
storage.wiredTiger.cacheSizeGB: 25
storage.wiredTiger.directoryForIndexes: true

### Stop
reason: c5.large not enough, CPU 100%

INSERTS:
```
  8646    *0     *0     *0       0  8659|0       0     0B 1.60G 72.0M      0 0|0 0|0  5.86m   1.49m  105 Dec 28 16:15:28.837
  8597    *0     *0     *0       0  8600|0       0     0B 1.60G 72.0M      0 0|0 0|0  5.82m   1.48m  105 Dec 28 16:15:29.842
  8577    *0     *0     *0       0  8554|0       0     0B 1.60G 72.0M      0 0|0 0|0  5.79m   1.48m  105 Dec 28 16:15:30.826
```

TX:
```
  2636    *0     *0     *0       0  7905|0       0     0B 1.61G 92.0M      0 0|0 0|0  3.82m   2.23m  105 Dec 28 16:17:15.010
  2576    *0     *0     *0       0  7752|0       0     0B 1.61G 92.0M      0 0|0 0|0  3.74m   2.18m  105 Dec 28 16:17:16.012
  2577    *0     *0     *0       0  7729|0       0     0B 1.61G 92.0M      0 0|0 0|0  3.73m   2.18m  105 Dec 28 16:17:17.008
  2585    *0     *0     *0       0  7761|0       0     0B 1.61G 92.0M      0 0|0 0|0  3.74m   2.19m  105 Dec 28 16:17:18.012
  2580    *0     *0     *0       0  7757|0       0     0B 1.61G 92.0M      0 0|0 0|0  3.74m   2.18m  105 Dec 28 16:17:19.016
  2594    *0     *0     *0       0  7796|0       0     0B 1.61G 92.0M      0 0|0 0|0  3.76m   2.20m  105 Dec 28 16:17:20.012
```

## Test3
Update from test1:
* mongos configuration up to `c5.2xlarge`

### topology:
* mongod:  x3 r5d.xlarge: 32 GiB of memory, 4 vCPUs, 1 x 150 NVMe SSD, 64-bit platform
* mongos:  x1 c5.2xlarge: 8 GiB of Memory, 4 vCPUs, EBS only, 64-bit platform
* config:  x1 x1 t3.small: t3.small, 2 GiB of Memory, 2 vCPUs, EBS only, 64-bit platform 

### Stop
reason: c5.2xlarge not enough, CPU 100%

INSERT:
```
 28379    *0     *0     *0       0 28396|0       0     0B 1.59G 79.0M      0 0|0 0|0  19.2m   4.84m  105 Dec 28 16:38:26.855
 28419    *0     *0     *0       0 28432|0       0     0B 1.59G 79.0M      0 0|0 0|0  19.2m   4.85m  105 Dec 28 16:38:27.854
 21677    *0     *0     *0       0 21679|0       0     0B 1.59G 79.0M      0 0|0 0|0  14.7m   3.70m  105 Dec 28 16:38:28.867
 28612    *0     *0     *0       0 28599|0       0     0B 1.59G 79.0M      0 0|0 0|0  19.4m   4.88m  105 Dec 28 16:38:29.865
```
TX:
```
  9455    *0     *0     *0       0 28380|0       0     0B 1.61G 91.0M      0 0|0 0|0  13.7m   7.89m  105 Dec 28 16:41:39.838
  9360    *0     *0     *0       0 28081|0       0     0B 1.61G 91.0M      0 0|0 0|0  13.6m   7.82m  105 Dec 28 16:41:40.835
  9430    *0     *0     *0       0 28310|0       0     0B 1.61G 91.0M      0 0|0 0|0  13.7m   7.86m  105 Dec 28 16:41:41.839
  9370    *0     *0     *0       0 28134|0       0     0B 1.61G 91.0M      0 0|0 0|0  13.6m   7.82m  105 Dec 28 16:41:42.848
```

## Test4
Update from test1:
* mongos configuration up to `c5.4xlarge`

### topology:
* mongod:  x3 r5d.xlarge: 32 GiB of memory, 4 vCPUs, 1 x 150 NVMe SSD, 64-bit platform
* mongos:  x1 c5.4xlarge: 32 GiB of Memory, 16 vCPUs, EBS only, 64-bit platform
* config:  x1 x1 t3.small: t3.small, 2 GiB of Memory, 2 vCPUs, EBS only, 64-bit platform 

#### Stop
reason: mongod `r5d.xlarge` CPU threshold exceed

INSERT:
```
insert query update delete getmore command flushes mapped vsize   res faults qrw arw net_in net_out conn                time
 31621    *0     *0     *0       0 31615|0       0     0B 1.63G  108M      0 0|0 0|0  21.4m   5.40m  105 Dec 28 17:22:23.682
 24024    *0     *0     *0       0 24074|0       0     0B 1.63G  108M      0 0|0 0|0  16.3m   4.10m  105 Dec 28 17:22:24.687
 26372    *0     *0     *0       0 26356|0       0     0B 1.63G  108M      0 0|0 0|0  17.8m   4.50m  105 Dec 28 17:22:25.689
 29757    *0     *0     *0       0 29742|0       0     0B 1.63G  108M      0 0|0 0|0  20.1m   5.08m  105 Dec 28 17:22:26.699
 30159    *0     *0     *0       0 30172|0       0     0B 1.63G  108M      0 0|0 0|0  20.4m   5.15m  105 Dec 28 17:22:27.723
```

TX:
```
insert query update delete getmore command flushes mapped vsize   res faults qrw arw net_in net_out conn                time
 10993    *0     *0     *0       0 32977|0       0     0B 1.62G  108M      0 0|0 0|0  15.9m   9.17m  105 Dec 28 17:23:53.684
 10768    *0     *0     *0       0 32349|0       0     0B 1.62G  108M      0 0|0 0|0  15.6m   8.99m  105 Dec 28 17:23:54.694
 10941    *0     *0     *0       0 32802|0       0     0B 1.62G  108M      0 0|0 0|0  15.8m   9.12m  105 Dec 28 17:23:55.682
```

## Test5
increase mongod instance + 1

### topology:
* mongod:  x4 r5d.xlarge: 32 GiB of memory, 4 vCPUs, 1 x 150 NVMe SSD, 64-bit platform
* mongos:  x1 c5.4xlarge: 32 GiB of Memory, 16 vCPUs, EBS only, 64-bit platform
* config:  x1 x1 t3.small: t3.small, 2 GiB of Memory, 2 vCPUs, EBS only, 64-bit platform 


### stop

INSERT:
```
insert query update delete getmore command flushes mapped vsize   res faults qrw arw net_in net_out conn                time
 35054    *0     *0     *0       0 35028|0       0     0B 1.61G 91.0M      0 0|0 0|0  23.7m   5.98m  105 Dec 29 10:02:07.672
 35580    *0     *0     *0       0 35605|0       0     0B 1.61G 91.0M      0 0|0 0|0  24.1m   6.07m  105 Dec 29 10:02:08.671
 26910    *0     *0     *0       0 26938|0       0     0B 1.61G 91.0M      0 0|0 0|0  18.2m   4.59m  105 Dec 29 10:02:09.663
 35530    *0     *0     *0       0 35476|0       0     0B 1.61G 91.0M      0 0|0 0|0  24.0m   6.06m  105 Dec 29 10:02:10.667
 28946    *0     *0     *0       0 28968|0       0     0B 1.61G 91.0M      0 0|0 0|0  19.6m   4.94m  105 Dec 29 10:02:11.677
```

## Test6
increase mongod instance  + 2

### topology:
* mongod:  x6 r5d.xlarge: 32 GiB of memory, 4 vCPUs, 1 x 150 NVMe SSD, 64-bit platform
* mongos:  x1 c5.4xlarge: 32 GiB of Memory, 16 vCPUs, EBS only, 64-bit platform
* config:  x1 x1 t3.small: t3.small, 2 GiB of Memory, 2 vCPUs, EBS only, 64-bit platform 


### stop

INSERT:
```
insert query update delete getmore command flushes mapped vsize   res faults qrw arw net_in net_out conn                time
 40261    *0     *0     *0       0 40266|0       0     0B 1.59G 74.0M      0 0|0 0|0  27.2m   6.86m  105 Dec 29 11:24:11.307
 39433    *0     *0     *0       0 39453|0       0     0B 1.60G 74.0M      0 0|0 0|0  26.7m   6.72m  105 Dec 29 11:24:12.324
 40790    *0     *0     *0       0 40762|0       0     0B 1.60G 74.0M      0 0|0 0|0  27.6m   6.95m  105 Dec 29 11:24:13.311
 40273    *0     *0     *0       0 40267|0       0     0B 1.60G 74.0M      0 0|0 0|0  27.2m   6.87m  105 Dec 29 11:24:14.305
```

TX:
```
 15801    *0     *0     *0       0 47399|0       0     0B 1.61G 93.0M      0 0|0 0|0  22.9m   13.2m  105 Dec 29 11:26:36.310
 15733    *0     *0     *0       0 47224|0       0     0B 1.61G 93.0M      0 0|0 0|0  22.8m   13.1m  105 Dec 29 11:26:37.303
 15721    *0     *0     *0       0 47125|0       0     0B 1.61G 93.0M      0 0|0 0|0  22.7m   13.1m  105 Dec 29 11:26:38.303
 15550    *0     *0     *0       0 46698|0       0     0B 1.61G 93.0M      0 0|0 0|0  22.5m   13.0m  105 Dec 29 11:26:39.307
```

## Test7
increase mongod instance = 2 instances
mongos up to c5.9xlarge
mongod r5d.xlarge =>r5d.2xlarge
### topology:
* mongod:  x2 r5d.2xlarge: 64 GiB of memory, 8 vCPUs, 1 x 150 NVMe SSD, 64-bit platform
* mongos:  x1 c5.4xlarge => c5.c5.9xlarge: 72 GiB of Memory, 36 vCPUs, EBS only, 64-bit platform
* config:  x1 x1 t3.small: t3.small, 2 GiB of Memory, 2 vCPUs, EBS only, 64-bit platform 

stop:

INSERT:
```
insert query update delete getmore command flushes mapped vsize   res faults qrw arw net_in net_out conn                time
 32227    *0     *0     *0       0 32236|0       0     0B 1.58G 55.0M      0 0|0 0|0  21.8m   5.50m  105 Dec 29 14:51:34.850
 33789    *0     *0     *0       0 33790|0       0     0B 1.58G 55.0M      0 0|0 0|0  22.9m   5.76m  105 Dec 29 14:51:35.853
 30590    *0     *0     *0       0 30608|0       0     0B 1.58G 55.0M      0 0|0 0|0  20.7m   5.22m  105 Dec 29 14:51:36.852
 34341    *0     *0     *0       0 34328|0       0     0B 1.58G 55.0M      0 0|0 0|0  23.2m   5.86m  105 Dec 29 14:51:37.851
 33054    *0     *0     *0       0 33124|0       0     0B 1.58G 55.0M      0 0|0 0|0  22.4m   5.64m  105 Dec 29 14:51:38.858
```
