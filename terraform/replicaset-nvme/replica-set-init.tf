# Автоматическая инициализация Replica Set
resource "null_resource" "replica_set_init" {
  count = var.auto_init_replica_set ? 1 : 0
  
  depends_on = [
    aws_instance.mongo_node,
    aws_eip.mongo_eip,
    null_resource.mongodb_health_check
  ]

  # Триггер для повторной инициализации при изменении нод
  triggers = {
    node_ids = join(",", aws_instance.mongo_node[*].id)
  }

  # Ждем дополнительно 30 секунд после health check
  provisioner "local-exec" {
    command = "sleep 30"
  }

  # Инициализация replica set
  provisioner "remote-exec" {
    connection {
      type        = "ssh"
      user        = "ec2-user"
      host        = aws_eip.mongo_eip[0].public_ip
      private_key = file(replace(var.ssh_public_key_path, ".pub", ""))
      timeout     = "5m"
    }

    inline = [
      "echo 'Initializing MongoDB Replica Set with ${var.use_public_ip ? "PUBLIC" : "PRIVATE"} IPs...'",
      "mongosh --port ${var.mongodb_port} --eval 'rs.initiate({_id: \"${var.replica_set_name}\", members: [{_id: 0, host: \"${var.use_public_ip ? aws_eip.mongo_eip[0].public_ip : aws_instance.mongo_node[0].private_ip}:${var.mongodb_port}\", priority: 2}, {_id: 1, host: \"${var.use_public_ip ? aws_eip.mongo_eip[1].public_ip : aws_instance.mongo_node[1].private_ip}:${var.mongodb_port}\", priority: 1}, {_id: 2, host: \"${var.use_public_ip ? aws_eip.mongo_eip[2].public_ip : aws_instance.mongo_node[2].private_ip}:${var.mongodb_port}\", priority: 1}], settings: {writeConcernMajorityJournalDefault: false, electionTimeoutMillis: 10000, heartbeatIntervalMillis: 2000, catchUpTimeoutMillis: 30000}})'",
      "sleep 10",
      "mongosh --port ${var.mongodb_port} --eval 'rs.status()'",
      "echo 'Connection string: mongodb://${join(",", [for eip in aws_eip.mongo_eip : "${eip.public_ip}:${var.mongodb_port}"])}/?replicaSet=${var.replica_set_name}'"
    ]
  }
}

# Output для проверки статуса replica set
output "replica_set_status_command" {
  value = "mongosh --host ${aws_eip.mongo_eip[0].public_ip}:${var.mongodb_port} --eval 'rs.status()'"
  description = "Command to check replica set status"
  depends_on = [null_resource.replica_set_init]
}

output "replica_set_initialized" {
  value = var.auto_init_replica_set ? "Replica set should be automatically initialized. Connection string: mongodb://${join(",", [for eip in aws_eip.mongo_eip : "${eip.public_ip}:${var.mongodb_port}"])}/?replicaSet=${var.replica_set_name}" : "Automatic replica set initialization disabled. Initialize manually."
  description = "Replica set initialization status"
  depends_on = [null_resource.replica_set_init]
}