#!/bin/bash
# Скрипт для проверки текущих spot цен

REGION="eu-central-1"
INSTANCE_TYPES=("i3en.large" "i3en.xlarge" "i3en.2xlarge" "i3en.3xlarge")

echo "Checking spot prices in $REGION..."
echo "================================="

for TYPE in "${INSTANCE_TYPES[@]}"; do
    echo -e "\n$TYPE:"
    
    # Получаем on-demand цену
    ON_DEMAND=$(aws ec2 describe-instance-types \
        --region $REGION \
        --instance-types $TYPE \
        --query 'InstanceTypes[0].InstanceType' \
        --output text 2>/dev/null)
    
    # Получаем последние spot цены
    SPOT_PRICES=$(aws ec2 describe-spot-price-history \
        --region $REGION \
        --instance-types $TYPE \
        --start-time $(date -u -d '1 hour ago' +%Y-%m-%dT%H:%M:%S) \
        --product-descriptions "Linux/UNIX" \
        --query 'SpotPriceHistory[].[SpotPrice,AvailabilityZone]' \
        --output text | sort -n | head -5)
    
    echo "$SPOT_PRICES" | while read PRICE AZ; do
        echo "  AZ: $AZ - Price: \$$PRICE/hour"
    done
    
    # Показываем среднюю цену
    AVG_PRICE=$(echo "$SPOT_PRICES" | awk '{sum+=$1; count++} END {printf "%.3f", sum/count}')
    echo "  Average: \$$AVG_PRICE/hour"
    
    # Сравнение с on-demand (примерные цены)
    case $TYPE in
        "i3en.large")   OD_PRICE=0.188 ;;
        "i3en.xlarge")  OD_PRICE=0.376 ;;
        "i3en.2xlarge") OD_PRICE=0.752 ;;
        "i3en.3xlarge") OD_PRICE=1.128 ;;
    esac
    
    if [ -n "$AVG_PRICE" ] && [ -n "$OD_PRICE" ]; then
        SAVINGS=$(awk "BEGIN {printf \"%.0f\", (1 - $AVG_PRICE / $OD_PRICE) * 100}")
        echo "  On-Demand: \$$OD_PRICE/hour"
        echo "  Savings: ~$SAVINGS%"
    fi
done

echo -e "\n================================="
echo "Recommendation for stable spot price:"
echo "Set spot_max_price to 40-50% of on-demand price"
echo "Example for i3en.2xlarge: spot_max_price = \"0.40\""