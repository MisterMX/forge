package render

import (
	"bytes"
	"path/filepath"
	"text/template"

	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"

	"github.com/mistermx/forge/pkg/forgefile"
	"github.com/mistermx/forge/pkg/log"
)

// An Engine creates a ForgeFile from a Go template file that renders to a YAML
// file.
type Engine struct {
	fs     afero.Fs
	logger log.Logger
}

// An EngineOption modifies an Engine upon creation.
type EngineOption func(e *Engine)

// NewEngine creates a new Engine.
func NewEngine(opts ...EngineOption) *Engine {
	engine := &Engine{
		fs:     afero.NewOsFs(),
		logger: log.NewNoopLogger(),
	}
	for _, o := range opts {
		o(engine)
	}
	return engine
}

// WithLogger modifies an Engine to use the given Logger.
func WithLogger(logger log.Logger) EngineOption {
	return func(e *Engine) {
		e.logger = logger
	}
}

// Render the ForgeFile at the given path.
func (e *Engine) Render(path string) (forgefile.ForgeFile, error) {

	content, err := afero.ReadFile(e.fs, path)
	if err != nil {
		return nil, err
	}

	tp, err := template.New("forgefile").Parse(string(content))
	if err != nil {
		return nil, err
	}

	data := map[string]any{
		"Forge": map[string]any{
			"ForgeFile":    path,
			"ForgeFileDir": filepath.Dir(path),
		},
	}

	renderResult := bytes.Buffer{}
	if err := tp.Execute(&renderResult, data); err != nil {
		return nil, err
	}

	result := forgefile.ForgeFile{}
	if err := yaml.Unmarshal(renderResult.Bytes(), result); err != nil {
		return nil, err
	}
	return result, nil
}
