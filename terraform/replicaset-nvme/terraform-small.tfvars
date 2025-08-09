# Конфигурация для меньших инстансов (в рамках лимита 16 vCPU)

# i3en.xlarge имеет 4 vCPU, итого 12 vCPU для 3 инстансов
instance_type = "i3en.xlarge"

# Уменьшаем кэш для меньшего инстанса (16GB RAM)
cache_size_gb = 8

# Опционально: использовать spot для экономии
use_spot_instances = true
spot_max_price = "0.20"  # ~50% от on-demand для i3en.xlarge