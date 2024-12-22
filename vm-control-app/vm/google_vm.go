package vm

import (
	"context"
	"fmt"
	"os"
	"time"

	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

func stringPointer(s string) *string {
	return &s
}

func createVM(projectID, zone, instanceName string, disks []*Disk) (string, error) {
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

	instance := &compute.Instance{
		Name: instanceName,
		//MachineType: fmt.Sprintf("zones/%s/machineTypes/e2-standard-2", zone),
		MachineType: fmt.Sprintf("zones/%s/machineTypes/n2-standard-4", zone),
		Disks: []*compute.AttachedDisk{
			{
				Boot:       true,
				AutoDelete: true,
				InitializeParams: &compute.AttachedDiskInitializeParams{
					SourceImage: "projects/ubuntu-os-cloud/global/images/family/ubuntu-2404-lts-amd64",
					DiskSizeGb:  50, // Основной диск
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

	var computeDisks []*compute.AttachedDisk
	for _, disk := range disks {
		computeDisks = append(computeDisks, mapDisk(disk, zone))
	}
	instance.Disks = append(instance.Disks, computeDisks...)

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

func mapDisk(disk *Disk, zone string) *compute.AttachedDisk {
	return &compute.AttachedDisk{
		AutoDelete: true,
		InitializeParams: &compute.AttachedDiskInitializeParams{
			DiskSizeGb: disk.Size,
			DiskType:   fmt.Sprintf("zones/%s/diskTypes/pd-standard", zone),
		},
	}
}

func deleteVM(projectID, zone, instanceName string) error {
	ctx := context.Background()

	// Создаем клиент для Compute Engine с аутентификацией
	service, err := compute.NewService(ctx, option.WithCredentialsFile("../gcloud-key.json"))
	if err != nil {
		return fmt.Errorf("failed to create Compute Engine service: %v", err)
	}

	// Удаляем виртуальную машину
	op, err := service.Instances.Delete(projectID, zone, instanceName).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to delete instance: %v", err)
	}

	fmt.Printf("Instance deletion started. Operation: %s\n", op.Name)
	return nil
}
