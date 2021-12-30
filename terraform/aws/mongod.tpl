# mongod.conf

# for documentation of all options, see:
#   http://docs.mongodb.org/manual/reference/configuration-options/

# where to write logging data.
systemLog:
   quiet: true
   destination: file
   logAppend: false
   logRotate: rename
   path: /var/log/mongodb/mongod.log

# Where and how to store data.
storage:
  dbPath: /data/mongodb
  directoryPerDB: true
  journal:
    enabled: true
    commitIntervalMs: 100
  wiredTiger:
    engineConfig:
      cacheSizeGB: 25
      directoryForIndexes: true
      journalCompressor: snappy # none

# how the process runs
processManagement:
  fork: true  # fork and run in background
  pidFilePath: /var/run/mongodb/mongod.pid  # location of pidfile
  timeZoneInfo: /usr/share/zoneinfo

# network interfaces
net:
  port: 50000
  bindIpAll: true
  #bindIp: 127.0.0.1  # Enter 0.0.0.0,:: to bind to all IPv4 and IPv6 addresses or, alternatively, use the net.bindIpAll setting.

#security:

#operationProfiling:

replication:
  replSetName: ${id}

sharding:
   clusterRole: ${clusterRole} # shardsvr - shard, configsvr - config serv
   archiveMovedChunks: false

## Enterprise-Only Options

#auditLog:

#snmp:
