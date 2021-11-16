# MongoDB-AB

Some test based benchmarks which help check and understand and assess atomic operation and their RPC with different configurations

## WIP
Repo right now proof-of-concept with purpose to build AB test utility

## Requirement
Mongo >= 3.6 installed with replica set. For custom launch look into Makefile commands that help you launch and setup on fly instance

## How to
Just be sure that you specify launch url

[source]
store/mongo/mongo_test.go
---
var dbConnect = "mongodb://localhost:27021/?replicaSet=rs1"
---

## Tests
### `TestLoadMakeTransaction`
This most important. It calculates actual rpc and show it
Example: 

[source]
---
comb/sec: 7669.716911777314 duration: 60.017860541 15344
---
