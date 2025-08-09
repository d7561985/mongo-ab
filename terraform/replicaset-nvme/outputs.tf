output "connection_string" {
  value = "mongodb://${join(",", [for eip in aws_eip.mongo_eip : "${eip.public_ip}:${var.mongodb_port}"])}/?replicaSet=${var.replica_set_name}"
  description = "MongoDB connection string for replica set"
}

output "connection_string_private" {
  value = "mongodb://${join(",", [for instance in aws_instance.mongo_node : "${instance.private_ip}:${var.mongodb_port}"])}/?replicaSet=${var.replica_set_name}"
  description = "MongoDB connection string using private IPs (for VPC internal access)"
}

output "node_details" {
  value = {
    for idx, eip in aws_eip.mongo_eip :
    "node_${idx + 1}" => {
      public_ip   = eip.public_ip
      private_ip  = aws_instance.mongo_node[idx].private_ip
      instance_id = aws_instance.mongo_node[idx].id
      ssh_command = "ssh -i ${var.ssh_public_key_path} ec2-user@${eip.public_ip}"
    }
  }
  description = "Detailed information about each node"
  sensitive   = false
}

output "init_commands" {
  value = {
    step1_copy_script = "scp -i ${replace(var.ssh_public_key_path, ".pub", "")} ${path.module}/init-replicaset.sh ec2-user@${aws_eip.mongo_eip[0].public_ip}:~/"
    step2_ssh_to_primary = "ssh -i ${replace(var.ssh_public_key_path, ".pub", "")} ec2-user@${aws_eip.mongo_eip[0].public_ip}"
    step3_init_replicaset = "chmod +x init-replicaset.sh && ./init-replicaset.sh ${join(" ", [for eip in aws_eip.mongo_eip : eip.public_ip])}"
    step4_check_status = "mongosh --host ${aws_eip.mongo_eip[0].public_ip}:${var.mongodb_port} --eval 'rs.status()'"
  }
  description = "Commands to initialize the replica set"
}

output "test_command" {
  value = "./mongo-ab mongo --addr \"${format("mongodb://%s/?replicaSet=%s", join(",", [for eip in aws_eip.mongo_eip : "${eip.public_ip}:${var.mongodb_port}"]), var.replica_set_name)}\" --threads 100 --operation tx"
  description = "Command to run performance test"
}

output "security_group_id" {
  value       = aws_security_group.mongo_sg.id
  description = "Security group ID for adding additional rules"
}

output "vpc_id" {
  value       = aws_vpc.mongo_vpc.id
  description = "VPC ID where MongoDB is deployed"
}

output "monitoring_commands" {
  value = {
    mongostat = "mongostat --host ${join(",", [for eip in aws_eip.mongo_eip : "${eip.public_ip}:${var.mongodb_port}"])}"
    mongotop  = "mongotop --host ${aws_eip.mongo_eip[0].public_ip}:${var.mongodb_port}"
    rs_status = "mongosh --host ${aws_eip.mongo_eip[0].public_ip}:${var.mongodb_port} --eval 'rs.status()'"
  }
  description = "Useful monitoring commands"
}

output "estimated_monthly_cost" {
  value = format("$%.2f USD (3x %s @ $%.3f/hour)", 
    3 * 0.752 * 24 * 30,  # Цена для i3en.2xlarge в eu-central-1
    var.instance_type,
    0.752
  )
  description = "Estimated monthly cost for EC2 instances only"
}

output "health_check_note" {
  value = "MongoDB health check will run automatically after ~3 minutes. If deployment fails, check the error messages above."
  description = "Information about automatic health checks"
}