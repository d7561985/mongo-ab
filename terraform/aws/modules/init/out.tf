output "public_ip" {
  value = {for i, v in aws_instance.mongo: i => v.public_ip}
}

output "config_ip" {
  value = aws_instance.config.public_ip
}
