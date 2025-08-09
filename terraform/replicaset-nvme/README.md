# MongoDB Replica Set с NVMe дисками на AWS

## Описание

Эта конфигурация развертывает высокопроизводительный MongoDB Replica Set с:
- 3 узла в регионе eu-central-1
- EC2 инстансы i3en.2xlarge с 2x 1.25TB NVMe SSD
- MongoDB 7.0 оптимизированная для максимальной производительности записи
- Доступ через порт 50000 из интернета

## Быстрый старт

### 1. Подготовка

```bash
# Клонируйте репозиторий
cd terraform/aws/replicaset-nvme

# Убедитесь что у вас есть SSH ключ
ls ~/.ssh/id_rsa.pub

# Настройте AWS credentials
aws configure
```

### 2. Настройка переменных (опционально)

Создайте файл `terraform.tfvars`:

```hcl
# Ограничить доступ только вашими IP адресами
allowed_ips = ["YOUR_IP/32", "OFFICE_IP/32"]

# Использовать другой тип инстанса
instance_type = "i3en.xlarge"  # Меньше и дешевле

# Другой SSH ключ
ssh_public_key_path = "~/.ssh/custom_key.pub"
```

### 3. Развертывание

```bash
# Инициализация
terraform init

# Планирование
terraform plan

# Развертывание
terraform apply

# Сохраните outputs
terraform output -json > outputs.json
```

### 4. Инициализация Replica Set

После развертывания получите IP адреса:

```bash
terraform output node_ips
```

SSH на первый узел:

```bash
ssh -i ~/.ssh/id_rsa ec2-user@<NODE1_PUBLIC_IP>
```

Запустите скрипт инициализации:

```bash
# Скопируйте скрипт на сервер
scp init-replicaset.sh ec2-user@<NODE1_PUBLIC_IP>:~/

# Запустите с IP адресами всех узлов
ssh ec2-user@<NODE1_PUBLIC_IP>
chmod +x init-replicaset.sh
./init-replicaset.sh <NODE1_IP> <NODE2_IP> <NODE3_IP>
```

### 5. Проверка

```bash
# Проверьте статус replica set
mongosh --host <NODE1_IP>:50000 --eval 'rs.status()'

# Проверьте конфигурацию
mongosh --host <NODE1_IP>:50000 --eval 'rs.conf()'
```

## Запуск тестов производительности

```bash
# Из корня проекта mongo-ab
./mongo-ab mongo \
  --addr "$(terraform output -raw connection_string)" \
  --threads 200 \
  --maxUser 1000000 \
  --operation tx \
  --compression snappy \
  --wcJournal=false \
  --W 1
```

## Оптимизации производительности

### 1. Конфигурация MongoDB

- **WiredTiger Cache**: 16GB (50% RAM)
- **Oplog Size**: 10GB для поддержки высокой нагрузки
- **Journal**: Включен, но `writeConcernMajorityJournalDefault: false`
- **Compression**: Snappy для баланса скорости и сжатия

### 2. Системные оптимизации

- **NVMe диски**: Смонтированы с `noatime,nodiratime,nobarrier`
- **THP отключен**: Transparent Huge Pages отключены
- **Лимиты**: Увеличены лимиты файлов и процессов
- **Сеть**: Оптимизированы TCP параметры

### 3. Рекомендации по Write Concern

Для максимальной производительности:
```javascript
// В приложении
{ w: 1, j: false }  // Подтверждение от primary без journal
```

Для надежности:
```javascript
{ w: "majority", j: true }  // Подтверждение от большинства с journal
```

## Мониторинг

### CloudWatch метрики (если включено)

- CPU использование
- Disk I/O
- Network throughput

### MongoDB метрики

```bash
# Статистика в реальном времени
mongosh --host <NODE_IP>:50000 --eval 'db.serverStatus()'

# Статистика репликации
mongosh --host <NODE_IP>:50000 --eval 'rs.printSlaveReplicationInfo()'
```

## Безопасность

⚠️ **ВАЖНО**: По умолчанию MongoDB доступна из любого места!

### Рекомендации:

1. **Ограничьте IP адреса** в `allowed_ips`
2. **Включите аутентификацию**:
   ```javascript
   use admin
   db.createUser({
     user: "admin",
     pwd: "STRONG_PASSWORD",
     roles: ["root"]
   })
   ```
3. **Используйте SSL/TLS** для шифрования трафика
4. **Настройте VPN** для безопасного доступа

## Стоимость

### On-Demand инстансы (по умолчанию)
Примерная стоимость (eu-central-1):
- i3en.2xlarge: ~$0.752/час x 3 = ~$2.26/час
- EBS root volumes: ~$3/месяц
- Сетевой трафик: зависит от использования

**Итого**: ~$1,600/месяц для 3 узлов

### Spot инстансы (экономия до 70%)
Для использования spot инстансов:

```bash
# Создайте terraform.tfvars
cp terraform.tfvars.example terraform.tfvars

# Отредактируйте параметры
use_spot_instances = true
spot_max_price = "0.40"  # ~50% от on-demand цены
```

Примерная стоимость со spot:
- i3en.2xlarge spot: ~$0.25-0.30/час x 3 = ~$0.75-0.90/час
- **Итого**: ~$540-650/месяц (экономия ~$1000/месяц!)

⚠️ **Важно для spot инстансов**:
- Данные на NVMe дисках сохраняются при остановке (`instance_interruption_behavior = "stop"`)
- Используйте persistent spot requests для автоматического перезапуска
- Мониторьте spot цены: `aws ec2 describe-spot-price-history --instance-types i3en.2xlarge --region eu-central-1`

## Очистка ресурсов

```bash
# Удалить все ресурсы
terraform destroy
```

## Troubleshooting

### MongoDB не запускается
```bash
# Проверьте логи
sudo journalctl -u mongod -f

# Проверьте диски
df -h
lsblk
```

### Проблемы с репликацией
```bash
# Проверьте сеть между узлами
ping <OTHER_NODE_PRIVATE_IP>

# Проверьте порты
sudo netstat -tlnp | grep 50000
```

### Низкая производительность
1. Проверьте использование CPU/RAM/Disk
2. Проверьте `rs.status()` для lag
3. Используйте `mongostat` для мониторинга


