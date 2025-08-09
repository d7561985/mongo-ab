# Альтернативные конфигурации с NVMe дисками

# Вариант 1: i3en.large (2 vCPU, 1x 1.25TB NVMe) - итого 6 vCPU
instance_type = "i3en.large"
cache_size_gb = 4  # 8GB RAM

# Вариант 2: m5d.xlarge (4 vCPU, 1x 150GB NVMe) - итого 12 vCPU  
# instance_type = "m5d.xlarge"
# cache_size_gb = 8  # 16GB RAM

# Вариант 3: r5d.large (2 vCPU, 1x 75GB NVMe) - итого 6 vCPU
# instance_type = "r5d.large"  
# cache_size_gb = 8  # 16GB RAM