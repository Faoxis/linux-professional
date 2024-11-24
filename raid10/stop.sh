#!/bin/sh

# Настройки
VM_NAME="ansible-vm"

echo "Получаем информацию о виртуальной машине..."
VM_JSON=$(yc compute instance get --name "$VM_NAME" --format json)

if [ -z "$VM_JSON" ]; then
  echo "Виртуальная машина $VM_NAME не найдена."
  exit 0
fi

VM_ID=$(echo "$VM_JSON" | jq -r '.id')
echo "Виртуальная машина найдена: ID=$VM_ID"

# Получаем список подключённых дисков
echo "Получаем список дисков, подключённых к виртуальной машине..."
DISK_IDS=$(echo "$VM_JSON" | jq -r '.secondary_disks[].disk_id')

# Удаляем подключённые дополнительные диски
for DISK_ID in $DISK_IDS; do
  echo "Отключаем и удаляем диск: $DISK_ID"
  yc compute instance detach-disk --id "$VM_ID" --disk-id "$DISK_ID"
  yc compute disk delete --id "$DISK_ID"
done

# Удаляем загрузочный диск, если он не автоудаляемый
BOOT_DISK_ID=$(echo "$VM_JSON" | jq -r '.boot_disk.disk_id')
AUTO_DELETE=$(echo "$VM_JSON" | jq -r '.boot_disk.auto_delete')
if [ "$AUTO_DELETE" != "true" ]; then
  echo "Удаляем загрузочный диск: $BOOT_DISK_ID"
  yc compute disk delete --id "$BOOT_DISK_ID"
else
  echo "Загрузочный диск настроен на автоудаление."
fi

# Удаляем виртуальную машину
echo "Удаляем виртуальную машину..."
yc compute instance delete --id "$VM_ID"

echo "Все ресурсы успешно удалены"
