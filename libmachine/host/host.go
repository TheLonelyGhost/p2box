package host

import (
	"os/exec"
	"regexp"

	"github.com/thelonelyghost/p2box/libmachine/auth"
	"github.com/thelonelyghost/p2box/libmachine/cert"
	"github.com/thelonelyghost/p2box/libmachine/drivers"
	"github.com/thelonelyghost/p2box/libmachine/engine"
	"github.com/thelonelyghost/p2box/libmachine/log"
	"github.com/thelonelyghost/p2box/libmachine/mcnerror"
	"github.com/thelonelyghost/p2box/libmachine/mcnutils"
	"github.com/thelonelyghost/p2box/libmachine/provision"
	"github.com/thelonelyghost/p2box/libmachine/provision/pkgaction"
	"github.com/thelonelyghost/p2box/libmachine/ssh"
	"github.com/thelonelyghost/p2box/libmachine/state"
)

var (
	validHostNamePattern                  = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9\-\.]*$`)
	stdSSHClientCreator  SSHClientCreator = &StandardSSHClientCreator{}
)

type SSHClientCreator interface {
	CreateSSHClient(d drivers.Driver) (ssh.Client, error)
	CreateExternalSSHClient(d drivers.Driver) (*ssh.ExternalClient, error)
	CreateExternalRootSSHClient(d drivers.Driver) (*ssh.ExternalClient, error)
}

type StandardSSHClientCreator struct {
	drivers.Driver
}

func SetSSHClientCreator(creator SSHClientCreator) {
	stdSSHClientCreator = creator
}

type Host struct {
	ConfigVersion int
	Driver        drivers.Driver
	DriverName    string
	HostOptions   *Options
	Name          string
	RawDriver     []byte `json:"-"`
}

type Options struct {
	Driver        string
	Memory        int
	Disk          int
	EngineOptions *engine.Options
	AuthOptions   *auth.Options
}

type Metadata struct {
	ConfigVersion int
	DriverName    string
	HostOptions   Options
}

func ValidateHostName(name string) bool {
	return validHostNamePattern.MatchString(name)
}

func (h *Host) RunSSHCommand(command string) (string, error) {
	return drivers.RunSSHCommandFromDriver(h.Driver, command)
}

func (h *Host) CreateSSHClient() (ssh.Client, error) {
	return stdSSHClientCreator.CreateSSHClient(h.Driver)
}

func (creator *StandardSSHClientCreator) CreateSSHClient(d drivers.Driver) (ssh.Client, error) {
	addr, err := d.GetSSHHostname()
	if err != nil {
		return &ssh.ExternalClient{}, err
	}

	port, err := d.GetSSHPort()
	if err != nil {
		return &ssh.ExternalClient{}, err
	}

	auth := &ssh.Auth{}
	if d.GetSSHKeyPath() != "" {
		auth.Keys = []string{d.GetSSHKeyPath()}
	}

	return ssh.NewClient(d.GetSSHUsername(), addr, port, auth)
}

func (h *Host) CreateExternalSSHClient() (*ssh.ExternalClient, error) {
	return stdSSHClientCreator.CreateExternalSSHClient(h.Driver)
}

func (creator *StandardSSHClientCreator) CreateExternalSSHClient(d drivers.Driver) (*ssh.ExternalClient, error) {
	sshBinaryPath, err := exec.LookPath("ssh")
	if err != nil {
		return &ssh.ExternalClient{}, err
	}

	addr, err := d.GetSSHHostname()
	if err != nil {
		return &ssh.ExternalClient{}, err
	}

	port, err := d.GetSSHPort()
	if err != nil {
		return &ssh.ExternalClient{}, err
	}

	auth := &ssh.Auth{}
	if d.GetSSHKeyPath() != "" {
		auth.Keys = []string{d.GetSSHKeyPath()}
	}

	return ssh.NewExternalClient(sshBinaryPath, d.GetSSHUsername(), addr, port, auth)
}

func (h *Host) CreateExternalRootSSHClient() (*ssh.ExternalClient, error) {
	return stdSSHClientCreator.CreateExternalRootSSHClient(h.Driver)
}

func (creator *StandardSSHClientCreator) CreateExternalRootSSHClient(d drivers.Driver) (*ssh.ExternalClient, error) {
	sshBinaryPath, err := exec.LookPath("ssh")
	if err != nil {
		return &ssh.ExternalClient{}, err
	}

	addr, err := d.GetSSHHostname()
	if err != nil {
		return &ssh.ExternalClient{}, err
	}

	port, err := d.GetSSHPort()
	if err != nil {
		return &ssh.ExternalClient{}, err
	}

	auth := &ssh.Auth{}
	if d.GetSSHKeyPath() != "" {
		auth.Keys = []string{d.GetSSHKeyPath()}
	}

	return ssh.NewExternalClient(sshBinaryPath, "root", addr, port, auth)
}

func (h *Host) runActionForState(action func() error, desiredState state.State) error {
	if drivers.MachineInState(h.Driver, desiredState)() {
		return mcnerror.ErrHostAlreadyInState{
			Name:  h.Name,
			State: desiredState,
		}
	}

	if err := action(); err != nil {
		return err
	}

	return mcnutils.WaitFor(drivers.MachineInState(h.Driver, desiredState))
}

func (h *Host) WaitForPodman() error {
	provisioner, err := provision.DetectProvisioner(h.Driver)
	if err != nil {
		return err
	}

	return provision.WaitForPodman(provisioner)
}

func (h *Host) Start() error {
	log.Infof("Starting %q...", h.Name)
	if err := h.runActionForState(h.Driver.Start, state.Running); err != nil {
		return err
	}

	log.Infof("Machine %q was started.", h.Name)

	return h.WaitForPodman()
}

func (h *Host) Stop() error {
	log.Infof("Stopping %q...", h.Name)
	if err := h.runActionForState(h.Driver.Stop, state.Stopped); err != nil {
		return err
	}

	log.Infof("Machine %q was stopped.", h.Name)
	return nil
}

func (h *Host) Kill() error {
	log.Infof("Killing %q...", h.Name)
	if err := h.runActionForState(h.Driver.Kill, state.Stopped); err != nil {
		return err
	}

	log.Infof("Machine %q was killed.", h.Name)
	return nil
}

func (h *Host) Restart() error {
	log.Infof("Restarting %q...", h.Name)
	if drivers.MachineInState(h.Driver, state.Stopped)() {
		if err := h.Start(); err != nil {
			return err
		}
	} else if drivers.MachineInState(h.Driver, state.Running)() {
		if err := h.Driver.Restart(); err != nil {
			return err
		}
		if err := mcnutils.WaitFor(drivers.MachineInState(h.Driver, state.Running)); err != nil {
			return err
		}
	}

	return h.WaitForPodman()
}

func (h *Host) Upgrade() error {
	machineState, err := h.Driver.GetState()
	if err != nil {
		return err
	}

	if machineState != state.Running {
		log.Info("Starting machine so machine can be upgraded...")
		if err := h.Start(); err != nil {
			return err
		}
	}

	provisioner, err := provision.DetectProvisioner(h.Driver)
	if err != nil {
		return err
	}

	log.Info("Upgrading podman...")
	if err := provisioner.Package("podman", pkgaction.Upgrade); err != nil {
		return err
	}

	return err
}

func (h *Host) URL() (string, error) {
	return h.Driver.GetURL()
}

func (h *Host) AuthOptions() *auth.Options {
	if h.HostOptions == nil {
		return nil
	}
	return h.HostOptions.AuthOptions
}

func (h *Host) ConfigureAuth() error {
	provisioner, err := provision.DetectProvisioner(h.Driver)
	if err != nil {
		return err
	}

	// TODO: This is kind of a hack (or is it?  I'm not really sure until
	// we have more clearly defined outlook on what the responsibilities
	// and modularity of the provisioners should be).
	//
	// Call provision to re-provision the certs properly.
	return provisioner.Provision(*h.HostOptions.AuthOptions, *h.HostOptions.EngineOptions)
}

func (h *Host) ConfigureAllAuth() error {
	log.Info("Regenerating local certificates")
	if err := cert.BootstrapCertificates(h.AuthOptions()); err != nil {
		return err
	}
	return h.ConfigureAuth()
}

func (h *Host) Provision() error {
	provisioner, err := provision.DetectProvisioner(h.Driver)
	if err != nil {
		return err
	}

	return provisioner.Provision(*h.HostOptions.AuthOptions, *h.HostOptions.EngineOptions)
}
