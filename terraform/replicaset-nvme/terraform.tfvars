use_spot_instances = true
use_public_ip = true

spot_max_price = "2.00"  # Для i3en.2xlarge
allowed_ips = ["0.0.0.0/0"]

# i3en.xlarge имеет 4 vCPU, итого 12 vCPU для 3 инстансов
instance_type = "i4i.xlarge"

# Уменьшаем кэш для меньшего инстанса (16GB RAM)
cache_size_gb = 32
