package main

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	flag "github.com/spf13/pflag"

	"github.com/mistermx/forge/pkg/forgefile/execute"
	"github.com/mistermx/forge/pkg/forgefile/render"
	"github.com/mistermx/forge/pkg/log"
	"github.com/mistermx/forge/pkg/utils/pointer"
)

const (
	errNoTargets = "no targets given"
)

var (
	forgeFilePath = flag.StringP("forgefile", "f", "./forgefile", "Path to the forgefile to execute")

	isDryRunEnabled = flag.BoolP("dryrun", "n", false, "Print selected targets but don't execute them")
	isDebugEnabled  = flag.Bool("debug", false, "Enable debug logging")
)

func init() {
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Forge - generic build tool powered by YAML and Go templates.")
		fmt.Fprintln(os.Stderr, "Usage:\n\tforge [options] <target> [<target>, ...]\nAvailable Options:")
		flag.PrintDefaults()
	}
}

func main() {
	flag.Parse()

	logger := createLogger()
	if err := forge(logger); err != nil {
		println(err.Error())
		os.Exit(1)
	}
}

func forge(logger log.Logger) error {
	targets := flag.Args()
	if len(targets) == 0 {
		flag.Usage()

		return errors.New(errNoTargets)
	}

	engineOpts := []render.EngineOption{
		render.WithLogger(logger),
	}

	engine := render.NewEngine(engineOpts...)
	forgeFile, err := engine.Render(*forgeFilePath)
	if err != nil {
		return err
	}

	runnerOpts := []execute.RunnerOption{
		execute.WithLogger(logger),
	}
	if pointer.Deref(isDryRunEnabled, false) {
		runnerOpts = append(runnerOpts, execute.WithExecuter(execute.NewDryRunexecutor(logger)))
	}
	runner := execute.NewRunner(runnerOpts...)
	return runner.Run(forgeFile, targets)
}

func createLogger() log.Logger {
	if pointer.Deref(isDebugEnabled, false) {
		return log.NewDebugLogger()
	}
	return log.NewInfoLogger()
}
