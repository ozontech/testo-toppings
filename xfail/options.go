package xfail

import "github.com/ozontech/testo/testoplugin"

type option func(x *PluginXFail)

// WithStrict will enable strict mode for xfail.
//
// In strict mode xfail will fail the test if it was marked with XFail and did not fail.
func WithStrict() testoplugin.Option {
	return testoplugin.Option{
		Value:     option(func(x *PluginXFail) { x.strict = true }),
		Propagate: true,
	}
}
