package provision

import (
	"testing"

	"github.com/thelonelyghost/p2box/drivers/fakedriver"
	"github.com/thelonelyghost/p2box/libmachine/auth"
	"github.com/thelonelyghost/p2box/libmachine/engine"
	"github.com/thelonelyghost/p2box/libmachine/provision/provisiontest"
)

func TestRedHatDefaultStorageDriver(t *testing.T) {
	p := NewRedHatProvisioner("", &fakedriver.Driver{})
	p.SSHCommander = provisiontest.NewFakeSSHCommander(provisiontest.FakeSSHCommanderOptions{})
	p.Provision(auth.Options{}, engine.Options{})
}
