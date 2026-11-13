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
// say, multiple runs to [github.com/ozontech/testo.RunSuite] will also run in parallel.
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

// WithBeforeEachPriority sets the priority of the plugin's BeforeEach hook.
// The default is zero. Lower values run earlier; equal priorities follow
// Testo's plugin declaration order, even with [testoplugin.TryFirst].
//
// Use this option to pause tests before other plugins allocate per-test
// resources. It only affects automatic parallelism when [SuiteTests] is in
// scope and synchronization is not requested. It does not change BeforeAll,
// BeforeEachSub, plan, or override priorities.
//
// The option propagates to child tests. Later options override earlier ones;
// passing zero restores the default. Options do not propagate through
// [github.com/ozontech/testo.RunSubSuite].
func WithBeforeEachPriority(priority testoplugin.Priority) testoplugin.Option {
	return testoplugin.Option{
		Value: option(func(p *PluginParallel) {
			p.beforeEachPriority = priority
		}),
		Propagate: true,
	}
}

// WithBeforeAllPriority sets the priority of the plugin's BeforeAll hook.
// The default is zero. Lower values run earlier; equal priorities follow
// Testo's plugin declaration order, even with [testoplugin.TryFirst].
//
// Pass this option to [github.com/ozontech/testo.RunSuite] to pause ordinary
// suites before other plugins allocate suite resources. It affects automatic parallelism when
// [Suites] or [Tests] is in scope and synchronization is not requested.
// A test annotation cannot change its suite's already executed BeforeAll.
// Suite-less tests use a separate plan, whose priority is not changed.
//
// The option propagates to child tests, but does not change BeforeEach,
// BeforeEachSub, plan, or override priorities. Later options override earlier
// ones; passing zero restores the default. Options do not propagate through
// [github.com/ozontech/testo.RunSubSuite].
func WithBeforeAllPriority(priority testoplugin.Priority) testoplugin.Option {
	return testoplugin.Option{
		Value: option(func(p *PluginParallel) {
			p.beforeAllPriority = priority
		}),
		Propagate: true,
	}
}

// WithSync signals that this test is to be run in sync with (and only with) other sync tests.
//
// A sub-suite started inside a sync test decides parallelism on its own:
// options do not propagate through [github.com/ozontech/testo.RunSubSuite], so with [Tests] in
// scope it can still mark the enclosing native test as parallel.
func WithSync() testoplugin.Option {
	return testoplugin.Option{
		Value: option(func(p *PluginParallel) {
			p.sync = true
		}),
		Propagate: true,
	}
}
