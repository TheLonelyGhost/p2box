package commands

import (
	"testing"

	"github.com/thelonelyghost/p2box/commands/commandstest"
	"github.com/thelonelyghost/p2box/drivers/fakedriver"
	"github.com/thelonelyghost/p2box/libmachine"
	"github.com/thelonelyghost/p2box/libmachine/host"
	"github.com/thelonelyghost/p2box/libmachine/libmachinetest"
	"github.com/thelonelyghost/p2box/libmachine/state"
	"github.com/stretchr/testify/assert"
)

func TestCmdStop(t *testing.T) {
	testCases := []struct {
		commandLine    CommandLine
		api            libmachine.API
		expectedErr    error
		expectedStates map[string]state.State
	}{
		{
			commandLine: &commandstest.FakeCommandLine{
				CliArgs: []string{},
			},
			api: &libmachinetest.FakeAPI{
				Hosts: []*host.Host{
					{
						Name: defaultMachineName,
						Driver: &fakedriver.Driver{
							MockState: state.Running,
						},
					},
				},
			},
			expectedErr: nil,
			expectedStates: map[string]state.State{
				defaultMachineName: state.Stopped,
			},
		},
		{
			commandLine: &commandstest.FakeCommandLine{
				CliArgs: []string{},
			},
			api: &libmachinetest.FakeAPI{
				Hosts: []*host.Host{
					{
						Name: "foobar",
						Driver: &fakedriver.Driver{
							MockState: state.Running,
						},
					},
				},
			},
			expectedErr: ErrNoDefault,
			expectedStates: map[string]state.State{
				"foobar": state.Running,
			},
		},
		{
			commandLine: &commandstest.FakeCommandLine{
				CliArgs: []string{"machineToStop1", "machineToStop2"},
			},
			api: &libmachinetest.FakeAPI{
				Hosts: []*host.Host{
					{
						Name: "machineToStop1",
						Driver: &fakedriver.Driver{
							MockState: state.Running,
						},
					},
					{
						Name: "machineToStop2",
						Driver: &fakedriver.Driver{
							MockState: state.Running,
						},
					},
					{
						Name: "machine",
						Driver: &fakedriver.Driver{
							MockState: state.Running,
						},
					},
				},
			},
			expectedErr: nil,
			expectedStates: map[string]state.State{
				"machineToStop1": state.Stopped,
				"machineToStop2": state.Stopped,
				"machine":        state.Running,
			},
		},
	}

	for _, tc := range testCases {
		err := cmdStop(tc.commandLine, tc.api)
		assert.Equal(t, tc.expectedErr, err)

		for hostName, expectedState := range tc.expectedStates {
			assert.Equal(t, expectedState, libmachinetest.State(tc.api, hostName))
		}
	}
}
