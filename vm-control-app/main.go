package main

import (
	"fmt"
	"github.com/google/uuid"
	"log"
	"vm-control-app/storage"
	"vm-control-app/vm"
)

func main() {
	err := createVM()
	if err != nil {
		log.Fatalf("VM cannot be created: %d", err)
	}
}

func createVM() error {
	newVm := vm.NewVM{
		Name: fmt.Sprintf("my-vm-%s", uuid.New()),
		Disks: []*vm.Disk{
			{
				Size: 10,
			},
			{
				Size: 20,
			},
			{
				Size: 10,
			},
		},
	}

	createdVm := vm.CreateVM(newVm)
	err := storage.AddVM(createdVm)
	if err != nil {
		log.Fatalf("Can't save vm stat: %v", err)
	}
	fmt.Println(createdVm)
	return nil
}
