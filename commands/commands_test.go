package commands

import (
	"errors"
	"flag"
	"testing"

	"github.com/thelonelyghost/p2box/commands/commandstest"
	"github.com/thelonelyghost/p2box/drivers/fakedriver"
	"github.com/thelonelyghost/p2box/libmachine"
	"github.com/thelonelyghost/p2box/libmachine/host"
	"github.com/thelonelyghost/p2box/libmachine/hosttest"
	"github.com/thelonelyghost/p2box/libmachine/provision"
	"github.com/thelonelyghost/p2box/libmachine/state"
	"github.com/urfave/cli"
	"github.com/stretchr/testify/assert"
)

func TestRunActionForeachMachine(t *testing.T) {
	defer provision.SetDetector(&provision.StandardDetector{})
	provision.SetDetector(&provision.FakeDetector{
		Provisioner: provision.NewNetstatProvisioner(),
	})

	// Assume a bunch of machines in randomly started or
	// stopped states.
	machines := []*host.Host{
		{
			Name:       "foo",
			DriverName: "fakedriver",
			Driver: &fakedriver.Driver{
				MockState: state.Running,
			},
		},
		{
			Name:       "bar",
			DriverName: "fakedriver",
			Driver: &fakedriver.Driver{
				MockState: state.Stopped,
			},
		},
		{
			Name: "baz",
			// Ssh, don't tell anyone but this
			// driver only _thinks_ it's named
			// virtualbox...  (to test serial actions)
			// It's actually FakeDriver!
			DriverName: "virtualbox",
			Driver: &fakedriver.Driver{
				MockState: state.Stopped,
			},
		},
		{
			Name:       "spam",
			DriverName: "virtualbox",
			Driver: &fakedriver.Driver{
				MockState: state.Running,
			},
		},
		{
			Name:       "eggs",
			DriverName: "fakedriver",
			Driver: &fakedriver.Driver{
				MockState: state.Stopped,
			},
		},
		{
			Name:       "ham",
			DriverName: "fakedriver",
			Driver: &fakedriver.Driver{
				MockState: state.Running,
			},
		},
	}

	runActionForeachMachine("start", machines)

	for _, machine := range machines {
		machineState, _ := machine.Driver.GetState()

		assert.Equal(t, state.Running, machineState)
	}

	runActionForeachMachine("stop", machines)

	for _, machine := range machines {
		machineState, _ := machine.Driver.GetState()

		assert.Equal(t, state.Stopped, machineState)
	}
}

func TestPrintIPEmptyGivenLocalEngine(t *testing.T) {
	stdoutGetter := commandstest.NewStdoutGetter()
	defer stdoutGetter.Stop()

	host, _ := hosttest.GetDefaultTestHost()
	err := printIP(host)()

	assert.NoError(t, err)
	assert.Equal(t, "\n", stdoutGetter.Output())
}

func TestPrintIPPrintsGivenRemoteEngine(t *testing.T) {
	stdoutGetter := commandstest.NewStdoutGetter()
	defer stdoutGetter.Stop()

	host, _ := hosttest.GetDefaultTestHost()
	host.Driver = &fakedriver.Driver{
		MockState: state.Running,
		MockIP:    "1.2.3.4",
	}
	err := printIP(host)()

	assert.NoError(t, err)
	assert.Equal(t, "1.2.3.4\n", stdoutGetter.Output())
}

func TestConsolidateError(t *testing.T) {
	cases := []struct {
		inputErrs   []error
		expectedErr error
	}{
		{
			inputErrs: []error{
				errors.New("Couldn't remove host 'bar'"),
			},
			expectedErr: errors.New("Couldn't remove host 'bar'"),
		},
		{
			inputErrs: []error{
				errors.New("Couldn't remove host 'bar'"),
				errors.New("Couldn't remove host 'foo'"),
			},
			expectedErr: errors.New("Couldn't remove host 'bar'\nCouldn't remove host 'foo'"),
		},
	}

	for _, c := range cases {
		assert.Equal(t, c.expectedErr, consolidateErrs(c.inputErrs))
	}
}

type MockCrashReporter struct {
	sent bool
}

func TestReturnExitCode1onError(t *testing.T) {
	command := func(commandLine CommandLine, api libmachine.API) error {
		return errors.New("foo is not bar")
	}

	exitCode := checkErrorCodeForCommand(command)

	assert.Equal(t, 1, exitCode)
}

func checkErrorCodeForCommand(command func(commandLine CommandLine, api libmachine.API) error) int {
	var setExitCode int

	originalOSExit := osExit

	defer func() {
		osExit = originalOSExit
	}()

	osExit = func(code int) {
		setExitCode = code
	}

	context := cli.NewContext(cli.NewApp(), &flag.FlagSet{}, nil)
	runCommand(command)(context)

	return setExitCode
}
