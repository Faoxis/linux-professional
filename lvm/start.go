package main

import (
	// "context"
	"fmt"
	// "log"
	// "strconv"
	// "google.golang.org/api/compute/v1"
)

// func createInstance(projectID, zone, instanceName, machineType, diskSizeGb string) error {
// 	ctx := context.Background()
// 	computeService, err := compute.NewService(ctx)
// 	if err != nil {
// 		return fmt.Errorf("failed to create compute service: %v", err)
// 	}

// 	// Преобразуем diskSizeGb из string в int64
// 	diskSize, err := strconv.ParseInt(diskSizeGb, 10, 64)
// 	if err != nil {
// 		return fmt.Errorf("invalid disk size: %v", err)
// 	}

// 	instance := &compute.Instance{
// 		Name:        instanceName,
// 		MachineType: fmt.Sprintf("zones/%s/machineTypes/%s", zone, machineType),
// 		Disks: []*compute.AttachedDisk{
// 			{
// 				Boot:       true,
// 				AutoDelete: true,
// 				InitializeParams: &compute.AttachedDiskInitializeParams{
// 					SourceImage: "projects/ubuntu-os-cloud/global/images/ubuntu-2404-noble-amd64-v20241115",
// 					DiskSizeGb:  diskSize,
// 				},
// 			},
// 		},
// 		NetworkInterfaces: []*compute.NetworkInterface{
// 			{
// 				AccessConfigs: []*compute.AccessConfig{
// 					{Type: "ONE_TO_ONE_NAT", Name: "External NAT"},
// 				},
// 				Network: "global/networks/default",
// 			},
// 		},
// 	}

// 	_, err = computeService.Instances.Insert(projectID, zone, instance).Context(ctx).Do()
// 	if err != nil {
// 		return fmt.Errorf("failed to create instance: %v", err)
// 	}

// 	log.Printf("Instance %s created successfully in project %s, zone %s", instanceName, projectID, zone)
// 	return nil
// }

func main() {
	arr := [6]int{1, 2, 3, 4, 5, 6}
	s := arr[1:6]
	fmt.Println(len(s)) // 2 (только 2 элемента)
	fmt.Println(cap(s)) // 5 (можно захватить ещё 5 элементов: [2, 3, 4, 5, 6])

	s = append(s, 10, 20) // добавляем 2 элемента
	fmt.Println(s)        // [2, 3, 10, 20]
	fmt.Println(arr)

	// projectID := "custom-sylph-442912-v4"
	// zone := "us-central1-b"
	// instanceName := "lvm-instance"
	// machineType := "n1-standard-1"
	// diskSizeGb := "10"

	// // Create instance
	// if err := createInstance(projectID, zone, instanceName, machineType, diskSizeGb); err != nil {
	// 	log.Fatalf("Error creating instance: %v", err)
	// }
}
