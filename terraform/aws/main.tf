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
  }},
  {
    "rs0" : {
      host : module.init.config_ip
      clusterRole : "configsvr",
    }
  },
  )
}

resource "null_resource" "upload_instances" {
  for_each = local.cfg

  provisioner "file" {
    destination = "/home/ec2-user/mongod.conf"
    content     = templatefile("${path.module}/mongod.tpl", {
      id : each.key
      clusterRole : each.value.clusterRole
    })

    connection {
      type  = "ssh"
      user  = "ec2-user"
      host  = each.value.host
      agent = true
    }
  }

  # port 27018 for shardsvr
  provisioner "file" {
    destination = local.initJs
    content     = templatefile("${path.module}/replica-init.tpl", {
      id : each.key
      host : each.value.host
    })

    connection {
      type  = "ssh"
      user  = "ec2-user"
      host  = each.value.host
      agent = true
    }
  }

  provisioner "file" {
    source      = "${path.module}/mongo.sh"
    destination = "/home/ec2-user/mongo.sh"

    connection {
      type  = "ssh"
      user  = "ec2-user"
      host  = each.value.host
      agent = true
    }
  }

  provisioner "remote-exec" {
    inline = [
      "sudo chmod 0700 ./mongo.sh", "sudo ./mongo.sh"
    ]

    connection {
      type  = "ssh"
      user  = "ec2-user"
      host  = each.value.host
      agent = true
    }
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
  provisioner "file" {
    source      = "${path.module}/mongos.service"
    destination = "/home/ec2-user/mongos.service"

    connection {
      type  = "ssh"
      user  = "ec2-user"
      host  = module.init.mongos_ip
      agent = true
    }
  }

  provisioner "file" {
    destination = "/home/ec2-user/mongod.conf"
    content     = templatefile("${path.module}/mongos.tpl", {
      configDB : "rs0/${module.init.config_ip}:27017"
    })

    connection {
      type  = "ssh"
      user  = "ec2-user"
      host  = module.init.mongos_ip
      agent = true
    }
  }

  provisioner "file" {
    destination = local.initJs
    content     = templatefile("${path.module}/mongos-init.tpl", {
      val : module.init.public_ip
      shardDB: var.shardDB
    })

    connection {
      type  = "ssh"
      user  = "ec2-user"
      host  = module.init.mongos_ip
      agent = true
    }
  }

  provisioner "file" {
    source      = "${path.module}/mongo.sh"
    destination = "/home/ec2-user/mongo.sh"

    connection {
      type  = "ssh"
      user  = "ec2-user"
      host  = module.init.mongos_ip
      agent = true
    }
  }

  provisioner "remote-exec" {
    inline = [
      "sudo chmod 0700 ./mongo.sh", "sudo ./mongo.sh",
    ]

    connection {
      type  = "ssh"
      user  = "ec2-user"
      host  = module.init.mongos_ip
      agent = true
    }
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
