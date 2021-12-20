terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 3.0"
    }
  }
}

# Configure the AWS Provider
provider "aws" {
  region = "eu-central-1"
}

data "aws_vpc" "default" {
  default = true
}

data "aws_subnet" "default" {
  availability_zone = "${var.AWS_REGION}a"
  vpc_id            = data.aws_vpc.default.id
}

data "aws_subnet" "b" {
  availability_zone = "${var.AWS_REGION}b"
  vpc_id            = data.aws_vpc.default.id
}

# MongoDB security group
resource "aws_security_group" "mongodb" {
  name        = "mongodb-${var.ENVIRONMENT}"
  description = "Security group for mongodb-${var.ENVIRONMENT}"
  vpc_id      = data.aws_vpc.default.id

  tags = {
    Name = "mongodb-${var.ENVIRONMENT}"
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_security_group_rule" "mongodb_allow_all" {
  type        = "egress"
  from_port   = 0
  to_port     = 0
  protocol    = "-1"
  cidr_blocks = ["0.0.0.0/0"]

  security_group_id = aws_security_group.mongodb.id
}

resource "aws_security_group_rule" "mongodb_ssh" {
  type        = "ingress"
  from_port   = 22
  to_port     = 22
  protocol    = "tcp"
  cidr_blocks = ["0.0.0.0/0"]

  security_group_id = aws_security_group.mongodb.id
}

resource "aws_security_group_rule" "mongodb_mongodb" {
  type        = "ingress"
  from_port   = 27017
  to_port     = 27017
  protocol    = "tcp"
  cidr_blocks = ["0.0.0.0/0"]

  security_group_id = aws_security_group.mongodb.id
}

resource "aws_security_group_rule" "mongodb_mongodb_replication" {
  type        = "ingress"
  from_port   = 27019
  to_port     = 27019
  protocol    = "tcp"
  cidr_blocks = ["0.0.0.0/0"]

  security_group_id = aws_security_group.mongodb.id
}

resource "aws_launch_template" "mongo" {
  name          = "mongo"
  key_name      = var.CERT_KEY
  instance_type = "t3.medium"
  image_id      = data.aws_ami.mongo_ami.id
  ebs_optimized = false

  instance_market_options {
    market_type = "spot"
    spot_options {
      max_price = var.SPOT_PRICE
    }
  }

  monitoring {
    enabled = true
  }

  tag_specifications {
    resource_type = "instance"

    tags = {
      Name = "${var.ENVIRONMENT}-mongodb"
    }
  }

  user_data = base64encode("${path.module}/mongo.sh")
}

resource "aws_instance" "mongo" {
  for_each = var.names

  #  availability_zone = "${var.AWS_REGION}a"
  subnet_id       = data.aws_subnet.default.id
  security_groups = [aws_security_group.mongodb.id]

  launch_template {
    id      = aws_launch_template.mongo.id
    version = aws_launch_template.mongo.latest_version
  }

  depends_on = [aws_launch_template.mongo]

  root_block_device {
    volume_size = 100
    volume_type = "gp2"
  }

  user_data_base64 = filebase64("${path.module}/mongo.sh")
}

resource "aws_instance" "config" {
  instance_type = "t2.small"

  #  availability_zone = "${var.AWS_REGION}a"
  subnet_id       = data.aws_subnet.b.id
  security_groups = [aws_security_group.mongodb.id]

  launch_template {
    id      = aws_launch_template.mongo.id
    version = aws_launch_template.mongo.latest_version
  }

  depends_on = [aws_launch_template.mongo]

  root_block_device {
    volume_size = 8
    volume_type = "gp2"
  }

  user_data_base64 = filebase64("${path.module}/mongo.sh")

  tags = {
    Name = "${var.ENVIRONMENT}-mongo-config"
  }
}

resource "aws_instance" "mongos" {
  instance_type = "t2.small"

  #  availability_zone = "${var.AWS_REGION}a"
  subnet_id       = data.aws_subnet.b.id
  security_groups = [aws_security_group.mongodb.id]

  launch_template {
    id      = aws_launch_template.mongo.id
    version = aws_launch_template.mongo.latest_version
  }

  depends_on = [aws_launch_template.mongo]

  root_block_device {
    volume_size = 8
    volume_type = "gp2"
  }

  user_data_base64 = filebase64("${path.module}/mongo.sh")

  tags = {
    Name = "${var.ENVIRONMENT}-mongos"
  }
}
