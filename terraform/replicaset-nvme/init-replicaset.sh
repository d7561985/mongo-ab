#!/bin/bash
# Скрипт для финальной инициализации replica set после развертывания

set -e

echo "MongoDB Replica Set Initialization Script"
echo "========================================"

# Проверяем аргументы
if [ "$#" -ne 3 ]; then
    echo "Usage: $0 <node1_ip> <node2_ip> <node3_ip>"
    echo "Example: $0 18.185.123.45 18.185.123.46 18.185.123.47"
    exit 1
fi

NODE1_IP=$1
NODE2_IP=$2
NODE3_IP=$3
PORT=50000

echo "Initializing replica set with nodes:"
echo "  Node 1 (Primary): $NODE1_IP:$PORT"
echo "  Node 2: $NODE2_IP:$PORT"
echo "  Node 3: $NODE3_IP:$PORT"
echo ""

# Создаем временный JS файл для инициализации
cat <<EOF > /tmp/init-rs.js
// Инициализация replica set
var config = {
  _id: "rs0",
  members: [
    { 
      _id: 0, 
      host: "${NODE1_IP}:${PORT}", 
      priority: 2,
      tags: { "dc": "eu-central-1", "use": "primary" }
    },
    { 
      _id: 1, 
      host: "${NODE2_IP}:${PORT}", 
      priority: 1,
      tags: { "dc": "eu-central-1", "use": "secondary" }
    },
    { 
      _id: 2, 
      host: "${NODE3_IP}:${PORT}", 
      priority: 1,
      tags: { "dc": "eu-central-1", "use": "secondary" }
    }
  ],
  settings: {
    // Оптимизация для производительности записи
    writeConcernMajorityJournalDefault: false,
    
    // Быстрое переключение при сбоях
    electionTimeoutMillis: 10000,
    heartbeatIntervalMillis: 2000,
    catchUpTimeoutMillis: 30000,
    
    // Оптимизация для записи
    chainingAllowed: true,
    
    // Конфигурация для read preference
    getLastErrorDefaults: { w: 1, wtimeout: 5000 }
  }
};

// Пытаемся инициализировать
var result = rs.initiate(config);
printjson(result);

// Ждем стабилизации
sleep(5000);

// Проверяем статус
var status = rs.status();
printjson(status);

// Дополнительная конфигурация после инициализации
if (result.ok == 1) {
  print("Replica set initialized successfully!");
  
  // Настраиваем read concern по умолчанию
  db.adminCommand({ 
    setDefaultRWConcern: 1, 
    defaultReadConcern: { level: "local" },
    defaultWriteConcern: { w: 1, j: false }
  });
  
  print("Default read/write concerns configured for maximum write performance");
}
EOF

# Выполняем инициализацию
echo "Connecting to primary node and initializing replica set..."
mongosh --host ${NODE1_IP}:${PORT} --quiet < /tmp/init-rs.js

# Очищаем временный файл
rm -f /tmp/init-rs.js

echo ""
echo "Replica set initialization completed!"
echo ""
echo "Connection string for your application:"
echo "mongodb://${NODE1_IP}:${PORT},${NODE2_IP}:${PORT},${NODE3_IP}:${PORT}/?replicaSet=rs0"
echo ""
echo "To check replica set status, connect to any node and run:"
echo "mongosh --host ${NODE1_IP}:${PORT} --eval 'rs.status()'"