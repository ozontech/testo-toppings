// Package xfail provides plugin for pytest-like xfail functionality.
package xfail

import (
	"sync/atomic"

	"github.com/ozontech/testo"
	"github.com/ozontech/testo/testoplugin"
)

var _ testoplugin.Plugin = (*PluginXFail)(nil)

// PluginXFail is a plugin implementation to embed in T.
type PluginXFail struct {
	*testo.T

	parent *PluginXFail
	active atomic.Bool
	strict bool

	failed atomic.Bool
}

// XFail signals that this test is expected to fail.
//
// In strict mode, see [WithStrict], passed tests marked with XFail are considered failed.
func (x *PluginXFail) XFail() {
	x.active.Store(true)
}

// Plugin implements [testoplugin.Plugin].
func (x *PluginXFail) Plugin(
	parent testoplugin.Plugin,
	options ...testoplugin.Option,
) testoplugin.Spec {
	return x.plugin(parent.(*PluginXFail), options...)
}

// Plugin implements [testoplugin.Plugin].
func (x *PluginXFail) plugin(parent *PluginXFail, options ...testoplugin.Option) testoplugin.Spec {
	x.parent = parent

	for _, o := range options {
		if o, ok := o.Value.(option); ok {
			o(x)
		}
	}

	return testoplugin.Spec{
		Hooks: testoplugin.Hooks{
			AfterEach: testoplugin.Hook{
				Func: x.afterEach,
			},
		},
		Overrides: testoplugin.Overrides{
			Fail: func(f testoplugin.FuncFail) testoplugin.FuncFail {
				x.Helper()

				if !x.isActive() {
					return f
				}

				return func() {
					x.Helper()

					x.fail()

					x.Skip("xfail: failed as expected, skipping")
				}
			},
			FailNow: func(f testoplugin.FuncFailNow) testoplugin.FuncFailNow {
				x.Helper()

				if !x.isActive() {
					return f
				}

				return func() {
					x.Helper()

					x.fail()

					x.Skip("xfail: failed as expected, skipping")
				}
			},
		},
	}
}

func (x *PluginXFail) afterEach() {
	x.Helper()

	if !x.isActive() {
		return
	}

	if x.strict && !x.failed.Load() {
		// disable it so that Fatalf won't be overridden by itself.
		x.active.Store(false)

		x.Fatal("xfail: expected to fail but did not (strict mode)")
	}

	if x.failed.Load() && !x.Skipped() {
		x.SkipNow()
	}
}

func (x *PluginXFail) fail() {
	x.failed.Store(true)

	// propagate failure

	parent := x.parent

	for parent != nil {
		parent.failed.Store(true)

		parent = parent.parent
	}
}

func (x *PluginXFail) isActive() bool {
	if x.active.Load() {
		return true
	}

	parent := x.parent

	for parent != nil {
		if parent.active.Load() {
			return true
		}

		parent = parent.parent
	}

	return false
}
