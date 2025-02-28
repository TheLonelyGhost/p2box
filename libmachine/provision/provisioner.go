package provision

import (
	"fmt"

	"github.com/thelonelyghost/p2box/libmachine/auth"
	"github.com/thelonelyghost/p2box/libmachine/drivers"
	"github.com/thelonelyghost/p2box/libmachine/engine"
	"github.com/thelonelyghost/p2box/libmachine/log"
	"github.com/thelonelyghost/p2box/libmachine/provision/pkgaction"
	"github.com/thelonelyghost/p2box/libmachine/provision/serviceaction"
)

var (
	provisioners          = make(map[string]*RegisteredProvisioner)
	detector     Detector = &StandardDetector{}
)

type SSHCommander interface {
	// Short-hand for accessing an SSH command from the driver.
	SSHCommand(args string) (string, error)
}

type Detector interface {
	DetectProvisioner(d drivers.Driver) (Provisioner, error)
}

type StandardDetector struct{}

func SetDetector(newDetector Detector) {
	detector = newDetector
}

// Provisioner defines distribution specific actions
type Provisioner interface {
	fmt.Stringer
	SSHCommander

	// Create the files for the daemon to consume configuration settings (return struct of content and path)
	GenerateEngineOptions() (*EngineOptions, error)

	// Get the directory where the settings files for engine are to be found
	GetEngineOptionsDir() string

	// Return the auth options used to configure remote connection for the daemon.
	GetAuthOptions() auth.Options

	// Run a package action e.g. install
	Package(name string, action pkgaction.PackageAction) error

	// Get Hostname
	Hostname() (string, error)

	// Set hostname
	SetHostname(hostname string) error

	// Figure out if this is the right provisioner to use based on /etc/os-release info
	CompatibleWithHost() bool

	// Do the actual provisioning piece:
	//     1. Set the hostname on the instance.
	//     2. Install engine if it is not present.
	//     3. Configure the daemon to accept connections over TLS.
	//     4. Copy the needed certificates to the server and local config dir.
	Provision(authOptions auth.Options, engineOptions engine.Options) error

	// Perform action on a named service e.g. stop
	Service(name string, action serviceaction.ServiceAction) error

	// Get the driver which is contained in the provisioner.
	GetDriver() drivers.Driver

	// Set the OS Release info depending on how it's represented
	// internally
	SetOsReleaseInfo(info *OsRelease)

	// Get the OS Release info for the current provisioner
	GetOsReleaseInfo() (*OsRelease, error)
}

// RegisteredProvisioner creates a new provisioner
type RegisteredProvisioner struct {
	New func(d drivers.Driver) Provisioner
}

func Register(name string, p *RegisteredProvisioner) {
	provisioners[name] = p
}

func DetectProvisioner(d drivers.Driver) (Provisioner, error) {
	return detector.DetectProvisioner(d)
}

func (detector StandardDetector) DetectProvisioner(d drivers.Driver) (Provisioner, error) {
	log.Info("Waiting for SSH to be available...")
	if err := drivers.WaitForSSH(d); err != nil {
		return nil, err
	}

	log.Info("Detecting the provisioner...")

	osReleaseOut, err := drivers.RunSSHCommandFromDriver(d, "cat /etc/os-release")
	if err != nil {
		return nil, fmt.Errorf("Error getting SSH command: %s", err)
	}

	osReleaseInfo, err := NewOsRelease([]byte(osReleaseOut))
	if err != nil {
		return nil, fmt.Errorf("Error parsing /etc/os-release file: %s", err)
	}

	for _, p := range provisioners {
		provisioner := p.New(d)
		provisioner.SetOsReleaseInfo(osReleaseInfo)

		if provisioner.CompatibleWithHost() {
			log.Debugf("found compatible host: %s", osReleaseInfo.ID)
			return provisioner, nil
		}
	}

	return nil, ErrDetectionFailed
}
