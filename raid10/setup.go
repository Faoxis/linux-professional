package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sync"
)

// Структуры для обработки JSON-ответа
type NetworkInterface struct {
	PrimaryV4Address struct {
		OneToOneNat struct {
			Address string `json:"address"`
		} `json:"one_to_one_nat"`
	} `json:"primary_v4_address"`
}

type VM struct {
	ID                string             `json:"id"`
	NetworkInterfaces []NetworkInterface `json:"network_interfaces"`
}

func main() {
	const (
		VMName        = "ansible-vm"
		Zone          = "ru-central1-a"
		ImageID       = "fd87c0qpl9prjv5up7mc"
		DiskSize      = 10
		DiskCount     = 4
		SubnetName    = "yc-public"
		InventoryFile = "ansible/inventory.yml"
	)

	fmt.Println("Создаём виртуальную машину...")

	// Создание ВМ
	vmID, vmIP, err := createVM(VMName, Zone, ImageID, SubnetName)
	if err != nil {
		fmt.Printf("Ошибка при создании виртуальной машины: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Виртуальная машина создана: ID=%s, IP=%s\n", vmID, vmIP)

	// Создание дисков параллельно
	var wg sync.WaitGroup
	for i := 1; i <= DiskCount; i++ {
		wg.Add(1)
		go func(diskNum int) {
			defer wg.Done()
			diskName := fmt.Sprintf("%s-disk-%d", VMName, diskNum)
			err := createAndAttachDisk(diskName, Zone, DiskSize, vmID)
			if err != nil {
				fmt.Printf("Ошибка при создании или подключении диска %s: %v\n", diskName, err)
				os.Exit(1)
			}
			fmt.Printf("Диск %s был успешно создан и подключён\n", diskName)
		}(i)
	}
	wg.Wait()

	// Обновление inventory
	fmt.Println("Обновляем файл inventory...")
	err = updateInventoryFile(InventoryFile, VMName, vmIP)
	if err != nil {
		fmt.Printf("Ошибка при обновлении файла inventory: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Все диски подключены. Виртуальная машина готова!")
	fmt.Println("Запускаем ansible playbook...")

	// Запуск ansible playbook
	err = runAnsiblePlaybook(InventoryFile)
	if err != nil {
		fmt.Printf("Ошибка при запуске ansible playbook: %v\n", err)
		os.Exit(1)
	}
}

func createVM(vmName, zone, imageID, subnetName string) (string, string, error) {
	// Получаем домашний каталог пользователя
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", "", fmt.Errorf("не удалось получить домашний каталог: %w", err)
	}
	sshKeyPath := fmt.Sprintf("%s/.ssh/id_rsa.pub", homeDir)

	// Создаём ВМ через Yandex CLI
	cmd := exec.Command(
		"yc", "compute", "instance", "create",
		"--name", vmName,
		"--zone", zone,
		"--memory", "2",
		"--cores", "2",
		"--create-boot-disk", fmt.Sprintf("image-id=%s,size=20GB", imageID),
		"--ssh-key", sshKeyPath,
		"--network-interface", fmt.Sprintf("subnet-name=%s,nat-ip-version=ipv4", subnetName),
		"--format", "json",
	)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		fmt.Printf("Ошибка выполнения команды:\n%s\n", stderr.String())
		return "", "", err
	}

	// Обрабатываем JSON-ответ
	var vm VM
	if err := json.Unmarshal(stdout.Bytes(), &vm); err != nil {
		fmt.Printf("Ошибка парсинга JSON ответа:\n%s\n", stdout.String())
		return "", "", err
	}

	// Проверяем наличие IP
	if len(vm.NetworkInterfaces) == 0 || vm.NetworkInterfaces[0].PrimaryV4Address.OneToOneNat.Address == "" {
		return "", "", fmt.Errorf("не удалось получить IP-адрес для ВМ")
	}

	vmID := vm.ID
	vmIP := vm.NetworkInterfaces[0].PrimaryV4Address.OneToOneNat.Address
	return vmID, vmIP, nil
}

func createAndAttachDisk(diskName, zone string, diskSize int, vmID string) error {
	// Проверка существования диска
	checkCmd := exec.Command("yc", "compute", "disk", "get", "--name", diskName, "--zone", zone)
	if err := checkCmd.Run(); err == nil {
		fmt.Printf("Диск %s уже существует. Пропускаем создание.\n", diskName)
		return attachDisk(diskName, zone, vmID)
	}

	// Создание нового диска
	fmt.Printf("Создаём диск: %s...\n", diskName)
	cmd := exec.Command(
		"yc", "compute", "disk", "create",
		"--name", diskName,
		"--size", fmt.Sprintf("%d", diskSize),
		"--type", "network-hdd",
		"--zone", zone,
		"--format", "json",
	)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ошибка при создании диска %s: %w", diskName, err)
	}

	return attachDisk(diskName, zone, vmID)
}

func attachDisk(diskName, zone, vmID string) error {
	fmt.Printf("Подключаем диск %s к виртуальной машине...\n", diskName)
	cmd := exec.Command(
		"yc", "compute", "instance", "attach-disk",
		"--id", vmID,
		"--disk-name", diskName,
	)
	return cmd.Run()
}

func updateInventoryFile(fileName, vmName, vmIP string) error {
	inventory := fmt.Sprintf(`all:
  hosts:
    %s:
      ansible_host: %s
      ansible_user: yc-user
      ansible_ssh_private_key_file: ~/.ssh/id_rsa
      ansible_ssh_common_args: '-o StrictHostKeyChecking=no'
`, vmName, vmIP)

	return os.WriteFile(fileName, []byte(inventory), 0644)
}

func runAnsiblePlaybook(inventoryFile string) error {
	cmd := exec.Command("ansible-playbook", "-i", inventoryFile, "ansible/raid.yml")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
