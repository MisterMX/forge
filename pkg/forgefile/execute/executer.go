package execute

import (
	"os"
	"os/exec"

	"github.com/pkg/errors"

	"github.com/mistermx/forge/pkg/log"
)

const (
	// The default shell that is used by the ShellExecutor.
	DefaultShell = "/bin/sh"

	errRunCommand = "failed to run command at index %d"
)

var _ CommandExecuter = &ShellExecutor{}

// A ShellExecutor executes commands of ForgeFile target in the system shell.
type ShellExecutor struct{}

// NewShellExecutor creates a new ShellExecutor.
func NewShellExecutor() *ShellExecutor {
	return &ShellExecutor{}
}

// Execute the commands of the target in a shell.
func (e *ShellExecutor) Execute(t Target) error {
	env := append(os.Environ(), t.Environment.AsSlice()...)

	for i, cmd := range t.Commands {
		shell := exec.Command(DefaultShell, "-c", cmd.GetCommand())
		shell.Env = env
		shell.Stdout = os.Stdout
		shell.Stderr = os.Stderr
		if err := shell.Run(); err != nil && (!isExitError(err) || cmd.ShouldIgnoreError()) {
			return errors.Wrapf(err, errRunCommand, i)
		}
	}
	return nil
}

func isExitError(err error) bool {
	_, ok := err.(*exec.ExitError)
	return ok
}

// A DryRunexecutor does not print the commands but logs them instead.
type DryRunexecutor struct {
	log log.Logger
}

// NewDryRunexecutor creates a new DryRunexecutor with the given logger.
func NewDryRunexecutor(logger log.Logger) *DryRunexecutor {
	return &DryRunexecutor{logger}
}

// Execute logs every command instead of executing it.
func (e *DryRunexecutor) Execute(t Target) error {
	for _, cmd := range t.Commands {
		e.log.Info(cmd.GetCommand())
	}
	return nil
}
