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

module "init" {
  source = "./modules/init"
  useNvME = var.useNvME
}


#data "terraform_remote_state" "init" {
#  backend = "local"
#
#  config = {
#    path = "./modules/init/terraform.tfstate"
#  }
#}

locals {
  initJs = "/home/ec2-user/init.js"

  cfg = merge(
  {for id, host in module.init.public_ip :id => {
    host : host,
    clusterRole : "shardsvr",
    ssd: var.useNvME
  }},
  {
    "rs0" : {
      host : module.init.config_ip
      clusterRole : "configsvr",
      ssd: false
    }
  },
  )
}

resource "null_resource" "upload_instances" {
  for_each = local.cfg

  connection {
    type  = "ssh"
    user  = "ec2-user"
    host  = each.value.host
    agent = true
  }

  provisioner "file" {
    destination = "/home/ec2-user/mongod.conf"
    content     = templatefile("${path.module}/mongod.tpl", {
      id : each.key
      clusterRole : each.value.clusterRole
    })
  }

  # port 27018 for shardsvr
  provisioner "file" {
    destination = local.initJs
    content     = templatefile("${path.module}/replica-init.tpl", {
      id : each.key
      host : each.value.host
    })
  }

  provisioner "file" {
    source      = "${path.module}/mongo.sh"
    destination = "/home/ec2-user/mongo.sh"
  }

  provisioner "remote-exec" {
    inline = [
      "sudo mkdir /data",

      # mount ssd
      "sudo mkfs -t xfs /dev/nvme0n1",
      "sudo mount /dev/nvme0n1 /data",

      # setup mongo server
      "sudo chmod 0700 ./mongo.sh", "sudo ./mongo.sh",
      "sudo mkdir /data/mongodb",
      "sudo chmod 0755 /var/run/mongodb",
      "sudo chown mongod:mongod /data/mongodb",
    ]
  }
}

# only after installation
resource "null_resource" "execute" {
  for_each = local.cfg

  provisioner "remote-exec" {
    inline = [
      "sudo cp /home/ec2-user/mongod.conf /etc/mongod.conf",
      "sudo systemctl start mongod",
      "sudo systemctl status mongod",
      # init replicaset
      "mongosh  --quiet ${local.initJs}"
    ]

    connection {
      type  = "ssh"
      user  = "ec2-user"
      host  = each.value.host
      agent = true
    }
  }

  depends_on = [null_resource.upload_instances]
}

# mongos part
resource "null_resource" "upload_mongos" {
  connection {
    type  = "ssh"
    user  = "ec2-user"
    host  = module.init.mongos_ip
    agent = true
  }

  provisioner "file" {
    source      = "${path.module}/mongos.service"
    destination = "/home/ec2-user/mongos.service"
  }

  provisioner "file" {
    destination = "/home/ec2-user/mongod.conf"
    content     = templatefile("${path.module}/mongos.tpl", {
      configDB : "rs0/${module.init.config_ip}:27017"
    })
  }

  provisioner "file" {
    destination = local.initJs
    content     = templatefile("${path.module}/mongos-init.tpl", {
      val : module.init.public_ip
      shardDB: var.shardDB
    })
  }

  provisioner "file" {
    source      = "${path.module}/mongo.sh"
    destination = "/home/ec2-user/mongo.sh"
  }

  provisioner "remote-exec" {
    inline = [
      "sudo chmod 0700 ./mongo.sh", "sudo ./mongo.sh",
    ]
  }
}

resource "null_resource" "execute_mongos" {
  provisioner "remote-exec" {
    inline = [
      "sudo cp /home/ec2-user/mongod.conf /etc/mongod.conf",
      "sudo cp /home/ec2-user/mongos.service /etc/systemd/system/mongos.service",
      "sudo systemctl start mongos",
      "sudo systemctl status mongos",
      # init replicaset
      "mongosh  --quiet ${local.initJs}"
    ]

    connection {
      type  = "ssh"
      user  = "ec2-user"
      host  = module.init.mongos_ip
      agent = true
    }
  }

  depends_on = [null_resource.execute]
}