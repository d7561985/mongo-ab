%{ for id, host in val }
sh.addShard( "${id}/${host}:27017")
%{ endfor }
