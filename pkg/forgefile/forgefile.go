package forgefile

import (
	"fmt"
	"strings"
)

const (
	// String that can be put before a command to signal forge to
	// ignore any non-zero error codes.
	CommandIgnoreErrorPrefix = "?"
)

// A ForgeFileCommand specifies a shell script to be run during the executing
// of a ForgeFileTarget.
type ForgeFileCommand string

// ShouldIgnoreError returns true if the command string is prefixed with `?`
// that tells the runner to ignore any non-zero exit codes returned by the
// command.
//
// Example:
//
//	`- ?echo "This error will be ignored" && exit 1`
func (c ForgeFileCommand) ShouldIgnoreError() bool {
	return strings.HasPrefix(string(c), CommandIgnoreErrorPrefix)
}

// Gets the command string to be executed trimmed by any prefixes.
func (c ForgeFileCommand) GetCommand() string {
	if strings.HasPrefix(string(c), CommandIgnoreErrorPrefix) {
		return string(c)[1:]
	}
	return string(c)
}

// ForgeFileCommandEnvironment specifies environment settings for a
// ForgeFileCommand
type ForgeFileCommandEnvironment map[string]string

// AsSlice converts e into a slice of `KEY=VALUE` strings.
func (e ForgeFileCommandEnvironment) AsSlice() []string {
	slice := make([]string, 0, len(e))
	for k, v := range e {
		slice = append(slice, fmt.Sprintf("%s=%s", k, v))
	}
	return slice
}

// ForgeFileTargetType represents the type of ForgeFileTarget.
type ForgeFileTargetType string

const (
	// The target's name matches a file and is only executed if this does not
	// exist.
	ForgeFileTargetTypeFile ForgeFileTargetType = "file"
	// The target's name matches a directory and is only executed if this does
	// not exist.
	ForgeFileTargetTypeDirectory ForgeFileTargetType = "directory"
	// The target does not match anything on the local file system and is
	// executed always. This is the default behaviour.
	ForgeFileTargetTypeVirtual ForgeFileTargetType = "virtual"
)

// ForgeFileTarget represents a target definition in a ForgeFile.
// Every target defines a list of commands to be executed, a number of dependent
// targets that should be executed before and optional options.
//
// Every target represents a path to a file relative the PWD of the forge
// process. If the file exists, the target and it's dependencies will not be
// executed. If a target is marked as virtual, it is executed regardless if a
// corresponding file exists or not.
type ForgeFileTarget struct {
	// Defines when the target is executed. This can be either `file`,
	// `directory` or `virtual`. The default is `virtual.`
	Type *ForgeFileTargetType `yaml:"type,omitempty"`
	// DependsOn is a list of dependent targets that should be executed before
	// this target.
	DependsOn []string `yaml:"dependsOn,omitempty"`
	// Commands that should be executed for this target.
	// All commands of a single target are executed in a seperated sub-shell
	// by default.
	//
	// If one command exits with a non-zero error code, the consecutive commands
	// are skipped and forge aborts with an error.
	//
	// If a command is prefix with `?` any non-zero error code is ignored and
	// forge continues as if the command ran successfully.
	Commands []ForgeFileCommand `yaml:"commands,omitempty"`
	// Environment variables that should be passed to the shell prior to the
	// execution of the first command.
	Environment ForgeFileCommandEnvironment `yaml:"environment,omitempty"`
}

// A ForgeFile is a YAML file that defines a set of targets that can be executed
// from the command line using `forge [target1] [target2]`.
//
// Targets are executed in the order they are specified including the
// dependencies they belong to.
type ForgeFile map[string]ForgeFileTarget
