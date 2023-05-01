# MongoDB-AB

Some test based benchmarks which help check and understand and assess atomic operation and their RPC with different configurations

## WIP
Repo right now proof-of-concept with purpose to build AB test utility

## Requirement
Mongo >= 5.0 installed with shards. For custom launch look into Makefile commands that help you launch and setup on fly instance

## How to
Just be sure that you specify launch url

store/mongo/mongo_test.go
```
var dbConnect = "mongodb://localhost:27021/?replicaSet=rs1"
```

```bash
Incorrect Usage: flag provided but not defined: -р

NAME:
   main mongo - 

USAGE:
   main mongo [command options] [arguments...]

DESCRIPTION:
   run mongodb compliance test which runs transactions

OPTIONS:
   --threads value, -t value    (default: 100) [$THREADS]
   --maxUser value, -m value    (default: 100000) [$MAX_USER]
   --operation value, -o value  What test start: tx - transaction intense, insert - only insert (default: "tx") [$OPERATION]
   --addr value                 (default: "mongodb://3.67.76.232:50000") [$MONGO_ADDR]
   --db value                   (default: "db") [$MONGO_DB]
   --balance value              (default: "bench_balance") [$MONGO_COLLECTION_BALANCE]
   --journal value              (default: "bench_journal") [$MONGO_COLLECTION_JOURNAL]
   --compression value          zlib, zstd, snappy (default: "snappy") [$MONGO_COMPRESSION]
   --compressionLevel value     zlib: max 9, zstd: max 20, snappy: not used (default: 0) [$MONGO_COMPRESSION_LEVEL]
   --wcJournal                  (default: false) [$MONGO_WRITE_CONCERN_J]
   --shards value               (default: 0) [$MONGO_SHARDS]
   --help, -h                   show help (default: false)

```

## Tests
### `TestLoadMakeTransaction`
This most important. It calculates actual rpc and show it
Example: 

```
comb/sec: 7669.716911777314 duration: 60.017860541 15344
```
