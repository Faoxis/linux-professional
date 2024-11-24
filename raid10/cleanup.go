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
type VM struct {
	ID       string `json:"id"`
	BootDisk struct {
		DiskID     string `json:"disk_id"`
		AutoDelete bool   `json:"auto_delete"`
	} `json:"boot_disk"`
	SecondaryDisks []struct {
		DiskID string `json:"disk_id"`
	} `json:"secondary_disks"`
}

func main() {
	const VMName = "ansible-vm"

	fmt.Println("Получаем информацию о виртуальной машине...")
	vm, err := getVM(VMName)
	if err != nil {
		fmt.Printf("Ошибка при получении информации о виртуальной машине: %v\n", err)
		os.Exit(1)
	}

	if vm.ID == "" {
		fmt.Printf("Виртуальная машина %s не найдена.\n", VMName)
		return
	}

	fmt.Printf("Виртуальная машина найдена: ID=%s\n", vm.ID)

	// Удаляем дополнительные диски параллельно
	var wg sync.WaitGroup
	for _, disk := range vm.SecondaryDisks {
		wg.Add(1)
		go func(diskID string) {
			defer wg.Done()
			fmt.Printf("Отключаем и удаляем диск: %s\n", diskID)
			if err := detachAndDeleteDisk(vm.ID, diskID); err != nil {
				fmt.Printf("Ошибка при удалении диска %s: %v\n", diskID, err)
			} else {
				fmt.Printf("Диск %s успешно удалён\n", diskID)
			}
		}(disk.DiskID)
	}
	wg.Wait()

	// Удаляем загрузочный диск, если он не автоудаляемый
	if !vm.BootDisk.AutoDelete {
		fmt.Printf("Удаляем загрузочный диск: %s\n", vm.BootDisk.DiskID)
		if err := deleteDisk(vm.BootDisk.DiskID); err != nil {
			fmt.Printf("Ошибка при удалении загрузочного диска: %v\n", err)
		} else {
			fmt.Printf("Загрузочный диск %s успешно удалён\n", vm.BootDisk.DiskID)
		}
	} else {
		fmt.Println("Загрузочный диск настроен на автоудаление.")
	}

	// Удаляем виртуальную машину
	fmt.Println("Удаляем виртуальную машину...")
	if err := deleteVM(vm.ID); err != nil {
		fmt.Printf("Ошибка при удалении виртуальной машины: %v\n", err)
	} else {
		fmt.Println("Виртуальная машина успешно удалена.")
	}

	fmt.Println("Все ресурсы успешно удалены.")
}

func getVM(vmName string) (*VM, error) {
	cmd := exec.Command(
		"yc", "compute", "instance", "get",
		"--name", vmName,
		"--format", "json",
	)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения команды: %s", stderr.String())
	}

	var vm VM
	if err := json.Unmarshal(stdout.Bytes(), &vm); err != nil {
		return nil, fmt.Errorf("ошибка парсинга JSON: %w", err)
	}

	return &vm, nil
}

func detachAndDeleteDisk(vmID, diskID string) error {
	// Отключаем диск
	detachCmd := exec.Command("yc", "compute", "instance", "detach-disk", "--id", vmID, "--disk-id", diskID)
	var detachErr bytes.Buffer
	detachCmd.Stderr = &detachErr
	if err := detachCmd.Run(); err != nil {
		return fmt.Errorf("ошибка при отключении диска: %s", detachErr.String())
	}

	// Удаляем диск
	return deleteDisk(diskID)
}

func deleteDisk(diskID string) error {
	deleteCmd := exec.Command("yc", "compute", "disk", "delete", "--id", diskID)
	var deleteErr bytes.Buffer
	deleteCmd.Stderr = &deleteErr
	if err := deleteCmd.Run(); err != nil {
		return fmt.Errorf("ошибка при удалении диска: %s", deleteErr.String())
	}
	return nil
}

func deleteVM(vmID string) error {
	cmd := exec.Command("yc", "compute", "instance", "delete", "--id", vmID)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ошибка при удалении виртуальной машины: %s", stderr.String())
	}
	return nil
}
