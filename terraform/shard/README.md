# Mongo shard cluster on AWS

Required default VPN + subnet

Note: zone is hardcoded: `eu-central-1`

Security rules:
* mongos port 50000 allow fron outside
* mongod 27017 only inside default VPC cidr
