package provision

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/thelonelyghost/p2box/libmachine/drivers"
	"github.com/thelonelyghost/p2box/libmachine/provision/serviceaction"
)

type SystemdProvisioner struct {
	GenericProvisioner
}

func (p *SystemdProvisioner) String() string {
	return "redhat"
}

func NewSystemdProvisioner(osReleaseID string, d drivers.Driver) SystemdProvisioner {
	return SystemdProvisioner{
		GenericProvisioner{
			SSHCommander: GenericSSHCommander{Driver: d},

			OsReleaseID: osReleaseID,
			Packages:    []string{},
			Driver:      d,
		},
	}
}

func (p *SystemdProvisioner) GenerateEngineOptions() (*EngineOptions, error) {
	var (
		engineCfg bytes.Buffer
	)

	driverNameLabel := fmt.Sprintf("provider=%s", p.Driver.DriverName())
	p.EngineOptions.Labels = append(p.EngineOptions.Labels, driverNameLabel)

	engineConfigTmpl := `--storage-driver {{.EngineOptions.StorageDriver}} --tlsverify --tlscacert {{.AuthOptions.CaCertRemotePath}} --tlscert {{.AuthOptions.ServerCertRemotePath}} --tlskey {{.AuthOptions.ServerKeyRemotePath}} {{ range .EngineOptions.Labels }}--label {{.}} {{ end }}{{ range .EngineOptions.InsecureRegistry }}--insecure-registry {{.}} {{ end }}{{ range .EngineOptions.RegistryMirror }}--registry-mirror {{.}} {{ end }}{{ range .EngineOptions.ArbitraryFlags }}--{{.}} {{ end }}
Environment={{range .EngineOptions.Env}}{{ printf "%q" . }} {{end}}
`
	t, err := template.New("engineConfig").Parse(engineConfigTmpl)
	if err != nil {
		return nil, err
	}

	engineConfigContext := EngineConfigContext{
		AuthOptions:   p.AuthOptions,
		EngineOptions: p.EngineOptions,
	}

	t.Execute(&engineCfg, engineConfigContext)

	return &EngineOptions{
		EngineOptionsString: engineCfg.String(),
		EngineOptionsPath:   p.DaemonOptionsFile,
	}, nil
}

func (p *SystemdProvisioner) Service(name string, action serviceaction.ServiceAction) error {
	reloadDaemon := false
	switch action {
	case serviceaction.Start, serviceaction.Restart:
		reloadDaemon = true
	}

	// systemd needs reloaded when config changes on disk; we cannot
	// be sure exactly when it changes from the provisioner so
	// we call a reload on every restart to be safe
	if reloadDaemon {
		if _, err := p.SSHCommand("sudo systemctl daemon-reload"); err != nil {
			return err
		}
	}

	command := fmt.Sprintf("sudo systemctl -f %s %s", action.String(), name)

	if _, err := p.SSHCommand(command); err != nil {
		return err
	}

	return nil
}
