variable "instance_type" {
  description = "EC2 instance type with NVMe disks"
  type        = string
  default     = "i3en.2xlarge"
  
  validation {
    condition = contains([
      # i3en family
      "i3en.large",    # 1x 1.25 TB NVMe SSD
      "i3en.xlarge",   # 1x 1.25 TB NVMe SSD
      "i3en.2xlarge",  # 2x 1.25 TB NVMe SSD
      "i3en.3xlarge",  # 1x 2.5 TB NVMe SSD
      "i3en.6xlarge",  # 2x 2.5 TB NVMe SSD
      "i3en.12xlarge", # 4x 2.5 TB NVMe SSD
      "i3en.24xlarge", # 8x 2.5 TB NVMe SSD
      # i4i family (новое поколение)
      "i4i.large",     # 1x 468 GB NVMe SSD
      "i4i.xlarge",    # 1x 937 GB NVMe SSD
      "i4i.2xlarge",   # 1x 1875 GB NVMe SSD
      "i4i.4xlarge",   # 1x 3750 GB NVMe SSD
      "i4i.8xlarge",   # 2x 3750 GB NVMe SSD
      "i4i.16xlarge",  # 4x 3750 GB NVMe SSD
      "i4i.32xlarge",  # 8x 3750 GB NVMe SSD
      # Другие инстансы с NVMe
      "m5d.large",     # 1x 75 GB NVMe SSD
      "m5d.xlarge",    # 1x 150 GB NVMe SSD
      "m5d.2xlarge",   # 1x 300 GB NVMe SSD
      "r5d.large",     # 1x 75 GB NVMe SSD
      "r5d.xlarge",    # 1x 150 GB NVMe SSD
      "r5d.2xlarge"    # 1x 300 GB NVMe SSD
    ], var.instance_type)
    error_message = "Instance type must have NVMe disks (i3en, i4i, m5d, or r5d families)."
  }
}

variable "mongodb_port" {
  description = "Port for MongoDB"
  type        = number
  default     = 50000
}

variable "allowed_ips" {
  description = "List of IPs allowed to connect to MongoDB (CIDR format)"
  type        = list(string)
  default     = ["0.0.0.0/0"]  # Открыто для всех - измените для безопасности!
}

variable "ssh_public_key_path" {
  description = "Path to SSH public key"
  type        = string
  default     = "~/.ssh/id_rsa.pub"
}

variable "mongodb_version" {
  description = "MongoDB version to install"
  type        = string
  default     = "7.0"
}

variable "replica_set_name" {
  description = "Name of the replica set"
  type        = string
  default     = "rs0"
}

variable "cache_size_gb" {
  description = "WiredTiger cache size in GB"
  type        = number
  default     = 16  # 50% от 32GB RAM для i3en.2xlarge
}

variable "oplog_size_mb" {
  description = "Oplog size in MB"
  type        = number
  default     = 10240  # 10GB для высокой нагрузки записи
}

variable "enable_monitoring" {
  description = "Enable CloudWatch monitoring"
  type        = bool
  default     = true
}

variable "backup_retention_days" {
  description = "Number of days to retain backups"
  type        = number
  default     = 7
}

variable "tags" {
  description = "Tags to apply to all resources"
  type        = map(string)
  default = {
    Project     = "MongoDB-Performance-Test"
    Environment = "Production"
    Terraform   = "true"
  }
}

variable "use_spot_instances" {
  description = "Use spot instances instead of on-demand"
  type        = bool
  default     = false
}

variable "spot_max_price" {
  description = "Maximum spot price (empty string for on-demand price)"
  type        = string
  default     = ""  # Текущая on-demand цена
}

variable "enable_health_check" {
  description = "Enable automatic MongoDB health check after deployment"
  type        = bool
  default     = true
}

variable "health_check_delay" {
  description = "Seconds to wait before health check (default 3 minutes)"
  type        = number
  default     = 180
}

variable "auto_init_replica_set" {
  description = "Automatically initialize replica set after deployment"
  type        = bool
  default     = true
}

variable "use_public_ip" {
  description = "Use public IPs for replica set configuration (needed for internet access)"
  type        = bool
  default     = true  # Изменено для доступа через интернет
}