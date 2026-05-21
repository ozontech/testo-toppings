package parallel

import (
	"flag"

	"github.com/ozontech/testo/testoplugin"
)

type option func(p *PluginParallel)

var flagSync = flag.Bool("parallel.sync", false, "make all tests synchronous")

type Scope struct {
	Tests      bool
	Suites     bool
	SuiteTests bool
}

func WithScope(scope Scope) testoplugin.Option {
	return testoplugin.Option{
		Value: option(func(p *PluginParallel) {
			p.scope = scope
		}),
		Propagate: true,
	}
}

// WithSync signals that this test is to be run in sync with (and only with) other sync tests.
func WithSync() testoplugin.Option {
	return testoplugin.Option{
		Value: option(func(p *PluginParallel) {
			p.sync = true
		}),
		Propagate: true,
	}
}
