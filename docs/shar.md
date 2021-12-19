# mongo shard

## prepare instances
```bash
sudo mkdir data/shard1 && sudo mongod --shardsvr --port 27021 --replSet rs1 --dbpath data/shard1 --bind_ip localhost
sudo mkdir data/shard2 && sudo mongod --shardsvr --port 27022 --replSet rs2 --dbpath data/shard2 --bind_ip localhost
sudo mkdir data/shard3 && sudo mongod --shardsvr --port 27023 --replSet rs3 --dbpath data/shard3 --bind_ip localhost
sudo mkdir data/shard4 && sudo mongod --shardsvr --port 27024 --replSet rs4 --dbpath data/shard4 --bind_ip localhost
sudo mkdir data/shard5 && sudo mongod --shardsvr --port 27025 --replSet rs5 --dbpath data/shard5 --bind_ip localhost
sudo mkdir data/shard6 && sudo mongod --shardsvr --port 27026 --replSet rs6 --dbpath data/shard6 --bind_ip localhost

rs.initiate( { _id : "rs1",  members: [{ _id: 0, host: "localhost:27021" }]})
rs.initiate( { _id : "rs2",  members: [{ _id: 0, host: "localhost:27022" }]})
rs.initiate( { _id : "rs3",  members: [{ _id: 0, host: "localhost:27023" }]})
rs.initiate( { _id : "rs4",  members: [{ _id: 0, host: "localhost:27024" }]})
rs.initiate( { _id : "rs5",  members: [{ _id: 0, host: "localhost:27025" }]})
rs.initiate( { _id : "rs6",  members: [{ _id: 0, host: "localhost:27026" }]})


```
# prepare config server
```bash
sudo mkdir data/config1 && sudo mongod --configsvr --port 27018 --replSet rs0 --dbpath data/config1 --bind_ip localhost
sudo mkdir data/config2 && sudo mongod --configsvr --port 27019 --replSet rs0 --dbpath data/config2 --bind_ip localhost
sudo mkdir data/config3 && sudo mongod --configsvr --port 27020 --replSet rs0 --dbpath data/config3 --bind_ip localhost

rs.initiate( { _id : "rs0",  configsvr: true,  members: [{ _id: 0, host: "localhost:27018" },{ _id: 1, host: "localhost:27019" },{ _id: 2, host: "localhost:27020" }]})
```

# start mongos
```bash
mongos --port 40000 --configdb rs0/localhost:27020,localhost:27019,localhost:27018
mongo --port 40000

sh.addShard( "rs1/3.122.244.130:27017")
sh.addShard( "rs2/3.121.112.233:27017")
sh.addShard( "rs3/18.159.196.118:27017")
sh.addShard( "rs4/localhost:27024")
sh.addShard( "rs5/localhost:27025")
sh.addShard( "rs6/localhost:27026")

db.printShardingStatus()

> sh.enableSharding("transactions")
> db.logs.ensureIndex({login: 1})  //устанавливаем индекс, который будет нам служить в качестве shard Key
> db.runCommand({shardCollection: "transactions.logs", key: {login: 1}})    //указываем шардируемую коллекцию, а вторым параметром - shard key для нее.

sh.addShard( "rs1/localhost:27021,localhost:27022,localhost:27023")
```
