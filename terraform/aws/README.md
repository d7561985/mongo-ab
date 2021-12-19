# XX

```bash
$ aws ec2 describe-images --filters "Name=root-device-type,Values=ebs" "Name=name,Values=aws*linux*" --owners self amazon
$ aws ec2 describe-images --filters "Name=root-device-type,Values=ebs" "Name=name,Values=amzn2-ami-kernel-5.10-hvm-2.0.20211201.0-x86_64-gp2"
$ aws ec2 describe-images --filters "Name=root-device-type,Values=ebs" "Name=name,Values=bitnami-mongodb-5.0.4-6-r01-linux-debian-10-x86_64-hvm-ebs-nami" 

```

Output:
```json
[
  {
    "Architecture": "x86_64",
    "CreationDate": "2011-10-07T09:09:03.000Z",
    "ImageId": "ami-ffecde8b",
    "ImageLocation": "063491364108/ubuntu-8.04-hardy-server-amd64-20111006",
    "ImageType": "machine",
    "Public": true,
    "KernelId": "aki-4cf5c738",
    "OwnerId": "063491364108",
    "RamdiskId": "ari-2ef5c75a",
    "State": "available",
    "BlockDeviceMappings": [
      {
        "DeviceName": "/dev/sda1",
        "Ebs": {
          "Encrypted": false,
          "DeleteOnTermination": true,
          "SnapshotId": "snap-eb7aa883",
          "VolumeSize": 8,
          "VolumeType": "standard"
        }
      }, {
        "DeviceName": "/dev/sdb",
        "VirtualName": "ephemeral0"
      }
    ],
    "Description": "Ubuntu 8.04 Hardy server amd64 20111006",
    "Hypervisor": "xen",
    "Name": "ubuntu-8.04-hardy-server-amd64-20111006",
    "RootDeviceName": "/dev/sda1",
    "RootDeviceType": "ebs",
    "VirtualizationType": "paravirtual"
  }
]
```


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


sh.addShard( "rs1/localhost:27021")
sh.addShard( "rs2/localhost:27022")
sh.addShard( "rs3/localhost:27023")
sh.addShard( "rs4/localhost:27024")
sh.addShard( "rs5/localhost:27025")
sh.addShard( "rs6/localhost:27026")

...sudo mkdir data/config1 && sudo mongod --configsvr --port 27018 --replSet rs0 --dbpath data/config1 --bind_ip localhost
sudo mkdir data/config2 && sudo mongod --configsvr --port 27019 --replSet rs0 --dbpath data/config2 --bind_ip localhost
sudo mkdir data/config3 && sudo mongod --configsvr --port 27020 --replSet rs0 --dbpath data/config3 --bind_ip localhost

rs.initiate( { _id : "rs0",  configsvr: true,  members: [{ _id: 0, host: "localhost:27018" },{ _id: 1, host: "localhost:27019" },{ _id: 2, host: "localhost:27020" }]})

mongos --port 40000 --configdb rs0/localhost:27020,localhost:27019,localhost:27018

mongo --port 40000

db.printShardingStatus()


> sh.enableSharding("transactions")
> db.logs.ensureIndex({login: 1})  //устанавливаем индекс, который будет нам служить в качестве shard Key
> db.runCommand({shardCollection: "transactions.logs", key: {login: 1}})    //указываем шардируемую коллекцию, а вторым параметром - shard key для нее.

sh.addShard( "rs1/localhost:27021,localhost:27022,localhost:27023")
```
