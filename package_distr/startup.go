package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

func createVM(projectID, zone, instanceName string) (string, error) {
	ctx := context.Background()

	// Читаем публичный ключ из ~/.ssh/id_rsa.pub
	sshKey, err := os.ReadFile(os.ExpandEnv("$HOME/.ssh/id_rsa.pub"))
	if err != nil {
		return "", fmt.Errorf("failed to read SSH public key: %v", err)
	}

	// Создаем клиент для Compute Engine
	service, err := compute.NewService(ctx, option.WithCredentialsFile("../gcloud-key.json"))
	if err != nil {
		return "", fmt.Errorf("failed to create Compute Engine service: %v", err)
	}

	// Конфигурация виртуальной машины
	instance := &compute.Instance{
		Name:        instanceName,
		MachineType: fmt.Sprintf("zones/%s/machineTypes/e2-micro", zone),
		Disks: []*compute.AttachedDisk{
			{
				Boot:       true,
				AutoDelete: true,
				InitializeParams: &compute.AttachedDiskInitializeParams{
					SourceImage: "projects/ubuntu-os-cloud/global/images/family/ubuntu-2404-lts-amd64",
				},
			},
		},
		NetworkInterfaces: []*compute.NetworkInterface{
			{
				Name: "default",
				AccessConfigs: []*compute.AccessConfig{
					{
						Name: "External NAT",
						Type: "ONE_TO_ONE_NAT",
					},
				},
			},
		},
		Metadata: &compute.Metadata{
			Items: []*compute.MetadataItems{
				{
					Key:   "ssh-keys",
					Value: stringPointer(fmt.Sprintf("myuser:%s", sshKey)),
				},
			},
		},
		Tags: &compute.Tags{
			Items: []string{"http-server"}, // Добавляем тег для разрешения HTTP-трафика
		},
	}

	// Создаём ВМ
	op, err := service.Instances.Insert(projectID, zone, instance).Context(ctx).Do()
	if err != nil {
		return "", fmt.Errorf("failed to create instance: %v", err)
	}

	fmt.Printf("Instance creation started. Operation: %s\n", op.Name)

	// Ждём, пока появится внешний IP
	fmt.Println("Waiting for external IP to be assigned...")
	var externalIP string
	for i := 0; i < 10; i++ { // Максимум 10 попыток
		time.Sleep(5 * time.Second) // Ожидаем 5 секунд перед каждой проверкой

		instanceInfo, err := service.Instances.Get(projectID, zone, instanceName).Context(ctx).Do()
		if err != nil {
			return "", fmt.Errorf("failed to get instance info: %v", err)
		}

		// Проверяем, есть ли внешний IP
		for _, ni := range instanceInfo.NetworkInterfaces {
			for _, ac := range ni.AccessConfigs {
				if ac.NatIP != "" {
					externalIP = ac.NatIP
					break
				}
			}
		}

		if externalIP != "" {
			fmt.Printf("External IP assigned: %s\n", externalIP)
			return externalIP, nil
		}

		fmt.Println("External IP not found yet. Retrying...")
	}

	return "", fmt.Errorf("external IP not assigned after multiple attempts")
}

func waitForAnsiblePing(ip string, timeout time.Duration) error {
	fmt.Println("Waiting for Ansible ping to succeed...")
	start := time.Now()

	for time.Since(start) < timeout {
		// Запускаем ansible ping
		cmd := exec.Command("ansible", "all", "-i", "ansible/inventory", "-m", "ping")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err := cmd.Run()
		if err == nil {
			fmt.Println("Ansible ping succeeded.")
			return nil
		}

		fmt.Println("Ansible ping failed. Retrying...")
		time.Sleep(10 * time.Second) // Ждём 10 секунд перед повторной попыткой
	}

	return fmt.Errorf("Ansible ping failed after %s", timeout)
}

func runAnsiblePlaybook() error {
	fmt.Println("Running Ansible playbook...")

	// Выполнение команды ansible-playbook
	cmd := exec.Command("ansible-playbook", "-i", "ansible/inventory", "ansible/playbook.yml")
	cmd.Stdout = os.Stdout // Вывод команд Ansible в консоль
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run Ansible playbook: %v", err)
	}

	fmt.Println("Ansible playbook executed successfully.")
	return nil
}

func stringPointer(s string) *string {
	return &s
}

func rewriteInventoryFile(ip string) error {
	// Генерируем Ansible inventory
	inventory := fmt.Sprintf("[google_vm]\n%s ansible_user=myuser ansible_ssh_private_key_file=~/.ssh/id_rsa ansible_ssh_common_args='-o StrictHostKeyChecking=no -o ConnectTimeout=60'\n", ip)

	// Создаём или открываем файл для записи
	inventoryFilePath := "ansible/inventory"

	if err := os.Remove(inventoryFilePath); err != nil {
		log.Printf("%v", err)
	}

	file, err := os.Create(inventoryFilePath)
	if err != nil {
		log.Fatalf("Error creating inventory file: %v", err)
	}
	defer file.Close() // Гарантированное закрытие файла

	// Записываем данные в файл
	_, err = file.WriteString(inventory)
	if err != nil {
		return err
	}

	// Принудительно сохраняем данные на диск
	err = file.Sync()
	if err != nil {
		return err
	}

	fmt.Println("Inventory created and synced successfully.")
	return nil
}

func main() {
	projectID := "custom-sylph-442912-v4"
	zone := "europe-north1-a"
	instanceName := "ubuntu-instance"

	// Создаем ВМ и получаем её IP
	externalIP, err := createVM(projectID, zone, instanceName)
	if err != nil {
		log.Fatalf("Error creating VM: %v", err)
	}

	// Пишем inventory file
	if err := rewriteInventoryFile(externalIP); err != nil {
		log.Fatalf("Error rewriting inventory file: %v", err)
	}

	// Ждём, пока SSH станет доступным
	if err := waitForAnsiblePing(externalIP, 3*time.Minute); err != nil {
		log.Fatalf("Error waiting for SSH: %v", err)
	}

	// Запускаем Ansible playbook
	if err := runAnsiblePlaybook(); err != nil {
		log.Fatalf("Error running Ansible playbook: %v", err)
	}
}
