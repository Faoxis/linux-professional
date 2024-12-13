package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	// "encoding/json"
)

type NetworkInterface struct {
	PrimaryV4Address struct {
		OneToOne struct {
			Address string `json:"address"`
		} `json:"one_to_one_nat`
	} `json:"primary_v4_address`
}

type VM struct {
	ID                 string             `json: "id"`
	NetworkInterfactes []NetworkInterface `json:"network_interfaces"`
}

func main() {
	const (
		VMName     = "ansible-demo-2"
		Zone       = "ru-central1-a"
		ImageId    = "fd87c0qpl9prjv5up7mc"
		DiskSize   = 10
		SubnetName = "yc-public"
	)

	fmt.Println("Получаем информацию о виртуальной машине...")
	vmID, vmIP, err := createVM(VMName, Zone, ImageId, SubnetName)
	if err != nil {
		fmt.Printf("Ошибка при получении информации о виртуальное машине: %v", err)
		os.Exit(1)
	}

	fmt.Printf("vmID: %s\n", vmID)
	fmt.Printf("vmIP: %s\n", vmIP)

}

func createVM(vmName, zone, imageID, subnetName string) (string, string, error) {
	homeDir, err := os.UserHomeDir()
	sshKeyPath := fmt.Sprintf("%s/.ssh/id_rsa.pub", homeDir)

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

	var vm VM
	if err := json.Unmarshal(stdout.Bytes(), &vm); err != nil {
		return "", "", err
	}

	vmIP := vm.NetworkInterfactes[0].PrimaryV4Address.OneToOne.Address
	return vm.ID, vmIP, nil
}
