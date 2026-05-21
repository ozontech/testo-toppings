// Package parallel provides plugin to mark tests as parallel by default.
package parallel

import (
	"fmt"
	"sync"

	"github.com/ozontech/testo"
	"github.com/ozontech/testo/testoplugin"
	"github.com/ozontech/testo/testoreflect"
)

var _ testoplugin.Plugin = (*PluginParallel)(nil)

// PluginParallel marks all tests as parallel by default.
type PluginParallel struct {
	*testo.T

	sync  bool
	scope Scope
}

var parallelTests sync.Map

// Plugin implements [testoplugin.Plugin].
func (p *PluginParallel) Plugin(_ testoplugin.Plugin, options ...testoplugin.Option) testoplugin.Spec {
	p.sync = *flagSync
	p.scope.SuiteTests = true

	for _, opt := range options {
		if o, ok := opt.Value.(option); ok {
			o(p)
		}
	}

	return testoplugin.Spec{
		Hooks: testoplugin.Hooks{
			BeforeAll: testoplugin.Hook{
				Func: func() {
					if p.sync {
						return
					}

					if p.scope.Suites {
						p.rawParallel()
					}

					if !p.scope.Tests {
						return
					}

					t := p.root()

					if _, ok := parallelTests.LoadOrStore(t.Name(), struct{}{}); ok {
						return
					}

					defer func() {
						r := recover()
						if r == nil {
							return
						}

						if fmt.Sprint(r) == "testing: t.Parallel called multiple times" {
							return
						}

						panic(r)
					}()

					t.Parallel()
				},
			},
			BeforeEach: testoplugin.Hook{
				Func: func() {
					if p.sync {
						return
					}

					if !p.scope.SuiteTests {
						return
					}

					p.rawParallel()
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

func (p *PluginParallel) rawParallel() {
	p.Helper()

	testo.Reflect(p).TestingT.Parallel()
}

func (p *PluginParallel) root() testoreflect.TestingT {
	s := testo.Reflect(p).Suite

	root := &s

	for root.Parent != nil {
		root = root.Parent
	}

	return root.TestingT
}
