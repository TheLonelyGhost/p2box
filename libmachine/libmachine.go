package libmachine

import (
	"fmt"
	"path/filepath"

	"io"

	"github.com/thelonelyghost/p2box/drivers/errdriver"
	"github.com/thelonelyghost/p2box/libmachine/auth"
	"github.com/thelonelyghost/p2box/libmachine/cert"
	"github.com/thelonelyghost/p2box/libmachine/drivers"
	"github.com/thelonelyghost/p2box/libmachine/drivers/plugin/localbinary"
	"github.com/thelonelyghost/p2box/libmachine/drivers/rpc"
	"github.com/thelonelyghost/p2box/libmachine/engine"
	"github.com/thelonelyghost/p2box/libmachine/host"
	"github.com/thelonelyghost/p2box/libmachine/log"
	"github.com/thelonelyghost/p2box/libmachine/mcnerror"
	"github.com/thelonelyghost/p2box/libmachine/mcnutils"
	"github.com/thelonelyghost/p2box/libmachine/persist"
	"github.com/thelonelyghost/p2box/libmachine/provision"
	"github.com/thelonelyghost/p2box/libmachine/ssh"
	"github.com/thelonelyghost/p2box/libmachine/state"
	"github.com/thelonelyghost/p2box/libmachine/version"
)

type API interface {
	io.Closer
	NewHost(driverName string, rawDriver []byte) (*host.Host, error)
	Create(h *host.Host) error
	persist.Store
	GetMachinesDir() string
}

type Client struct {
	certsDir       string
	IsDebug        bool
	SSHClientType  ssh.ClientType
	GithubAPIToken string
	*persist.Filestore
	clientDriverFactory rpcdriver.RPCClientDriverFactory
}

func NewClient(storePath, certsDir string) *Client {
	return &Client{
		certsDir:            certsDir,
		IsDebug:             false,
		SSHClientType:       ssh.External,
		Filestore:           persist.NewFilestore(storePath, certsDir, certsDir),
		clientDriverFactory: rpcdriver.NewRPCClientDriverFactory(),
	}
}

func (api *Client) NewHost(driverName string, rawDriver []byte) (*host.Host, error) {
	driver, err := api.clientDriverFactory.NewRPCClientDriver(driverName, rawDriver)
	if err != nil {
		return nil, err
	}

	return &host.Host{
		ConfigVersion: version.ConfigVersion,
		Name:          driver.GetMachineName(),
		Driver:        driver,
		DriverName:    driver.DriverName(),
		HostOptions: &host.Options{
			AuthOptions: &auth.Options{
				CertDir:          api.certsDir,
				CaCertPath:       filepath.Join(api.certsDir, "ca.pem"),
				CaPrivateKeyPath: filepath.Join(api.certsDir, "ca-key.pem"),
				ClientCertPath:   filepath.Join(api.certsDir, "cert.pem"),
				ClientKeyPath:    filepath.Join(api.certsDir, "key.pem"),
				ServerCertPath:   filepath.Join(api.GetMachinesDir(), "server.pem"),
				ServerKeyPath:    filepath.Join(api.GetMachinesDir(), "server-key.pem"),
			},
			EngineOptions: &engine.Options{
				InstallURL:    drivers.DefaultEngineInstallURL,
				StorageDriver: "overlay2",
				TLSVerify:     true,
			},
		},
	}, nil
}

func (api *Client) Load(name string) (*host.Host, error) {
	h, err := api.Filestore.Load(name)
	if err != nil {
		return nil, err
	}

	d, err := api.clientDriverFactory.NewRPCClientDriver(h.DriverName, h.RawDriver)
	if err != nil {
		// Not being able to find a driver binary is a "known error"
		if _, ok := err.(localbinary.ErrPluginBinaryNotFound); ok {
			h.Driver = errdriver.NewDriver(h.DriverName)
			return h, nil
		}
		return nil, err
	}

	if h.DriverName == "virtualbox" {
		h.Driver = drivers.NewSerialDriver(d)
	} else {
		h.Driver = d
	}

	return h, nil
}

// Create is the wrapper method which covers all of the boilerplate around
// actually creating, provisioning, and persisting an instance in the store.
func (api *Client) Create(h *host.Host) error {
	if err := cert.BootstrapCertificates(h.AuthOptions()); err != nil {
		return fmt.Errorf("Error generating certificates: %s", err)
	}

	log.Info("Running pre-create checks...")

	if err := h.Driver.PreCreateCheck(); err != nil {
		return mcnerror.ErrDuringPreCreate{
			Cause: err,
		}
	}

	if err := api.Save(h); err != nil {
		return fmt.Errorf("Error saving host to store before attempting creation: %s", err)
	}

	log.Info("Creating machine...")

	if err := api.performCreate(h); err != nil {
		return fmt.Errorf("Error creating machine: %s", err)
	}

	log.Debug("Reticulating splines...")

	return nil
}

func (api *Client) performCreate(h *host.Host) error {
	if err := h.Driver.Create(); err != nil {
		return fmt.Errorf("Error in driver during machine creation: %s", err)
	}

	if err := api.Save(h); err != nil {
		return fmt.Errorf("Error saving host to store after attempting creation: %s", err)
	}

	// TODO: Not really a fan of just checking "none" or "ci-test" here.
	if h.Driver.DriverName() == "none" || h.Driver.DriverName() == "ci-test" {
		return nil
	}

	log.Info("Waiting for machine to be running, this may take a few minutes...")
	if err := mcnutils.WaitFor(drivers.MachineInState(h.Driver, state.Running)); err != nil {
		return fmt.Errorf("Error waiting for machine to be running: %s", err)
	}

	log.Info("Detecting operating system of created instance...")
	provisioner, err := provision.DetectProvisioner(h.Driver)
	if err != nil {
		return fmt.Errorf("Error detecting OS: %s", err)
	}

	log.Infof("Provisioning with %s...", provisioner.String())
	if err := provisioner.Provision(*h.HostOptions.AuthOptions, *h.HostOptions.EngineOptions); err != nil {
		return fmt.Errorf("Error running provisioning: %s", err)
	}

	// We should check the connection to podman here
	log.Info("Checking connection to Podman...")
	client, err := h.CreateSSHClient()
	if err != nil {
		return fmt.Errorf("Error creating SSH client: %s", err)
	}
	version, err := client.Output("podman --version")
	if err != nil {
		return fmt.Errorf("Error getting podman version: %s", err)
	}
	log.Debugf("%s", version)

	log.Info("Podman is up and running!")
	return nil
}

func (api *Client) Close() error {
	return api.clientDriverFactory.Close()
}
