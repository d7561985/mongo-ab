#!/bin/bash
# Скрипт для исправления проблем с MongoDB

echo "Fixing MongoDB configuration..."

# 1. Проверяем и создаем директорию
sudo mkdir -p /data/mongodb
sudo chown -R mongod:mongod /data/mongodb

# 2. Создаем правильную конфигурацию
sudo tee /etc/mongod.conf > /dev/null <<'EOF'
# Storage configuration
storage:
  dbPath: /data/mongodb
  journal:
    enabled: true
  engine: wiredTiger
  wiredTiger:
    engineConfig:
      cacheSizeGB: 16
      journalCompressor: snappy
      directoryForIndexes: true
    collectionConfig:
      blockCompressor: snappy
    indexConfig:
      prefixCompression: true

# Network configuration  
net:
  port: 50000
  bindIp: 0.0.0.0
  maxIncomingConnections: 10000

# Process management
processManagement:
  fork: true
  pidFilePath: /var/run/mongodb/mongod.pid

# Replication
replication:
  replSetName: rs0
  oplogSizeMB: 10240

# System log
systemLog:
  destination: file
  logAppend: true
  path: /var/log/mongodb/mongod.log
  verbosity: 0

# Security (временно отключена)
security:
  authorization: disabled
EOF

# 3. Проверяем конфигурацию
echo "Testing configuration..."
mongod --config /etc/mongod.conf --test

# 4. Перезапускаем MongoDB
echo "Restarting MongoDB..."
sudo systemctl restart mongod

# 5. Проверяем статус
sleep 5
sudo systemctl status mongod

# 6. Проверяем логи
echo "Last log entries:"
sudo tail -20 /var/log/mongodb/mongod.log