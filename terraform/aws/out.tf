output "public_ip" {
  value = module.init.public_ip
}

output "config_ip" {
  value = module.init.config_ip
}

output "mongos_ip" {
  value = module.init.mongos_ip
}
