terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = "eu-central-1"
}

# Получаем текущий IP для безопасности
data "http" "myip" {
  url = "http://ipv4.icanhazip.com"
}

# VPC и сеть
resource "aws_vpc" "mongo_vpc" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name = "mongo-replicaset-vpc"
  }
}

resource "aws_internet_gateway" "mongo_igw" {
  vpc_id = aws_vpc.mongo_vpc.id

  tags = {
    Name = "mongo-replicaset-igw"
  }
}

resource "aws_subnet" "mongo_subnet" {
  vpc_id                  = aws_vpc.mongo_vpc.id
  cidr_block              = "10.0.1.0/24"
  availability_zone       = "eu-central-1a"
  map_public_ip_on_launch = true

  tags = {
    Name = "mongo-replicaset-subnet"
  }
}

resource "aws_route_table" "mongo_rt" {
  vpc_id = aws_vpc.mongo_vpc.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.mongo_igw.id
  }

  tags = {
    Name = "mongo-replicaset-rt"
  }
}

resource "aws_route_table_association" "mongo_rta" {
  subnet_id      = aws_subnet.mongo_subnet.id
  route_table_id = aws_route_table.mongo_rt.id
}

# Security Group с ограниченным доступом
resource "aws_security_group" "mongo_sg" {
  name_prefix = "mongo-replicaset-"
  description = "Security group for MongoDB Replica Set"
  vpc_id      = aws_vpc.mongo_vpc.id

  # SSH доступ только с вашего IP
  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["${chomp(data.http.myip.response_body)}/32"]
    description = "SSH from my IP"
  }

  # MongoDB порт 50000 - доступ из интернета (можно ограничить)
  ingress {
    from_port   = 50000
    to_port     = 50000
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]  # Замените на конкретные IP для безопасности
    description = "MongoDB custom port"
  }

  # Внутренняя коммуникация между узлами
  ingress {
    from_port = 0
    to_port   = 65535
    protocol  = "tcp"
    self      = true
    description = "Internal replica set communication"
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "mongo-replicaset-sg"
  }
}

# Key pair для SSH доступа
resource "aws_key_pair" "mongo_key" {
  key_name   = "mongo-replicaset-key"
  public_key = file("~/.ssh/id_rsa.pub")  # Укажите путь к вашему публичному ключу
}

# EC2 инстансы с NVMe дисками
resource "aws_instance" "mongo_node" {
  count = 3

  # i3en.2xlarge имеет 2x 1.25TB NVMe SSD
  # Альтернативы: i3en.xlarge (1x 1.25TB), i3en.4xlarge (2x 2.5TB)
  instance_type = var.instance_type
  
  ami                    = data.aws_ami.amazon_linux_2.id
  subnet_id              = aws_subnet.mongo_subnet.id
  vpc_security_group_ids = [aws_security_group.mongo_sg.id]
  key_name              = aws_key_pair.mongo_key.key_name

  # Spot instance configuration
  dynamic "instance_market_options" {
    for_each = var.use_spot_instances ? [1] : []
    content {
      market_type = "spot"
      spot_options {
        max_price                      = var.spot_max_price
        spot_instance_type             = "persistent"
        instance_interruption_behavior = "stop"  # Важно для данных!
      }
    }
  }

  # Отключаем Nitro Enclaves для лучшей производительности
  enclave_options {
    enabled = false
  }

  # Оптимизация для производительности
  ebs_optimized = true
  
  # Credit specification для постоянной производительности
  credit_specification {
    cpu_credits = "unlimited"
  }

  root_block_device {
    volume_type = "gp3"
    volume_size = 30
    iops        = 3000
    throughput  = 125
    encrypted   = true
  }

  user_data = base64encode(templatefile("${path.module}/init-node.sh", {
    node_index  = count.index
    total_nodes = 3
    port        = 50000
  }))

  tags = {
    Name = "mongo-replicaset-node-${count.index + 1}"
    Type = "replica-set-member"
    Role = count.index == 0 ? "primary-candidate" : "secondary"
  }
}

# Elastic IPs для стабильного подключения
resource "aws_eip" "mongo_eip" {
  count    = 3
  instance = aws_instance.mongo_node[count.index].id
  domain   = "vpc"

  tags = {
    Name = "mongo-replicaset-eip-${count.index + 1}"
  }
}

# Получаем последний Amazon Linux 2 AMI
data "aws_ami" "amazon_linux_2" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn2-ami-hvm-*-x86_64-gp2"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }
}

# Health check для MongoDB
resource "null_resource" "mongodb_health_check" {
  depends_on = [aws_instance.mongo_node]
  
  # Триггер для повторной проверки при изменении инстансов
  triggers = {
    instance_ids = join(",", aws_instance.mongo_node[*].id)
  }

  # Ждем 3 минуты для завершения user_data
  provisioner "local-exec" {
    command = "sleep 180"
  }

  # Проверяем статус MongoDB на каждой ноде
  provisioner "local-exec" {
    command = <<-EOT
      echo "Checking MongoDB status on all nodes..."
      FAILED=0
      
      %{ for idx, eip in aws_eip.mongo_eip ~}
      echo "Checking node ${idx + 1} (${eip.public_ip})..."
      ssh -o StrictHostKeyChecking=no -o ConnectTimeout=10 -i ${replace(var.ssh_public_key_path, ".pub", "")} ec2-user@${eip.public_ip} \
        "sudo systemctl is-active mongod && echo '✓ MongoDB running on node ${idx + 1}' || (echo '✗ MongoDB FAILED on node ${idx + 1}' && exit 1)" \
        || FAILED=$((FAILED + 1))
      %{ endfor ~}
      
      if [ $FAILED -gt 0 ]; then
        echo "❌ MongoDB deployment FAILED on $FAILED nodes!"
        echo "Check logs with: ssh -i ${replace(var.ssh_public_key_path, ".pub", "")} ec2-user@<NODE_IP> 'sudo journalctl -u mongod -n 50'"
        exit 1
      else
        echo "✅ All MongoDB nodes are running successfully!"
      fi
    EOT
    
    interpreter = ["bash", "-c"]
  }
}

