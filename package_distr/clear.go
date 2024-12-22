package main

import (
	"context"
	"fmt"
	"log"

	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

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

func main() {
	projectID := "custom-sylph-442912-v4" // Укажите ID вашего проекта
	zone := "europe-north1-a"             // Укажите зону
	instanceName := "ubuntu-instance"     // Название удаляемой виртуалки

	if err := deleteVM(projectID, zone, instanceName); err != nil {
		log.Fatalf("Error deleting VM: %v", err)
	} else {
		fmt.Println("Instance deletion initiated successfully.")
	}
}
