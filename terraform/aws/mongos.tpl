# mongod.conf
#
# for documentation of all options, see:
#   https://docs.mongodb.com/manual/reference/configuration-file-settings-command-line-options-mapping/#std-label-conf-file-command-line-mapping

# where to write logging data.
systemLog:
   quiet: false
   destination: file
   logAppend: false
   path: /var/log/mongodb/mongod.log

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

sharding:
   configDB: ${configDB}
