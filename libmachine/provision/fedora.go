package provision

import (
	"github.com/thelonelyghost/p2box/libmachine/drivers"
)

func init() {
	Register("Fedora", &RegisteredProvisioner{
		New: NewFedoraProvisioner,
	})
}

func NewFedoraProvisioner(d drivers.Driver) Provisioner {
	return &FedoraProvisioner{
		NewRedHatProvisioner("fedora", d),
	}
}

type FedoraProvisioner struct {
	*RedHatProvisioner
}

func (provisioner *FedoraProvisioner) String() string {
	return "fedora"
}
