package commands

import (
	"testing"

	"github.com/thelonelyghost/p2box/commands/commandstest"
	"github.com/thelonelyghost/p2box/drivers/fakedriver"
	"github.com/thelonelyghost/p2box/libmachine"
	"github.com/thelonelyghost/p2box/libmachine/auth"
	"github.com/thelonelyghost/p2box/libmachine/engine"
	"github.com/thelonelyghost/p2box/libmachine/host"
	"github.com/thelonelyghost/p2box/libmachine/libmachinetest"
	"github.com/thelonelyghost/p2box/libmachine/provision"
	"github.com/stretchr/testify/assert"
)

func TestCmdProvision(t *testing.T) {
	testCases := []struct {
		commandLine CommandLine
		api         libmachine.API
		expectedErr error
	}{
		{
			commandLine: &commandstest.FakeCommandLine{
				CliArgs: []string{"foo", "bar"},
			},
			api: &libmachinetest.FakeAPI{
				Hosts: []*host.Host{
					{
						Name:   "foo",
						Driver: &fakedriver.Driver{},
						HostOptions: &host.Options{
							EngineOptions: &engine.Options{},
							AuthOptions:   &auth.Options{},
						},
					},
					{
						Name:   "bar",
						Driver: &fakedriver.Driver{},
						HostOptions: &host.Options{
							EngineOptions: &engine.Options{},
							AuthOptions:   &auth.Options{},
						},
					},
				},
			},
			expectedErr: nil,
		},
	}

	provision.SetDetector(&provision.FakeDetector{
		Provisioner: provision.NewFakeProvisioner(nil),
	})

	// fakeprovisioner always returns "true" for compatible host, so we
	// just need to register it.
	provision.Register("fakeprovisioner", &provision.RegisteredProvisioner{
		New: provision.NewFakeProvisioner,
	})

	for _, tc := range testCases {
		assert.Equal(t, tc.expectedErr, cmdProvision(tc.commandLine, tc.api))
	}
}
