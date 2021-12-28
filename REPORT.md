# Reports

## AWS

## Test1

### topology:
mongod:  x3 r5d.xlarge: 32 GiB of memory, 4 vCPUs, 1 x 150 NVMe SSD, 64-bit platform
mongos:  x1 t3.small: t3.small, 2 GiB of Memory, 2 vCPUs, EBS only, 64-bit platform
config:  x1 x1 t3.small: t3.small, 2 GiB of Memory, 2 vCPUs, EBS only, 64-bit platform 

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
