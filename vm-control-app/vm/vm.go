package vm

import "log"

type Disk struct {
	Size int64
}

type NewVM struct {
	Disks []*Disk
	Name  string
}

type CreatedVM struct {
	Ip   string `json:"ip"`
	Name string `json:"name"`
}

type VMList struct {
	VMs []*CreatedVM `json:"vms"`
}

const (
	projectId = "custom-sylph-442912-v4"
	zone      = "europe-north1-a"
)

func CreateVM(vm NewVM) *CreatedVM {
	ip, err := createVM(
		projectId,
		zone,
		vm.Name,
		vm.Disks,
	)
	if err != nil {
		panic(err)
	}

	return &CreatedVM{
		Ip:   ip,
		Name: vm.Name,
	}
}

func clearVms(vms []CreatedVM) error {
	for _, vm := range vms {
		err := deleteVM(projectId, zone, vm.Name)
		if err != nil {
			log.Printf("Can't delete vm %s", vm)
		}
	}
	return nil
}
