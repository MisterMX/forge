package execute

import (
	"io/fs"

	"github.com/pkg/errors"
	"github.com/spf13/afero"

	"github.com/mistermx/forge/pkg/forgefile"
	"github.com/mistermx/forge/pkg/log"
	"github.com/mistermx/forge/pkg/utils/pointer"
)

const (
	errTargetNotFound    = "target '%s' not found"
	errResolveTarget     = "failed to resolve target '%s'"
	errResolveDependency = "failed to resolve dependency '%s'"
	errDependencySelf    = "targets cannot depend on themselves"
	errRunTarget         = "failed to run target '%s'"
	errInvalidTargetType = "unkown type '%s' for target '%s'"
)

// A target is a resolved ForgeFile target in a TargetChain.
type Target struct {
	forgefile.ForgeFileTarget
	name string
}

// A TargetChain represents an execution
type TargetChain []Target

func (c TargetChain) ContainsTarget(name string) bool {
	for _, t := range c {
		if t.name == name {
			return true
		}
	}
	return false
}

// CommandExecuter runs commands in a ForgeFile.
type CommandExecuter interface {
	Execute(t Target) error
}

// A Runner executes targets of a ForgeFile.
type Runner struct {
	commandExecuter CommandExecuter
	fs              afero.Fs
	log             log.Logger
}

// A RunnerOption modifies a runner upon creation.
type RunnerOption func(r *Runner)

// NewRunner creates a new Runner.
func NewRunner(opts ...RunnerOption) *Runner {
	r := &Runner{
		commandExecuter: NewShellExecutor(),
		fs:              afero.NewOsFs(),
		log:             log.NewNoopLogger(),
	}
	for _, o := range opts {
		o(r)
	}
	return nil
}

// WithLogger modifies a Runner to use the given Logger.
func WithLogger(logger log.Logger) RunnerOption {
	return func(r *Runner) {
		r.log = logger
	}
}

// WithExecuter modifies a Runner to execute commands using the given
// CommandExecuter.
func WithExecuter(executer CommandExecuter) RunnerOption {
	return func(r *Runner) {
		r.commandExecuter = executer
	}
}

// Run the targets of a ForgeFile.
func (r *Runner) Run(ff forgefile.ForgeFile, targets []string) error {
	chain, err := r.buildTargetChain(targets, ff)
	if err != nil {
		return err
	}
	return r.runTargets(chain)
}

func (r *Runner) runTargets(chain TargetChain) error {
	r.log.Debugf("Executing target chain %v", chain)
	for _, t := range chain {
		r.log.Debugf("Executing target '%s'", t.name)
		if err := r.runTarget(t); err != nil {
			return errors.Wrapf(err, errRunTarget, t.name)
		}
	}
	return nil
}

func (r *Runner) runTarget(t Target) error {
	targetType := pointer.Deref(t.Type, forgefile.ForgeFileTargetTypeVirtual)
	switch targetType {
	case forgefile.ForgeFileTargetTypeVirtual:
		break
	case forgefile.ForgeFileTargetTypeFile:
		fileExists, err := r.isFile(t.name)
		if err != nil {
			return err
		}
		if fileExists {
			r.log.Debugf("Target file %s already exists. Skipping.", t.name)
			return nil
		}
	case forgefile.ForgeFileTargetTypeDirectory:
		dirExists, err := r.isDirectory(t.name)
		if err != nil {
			return err
		}
		if dirExists {
			r.log.Debugf("Target directory %s already exists. Skipping.", t.name)
			return nil
		}
	default:
		return errors.Errorf(errInvalidTargetType, string(targetType), t.name)
	}

	if err := r.commandExecuter.Execute(t); err != nil {
		return err
	}
	return nil
}

func (r *Runner) buildTargetChain(targets []string, ff forgefile.ForgeFile) (TargetChain, error) {
	resolved := make(TargetChain, 0, len(targets))
	for _, t := range targets {
		var err error
		resolved, err = r.resolveTarget(t, resolved, ff)
		if err != nil {
			return nil, errors.Wrapf(err, errResolveTarget, t)
		}
	}
	return resolved, nil
}

func (r *Runner) resolveTarget(name string, resolved TargetChain, ff forgefile.ForgeFile) (TargetChain, error) {
	ft, exists := ff[name]
	if !exists {
		return nil, errors.Errorf(errTargetNotFound, name)
	}

	for _, dep := range ft.DependsOn {
		if dep == name {
			return nil, errors.New(errDependencySelf)
		}
		if resolved.ContainsTarget(dep) {
			continue
		}

		var err error
		resolved, err = r.resolveTarget(dep, resolved, ff)
		if err != nil {
			return nil, errors.Wrapf(err, errResolveDependency, dep)
		}
	}
	return append(resolved, Target{ft, name}), nil
}

func (r *Runner) isDirectory(path string) (bool, error) {
	info, err := r.fs.Stat(path)
	if err != nil {
		return false, ignoreNotFound(err)
	}
	return info.IsDir(), nil
}

func (r *Runner) isFile(path string) (bool, error) {
	info, err := r.fs.Stat(path)
	if err != nil {
		return false, ignoreNotFound(err)
	}
	return !info.IsDir(), nil
}

func ignoreNotFound(err error) error {
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	return err
}
