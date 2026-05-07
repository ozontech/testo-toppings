// Package parallel provides plugin to mark tests as parallel by default.
package parallel

import (
	"github.com/ozontech/testo"
	"github.com/ozontech/testo/testoplugin"
	"github.com/ozontech/testo/testoreflect"
)

var _ testoplugin.Plugin = (*PluginParallel)(nil)

// PluginParallel marks all tests as parallel by default.
type PluginParallel struct {
	*testo.T

	sync bool
}

// Plugin implements [testoplugin.Plugin].
func (p *PluginParallel) Plugin(_ testoplugin.Plugin, options ...testoplugin.Option) testoplugin.Spec {
	p.sync = *flagSync

	for _, opt := range options {
		if o, ok := opt.Value.(option); ok {
			o(p)
		}
	}

	return testoplugin.Spec{
		Hooks: testoplugin.Hooks{
			BeforeEach: testoplugin.Hook{
				Func: func() {
					if p.sync {
						return
					}

					r := testo.Reflect(p)

					r.TestingT.Parallel()
				},
			},
		},
		Overrides: testoplugin.Overrides{
			Parallel: func(f testoplugin.FuncParallel) testoplugin.FuncParallel {
				return func() {
					regular, ok := testo.Reflect(p).Test.(testoreflect.RegularTestInfo)
					if !ok {
						return
					}

					if regular.IsSubtest {
						f()
					}
				}
			},
		},
	}
}
