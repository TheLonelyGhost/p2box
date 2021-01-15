package provision

import (
	"github.com/thelonelyghost/p2box/libmachine/auth"
	"github.com/thelonelyghost/p2box/libmachine/drivers"
	"github.com/thelonelyghost/p2box/libmachine/engine"
	"github.com/thelonelyghost/p2box/libmachine/provision/pkgaction"
	"github.com/thelonelyghost/p2box/libmachine/provision/serviceaction"
)

type FakeDetector struct {
	Provisioner
}

func (fd *FakeDetector) DetectProvisioner(d drivers.Driver) (Provisioner, error) {
	return fd.Provisioner, nil
}

type FakeProvisioner struct{}

func NewFakeProvisioner(d drivers.Driver) Provisioner {
	return &FakeProvisioner{}
}

func (fp *FakeProvisioner) SSHCommand(args string) (string, error) {
	return "", nil
}

func (fp *FakeProvisioner) String() string {
	return "fakeprovisioner"
}

func (fp *FakeProvisioner) GenerateEngineOptions() (*EngineOptions, error) {
	return nil, nil
}

func (fp *FakeProvisioner) GetEngineOptionsDir() string {
	return ""
}

func (fp *FakeProvisioner) GetAuthOptions() auth.Options {
	return auth.Options{}
}

func (fp *FakeProvisioner) Package(name string, action pkgaction.PackageAction) error {
	return nil
}

func (fp *FakeProvisioner) Hostname() (string, error) {
	return "", nil
}

func (fp *FakeProvisioner) SetHostname(hostname string) error {
	return nil
}

func (fp *FakeProvisioner) CompatibleWithHost() bool {
	return true
}

func (fp *FakeProvisioner) Provision(authOptions auth.Options, engineOptions engine.Options) error {
	return nil
}

func (fp *FakeProvisioner) Service(name string, action serviceaction.ServiceAction) error {
	return nil
}

func (fp *FakeProvisioner) GetDriver() drivers.Driver {
	return nil
}

func (fp *FakeProvisioner) SetOsReleaseInfo(info *OsRelease) {}

func (fp *FakeProvisioner) GetOsReleaseInfo() (*OsRelease, error) {
	return nil, nil
}

type NetstatProvisioner struct {
	*FakeProvisioner
}

func (p *NetstatProvisioner) SSHCommand(args string) (string, error) {
	return `Active Internet connections (servers and established)
Proto Recv-Q Send-Q Local Address           Foreign Address         State
tcp        0      0 0.0.0.0:ssh             0.0.0.0:*               LISTEN
tcp        0     72 192.168.25.141:ssh      192.168.25.1:63235      ESTABLISHED
tcp        0      0 :::ssh                  :::*                    LISTEN
Active UNIX domain sockets (servers and established)
Proto RefCnt Flags       Type       State         I-Node Path
unix  2      [ ACC ]     STREAM     LISTENING      17990 /var/run/acpid.socket
unix  2      [ ACC ]     SEQPACKET  LISTENING      14233 /run/udev/control
unix  2      [ ACC ]     STREAM     LISTENING      19365 /var/run/podman.sock
unix  3      [ ]         STREAM     CONNECTED      19774
unix  3      [ ]         STREAM     CONNECTED      19775
unix  3      [ ]         DGRAM                     14243
unix  3      [ ]         DGRAM                     14242`, nil
}

func NewNetstatProvisioner() Provisioner {
	return &NetstatProvisioner{
		&FakeProvisioner{},
	}
}
