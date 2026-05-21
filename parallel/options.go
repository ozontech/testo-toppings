package parallel

import (
	"flag"

	"github.com/ozontech/testo/testoplugin"
)

type option func(p *PluginParallel)

var flagSync = flag.Bool("parallel.sync", false, "make all tests synchronous")

// WithScope sets a [Scope] of a plugin.
// Scope defines to what extent tests should become parallel.
//
// This plugin, by default, marks as parallel only suites' tests.
// You may want to extend its reach to suites as a whole, so that,
// say, multiple runs to [testo.RunSuite] will also run in parallel.
//
//	parallel.WithScope(parallel.SuiteTests | parallel.Suites)
//
// It's also possible to mark "native" tests as parallel by adding a [Tests] scope.
//
//	parallel.WithScope(parallel.SuiteTests | parallel.Suites | parallel.Tests)
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
