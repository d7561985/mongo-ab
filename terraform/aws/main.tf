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

#module "init" {
#  source = "./modules/init"
#}


data "terraform_remote_state" "init" {
  backend = "local"

  config = {
    path = "./modules/init/terraform.tfstate"
  }
}

locals {
  initJs = "/home/ec2-user/init.js"
}

locals {
  cfg = merge(
  {for id, host in data.terraform_remote_state.init.outputs.public_ip :id => {
    host : host,
    clusterRole : "shardsvr",
  }},
  {
    "rs0" : {
      host : data.terraform_remote_state.init.outputs.config_ip
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
