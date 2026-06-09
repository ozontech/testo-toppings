package environment

import "github.com/ozontech/testo/testoplugin"

type option func(p *PluginEnvironment)

// WithEnvironments specifies which .env files to load.
// Later files overwrite earlier files for the same environment variable.
// If no files are specified, defaults to loading .env from the current directory.
func WithEnvironments(filenames ...string) testoplugin.Option {
	return testoplugin.Option{
		Value: option(func(p *PluginEnvironment) {
			p.filenames = filenames
		}),
		Propagate: true,
	}
}
