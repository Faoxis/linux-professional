#!/bin/sh

# Настройки
VM_NAME="ansible-vm"
ZONE="ru-central1-a"
IMAGE_ID="fd87c0qpl9prjv5up7mc"  # ubuntu-24-04-lts-v20240812
DISK_SIZE=10  # Размер дополнительных дисков в ГБ
DISK_COUNT=4  # Количество дополнительных дисков
SUBNET_NAME="yc-public"  # Имя подсети
INVENTORY_FILE="ansible/inventory.yml"  # Файл inventory для Ansible

# Создание ВМ
echo "Создаём виртуальную машину..."
VM_JSON=$(yc compute instance create \
  --name "$VM_NAME" \
  --zone "$ZONE" \
  --memory 2 \
  --cores 2 \
  --create-boot-disk image-id="$IMAGE_ID",size=20GB \
  --ssh-key ~/.ssh/id_rsa.pub \
  --network-interface subnet-name="$SUBNET_NAME",nat-ip-version=ipv4 \
  --format json)

# Проверка успешности создания ВМ
if [ $? -ne 0 ]; then
  echo "Ошибка при создании виртуальной машины."
  exit 1
fi

VM_ID=$(echo "$VM_JSON" | jq -r '.id')
VM_IP=$(echo "$VM_JSON" | jq -r '.network_interfaces[0].primary_v4_address.one_to_one_nat.address')

echo "Виртуальная машина создана: ID=$VM_ID, IP=$VM_IP"

# Создание дополнительных дисков
for i in $(seq 1 "$DISK_COUNT"); do
  DISK_NAME="${VM_NAME}-disk-$i"
  echo "Создаём диск: $DISK_NAME..."
  
  # Проверяем, существует ли диск
  EXISTING_DISK=$(yc compute disk get --name "$DISK_NAME" --zone "$ZONE" 2>/dev/null)
  if [ $? -eq 0 ]; then
    echo "Диск $DISK_NAME уже существует. Пропускаем создание."
    DISK_ID=$(echo "$EXISTING_DISK" | jq -r '.id')
  else
    DISK_JSON=$(yc compute disk create \
      --name "$DISK_NAME" \
      --size "$DISK_SIZE" \
      --type network-hdd \
      --zone "$ZONE" \
      --format json)

    # Проверка успешности создания диска
    if [ $? -ne 0 ]; then
      echo "Ошибка при создании диска $DISK_NAME."
      exit 1
    fi

    DISK_ID=$(echo "$DISK_JSON" | jq -r '.id')
  fi

  echo "Подключаем диск $DISK_NAME к виртуальной машине..."
  yc compute instance attach-disk --id "$VM_ID" --disk-id "$DISK_ID"

  # Проверка успешности подключения диска
  if [ $? -ne 0 ]; then
    echo "Ошибка при подключении диска $DISK_NAME к виртуальной машине."
    exit 1
  fi
done


# Обновление inventory
echo "Обновляем файл inventory: $INVENTORY_FILE"
cat > "$INVENTORY_FILE" <<EOF
all:
  hosts:
    $VM_NAME:
      ansible_host: $VM_IP
      ansible_user: yc-user
      ansible_ssh_private_key_file: ~/.ssh/id_rsa
      ansible_ssh_common_args: '-o StrictHostKeyChecking=no'
EOF

echo "Файл inventory обновлён:"
cat "$INVENTORY_FILE"

echo "Все диски подключены. Виртуальная машина готова!"

echo "Запускаем ansible скрипты"
ansible-playbook -i ansible/inventory.yml ansible/raid.yml
