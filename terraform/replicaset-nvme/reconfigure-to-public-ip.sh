#!/bin/bash
# Скрипт для переконфигурации существующего replica set на публичные IP

set -e

if [ "$#" -ne 3 ]; then
    echo "Usage: $0 <public_ip_1> <public_ip_2> <public_ip_3>"
    echo "Example: $0 52.29.31.23 63.178.115.97 3.78.197.221"
    exit 1
fi

PUBLIC_IP1=$1
PUBLIC_IP2=$2
PUBLIC_IP3=$3
PORT=50000

echo "Reconfiguring replica set to use public IPs..."
echo "  Node 1: $PUBLIC_IP1:$PORT"
echo "  Node 2: $PUBLIC_IP2:$PORT"
echo "  Node 3: $PUBLIC_IP3:$PORT"

# Создаем JS файл для переконфигурации
cat > /tmp/reconfig-public.js <<EOF
print("Getting current replica set configuration...");
var cfg = rs.conf();

print("Updating member hosts to public IPs...");
cfg.members[0].host = "$PUBLIC_IP1:$PORT";
cfg.members[1].host = "$PUBLIC_IP2:$PORT";
cfg.members[2].host = "$PUBLIC_IP3:$PORT";

print("Applying new configuration...");
var result = rs.reconfig(cfg);
printjson(result);

if (result.ok == 1) {
    print("✅ Replica set reconfigured successfully!");
    sleep(5000);
    
    print("New configuration:");
    printjson(rs.conf());
    
    print("");
    print("Connection string for internet access:");
    print("mongodb://$PUBLIC_IP1:$PORT,$PUBLIC_IP2:$PORT,$PUBLIC_IP3:$PORT/?replicaSet=rs0");
} else {
    print("❌ Failed to reconfigure replica set");
    quit(1);
}
EOF

# Копируем скрипт на удаленную машину
echo "Copying script to remote server..."
scp -o StrictHostKeyChecking=no -i ~/.ssh/id_rsa /tmp/reconfig-public.js ec2-user@$PUBLIC_IP1:/tmp/

# Подключаемся к первой ноде и выполняем переконфигурацию
echo "Connecting to $PUBLIC_IP1 to reconfigure..."
ssh -o StrictHostKeyChecking=no -i ~/.ssh/id_rsa ec2-user@$PUBLIC_IP1 \
    "mongosh --port $PORT --quiet /tmp/reconfig-public.js && rm -f /tmp/reconfig-public.js"

rm -f /tmp/reconfig-public.js

echo ""
echo "✅ Done! You can now connect from internet using:"
echo "mongodb://$PUBLIC_IP1:$PORT,$PUBLIC_IP2:$PORT,$PUBLIC_IP3:$PORT/?replicaSet=rs0"