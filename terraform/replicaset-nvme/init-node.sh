#!/bin/bash
set -e

# Переменные из Terraform
NODE_INDEX=${node_index}
TOTAL_NODES=${total_nodes}
PORT=${port}

# Ждем инициализации системы
sleep 30

# Обновляем систему
sudo yum update -y

# Устанавливаем необходимые пакеты
sudo yum install -y nvme-cli

# Настраиваем NVMe диски
echo "Setting up NVMe disks..."
NVME_DEVICES=$(sudo nvme list | grep "Amazon EC2 NVMe Instance Storage" | awk '{print $1}')

# Используем первый NVMe диск для MongoDB
NVME_DEVICE=$(echo $NVME_DEVICES | awk '{print $1}')

if [ -n "$NVME_DEVICE" ]; then
    echo "Found NVMe device: $NVME_DEVICE"
    
    # Форматируем диск с XFS для лучшей производительности
    sudo mkfs.xfs -f $NVME_DEVICE
    
    # Создаем директорию для данных
    sudo mkdir -p /data/mongodb
    
    # Монтируем с оптимальными параметрами для MongoDB
    sudo mount -o noatime,nodiratime $NVME_DEVICE /data/mongodb
    
    # Добавляем в fstab для автомонтирования
    echo "$NVME_DEVICE /data/mongodb xfs noatime,nodiratime 0 0" | sudo tee -a /etc/fstab
else
    echo "ERROR: No NVMe device found!"
    exit 1
fi

# Устанавливаем MongoDB 7.0
cat <<EOF | sudo tee /etc/yum.repos.d/mongodb-org-7.0.repo
[mongodb-org-7.0]
name=MongoDB Repository
baseurl=https://repo.mongodb.org/yum/amazon/2/mongodb-org/7.0/x86_64/
gpgcheck=1
enabled=1
gpgkey=https://pgp.mongodb.com/server-7.0.asc
EOF

sudo yum install -y mongodb-org

# Создаем оптимизированную конфигурацию MongoDB
cat <<EOF | sudo tee /etc/mongod.conf
# Storage configuration
storage:
  dbPath: /data/mongodb
  wiredTiger:
    engineConfig:
      cacheSizeGB: 16  # 50% от RAM для i3en.2xlarge (32GB)
      journalCompressor: snappy
      directoryForIndexes: true
    collectionConfig:
      blockCompressor: snappy
    indexConfig:
      prefixCompression: true

# Network configuration
net:
  port: $PORT
  bindIp: 0.0.0.0
  maxIncomingConnections: 10000

# Process management
processManagement:
  fork: true
  pidFilePath: /var/run/mongodb/mongod.pid

# Replication
replication:
  replSetName: rs0
  oplogSizeMB: 10240  # 10GB oplog для высокой нагрузки записи

# Operation profiling
operationProfiling:
  mode: off

# Security (временно отключена - включите после настройки!)
security:
  authorization: disabled

# System log
systemLog:
  destination: file
  logAppend: true
  path: /var/log/mongodb/mongod.log
  verbosity: 0

# Оптимизация для производительности записи
setParameter:
  # Увеличиваем размер пакетов для репликации
  replWriterThreadCount: 32
  replBatchLimitBytes: 104857600  # 100MB
  # Оптимизация WiredTiger
  wiredTigerConcurrentReadTransactions: 256
  wiredTigerConcurrentWriteTransactions: 256
  # Увеличиваем лимиты
  internalQueryMaxBlockingSortMemoryUsageBytes: 536870912  # 512MB
EOF

# Настраиваем права доступа
sudo chown -R mongod:mongod /data/mongodb
sudo chmod 755 /data/mongodb

# Оптимизация системы для MongoDB
# Отключаем THP (Transparent Huge Pages)
echo never | sudo tee /sys/kernel/mm/transparent_hugepage/enabled
echo never | sudo tee /sys/kernel/mm/transparent_hugepage/defrag

# Настраиваем лимиты
cat <<EOF | sudo tee /etc/security/limits.d/99-mongodb.conf
mongod soft nofile 64000
mongod hard nofile 64000
mongod soft nproc 64000
mongod hard nproc 64000
EOF

# Настраиваем sysctl для производительности
cat <<EOF | sudo tee /etc/sysctl.d/99-mongodb.conf
# Сетевые оптимизации
net.core.somaxconn = 4096
net.ipv4.tcp_max_syn_backlog = 4096
net.ipv4.tcp_fin_timeout = 30
net.ipv4.tcp_keepalive_intvl = 30
net.ipv4.tcp_keepalive_time = 120
net.ipv4.tcp_max_tw_buckets = 4096

# Оптимизация памяти
vm.swappiness = 1
vm.zone_reclaim_mode = 0
EOF

sudo sysctl -p /etc/sysctl.d/99-mongodb.conf

# Запускаем MongoDB
sudo systemctl enable mongod
sudo systemctl start mongod

# Ждем запуска MongoDB
sleep 10

# Проверяем статус
sudo systemctl status mongod

# Создаем скрипт инициализации replica set (только на первом узле)
if [ "$NODE_INDEX" -eq 0 ]; then
    cat <<'INIT_SCRIPT' > /home/ec2-user/init-replicaset.js
// Ждем доступности всех узлов
sleep(30000);

// Получаем IP адреса всех узлов через AWS metadata
const myIp = db.adminCommand({ getParameter: 1, "net.bindIp": 1 }).net.bindIp;

rs.initiate({
  _id: "rs0",
  members: [
    { _id: 0, host: "NODE_0_IP:50000", priority: 2 },
    { _id: 1, host: "NODE_1_IP:50000", priority: 1 },
    { _id: 2, host: "NODE_2_IP:50000", priority: 1 }
  ],
  settings: {
    // Оптимизация для производительности записи
    writeConcernMajorityJournalDefault: false,
    catchUpTimeoutMillis: 30000,
    electionTimeoutMillis: 10000,
    heartbeatIntervalMillis: 2000
  }
});

// Ждем стабилизации
sleep(10000);

// Проверяем статус
rs.status();
INIT_SCRIPT

    echo "Replica set init script created at /home/ec2-user/init-replicaset.js"
    echo "Run it manually after all nodes are ready!"
fi

echo "MongoDB node $NODE_INDEX setup completed!"
echo "MongoDB is running on port $PORT"