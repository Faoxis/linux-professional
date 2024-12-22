package storage

import (
	"encoding/json"
	"os"

	"vm-control-app/vm"
)

const (
	vmStorageFile = "createdVM.json"
)

func Save(vm *vm.VMList) error {
	data, err := json.MarshalIndent(vm, "", " ")
	if err != nil {
		return err
	}
	err = os.WriteFile(vmStorageFile, data, 0644)
	if err != nil {
		return err
	}
	return nil
}

func LoadVMs() (*vm.VMList, error) {
	data, err := os.ReadFile(vmStorageFile)
	if err != nil {
		return nil, err
	}
	var vmList vm.VMList
	if err := json.Unmarshal(data, &vmList); err != nil {
		return nil, err
	}
	return &vmList, nil
}

// Добавление новой ВМ
func AddVM(vm *vm.CreatedVM) error {
	vms, err := LoadVMs()
	if err != nil {
		return err
	}
	vms.VMs = append(vms.VMs, vm)

	err = Save(vms)
	if err != nil {
		return err
	}

	return nil
}

// Удаление ВМ по имени
func RemoveVM(name string) error {

	vms, err := LoadVMs()
	if err != nil {
		return err
	}

	for i, vm := range vms.VMs {
		if vm.Name == name {
			vms.VMs = append(vms.VMs[:i], vms.VMs[i+1:]...)
			break
		}
	}
	return nil
}
