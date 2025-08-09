%{ for id, host in val ~}
sh.addShard( "${id}/${host}:50000")
%{ endfor ~}
sh.enableSharding("${shardDB}")
