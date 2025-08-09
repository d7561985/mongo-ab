#!/bin/bash
# Упрощенный скрипт установки MongoDB для быстрого развертывания

set -e

if [ "$#" -ne 1 ]; then
    echo "Usage: $0 <node_index_0_1_or_2>"
    exit 1
fi

NODE_INDEX=$1
PORT=50000

echo "Setting up MongoDB on node $NODE_INDEX..."

# 1. Быстрая настройка NVMe диска
echo "Setting up NVMe disk..."
NVME_DEVICE=$(lsblk | grep nvme | grep -v nvme0n1 | head -1 | awk '{print "/dev/"$1}')

if [ -n "$NVME_DEVICE" ]; then
    echo "Found NVMe device: $NVME_DEVICE"
    sudo mkfs.xfs -f $NVME_DEVICE 2>/dev/null || true
    sudo mkdir -p /data/mongodb
    sudo mount -o noatime,nodiratime $NVME_DEVICE /data/mongodb || true
    sudo chown ec2-user:ec2-user /data/mongodb
else
    echo "Warning: No NVMe device found, using EBS"
    sudo mkdir -p /data/mongodb
    sudo chown ec2-user:ec2-user /data/mongodb
fi

# 2. Установка MongoDB (без update системы для скорости)
echo "Installing MongoDB..."
cat <<EOF | sudo tee /etc/yum.repos.d/mongodb-org-7.0.repo
[mongodb-org-7.0]
name=MongoDB Repository
baseurl=https://repo.mongodb.org/yum/amazon/2/mongodb-org/7.0/x86_64/
gpgcheck=1
enabled=1
gpgkey=https://pgp.mongodb.com/server-7.0.asc
EOF

sudo yum install -y mongodb-org

# 3. Минимальная конфигурация
cat <<EOF | sudo tee /etc/mongod.conf
storage:
  dbPath: /data/mongodb
  journal:
    enabled: true

net:
  port: $PORT
  bindIp: 0.0.0.0

replication:
  replSetName: rs0

processManagement:
  fork: true
  pidFilePath: /var/run/mongodb/mongod.pid

systemLog:
  destination: file
  path: /var/log/mongodb/mongod.log
EOF

# 4. Запуск MongoDB
sudo chown -R mongod:mongod /data/mongodb
sudo systemctl enable mongod
sudo systemctl start mongod

echo "MongoDB installed and running on port $PORT"
echo "Run rs.status() to check replica set status"