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

	// allowParallel lifts the Overrides.Parallel gate while the plugin
	// itself is calling T.Parallel.
	allowParallel bool
}

var parallelTests sync.Map

// Plugin implements [testoplugin.Plugin].
func (p *PluginParallel) Plugin(
	_ testoplugin.Plugin,
	options ...testoplugin.Option,
) testoplugin.Spec {
	p.sync = *flagSync
	p.scope = SuiteTests

	for _, opt := range options {
		if o, ok := opt.Value.(option); ok {
			o(p)
		}
	}

	return testoplugin.Spec{
		Plan:  p.plan(),
		Hooks: p.hooks(),
		Overrides: testoplugin.Overrides{
			Parallel: func(f testoplugin.FuncParallel) testoplugin.FuncParallel {
				return func() {
					if p.allowParallel {
						f()

						return
					}

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

// plan decides suite and root parallelism for suite-less tests
// ([testo.Test] and [testo.RunTest]).
//
// Regular suites decide in BeforeAll: pausing before the suite's own
// BeforeAll lets its setup run in the parallel phase.
func (p *PluginParallel) plan() testoplugin.Plan {
	return testoplugin.Plan{
		Priority: testoplugin.TryLast,
		Prepare: func(suite testoreflect.SuiteInfo, tests *[]testoplugin.PlannedTest) {
			p.Helper()

			// a panic here (say, Parallel on a root test that used t.Setenv)
			// would crash the whole binary, so turn it into a test failure.
			defer func() {
				if r := recover(); r != nil {
					p.Fatalf("parallel: %v", r)
				}
			}()

			if suite.Name != "" {
				return
			}

			seen := false

			for _, t := range *tests {
				if t == nil {
					continue
				}

				seen = true

				for _, opt := range t.Annotations() {
					if o, ok := opt.Value.(option); ok {
						o(p)
					}
				}
			}

			if !seen {
				return
			}

			p.parallelizeSuite()
		},
	}
}

func (p *PluginParallel) hooks() testoplugin.Hooks {
	return testoplugin.Hooks{
		BeforeAll: testoplugin.Hook{
			Func: func() {
				if testo.Reflect(p).Suite.Name == "" {
					return
				}

				p.parallelizeSuite()
			},
		},
		BeforeEach: testoplugin.Hook{
			Func: func() {
				if p.sync {
					return
				}

				if !p.scope.has(SuiteTests) {
					return
				}

				p.parallel()
			},
		},
	}
}

func (p *PluginParallel) parallelizeSuite() {
	if p.sync {
		return
	}

	if p.scope.has(Suites) {
		p.parallel()
	}

	if !p.scope.has(Tests) {
		return
	}

	t := p.root()

	if _, ok := parallelTests.LoadOrStore(t, struct{}{}); ok {
		return
	}

	t.Cleanup(func() { parallelTests.Delete(t) })

	defer swallowRepeatedParallel()

	t.Parallel()
}

func (p *PluginParallel) parallel() {
	p.Helper()

	p.allowParallel = true
	defer func() { p.allowParallel = false }()

	defer swallowRepeatedParallel()

	p.Parallel()
}

func swallowRepeatedParallel() {
	r := recover()
	if r == nil {
		return
	}

	if fmt.Sprint(r) != "testing: t.Parallel called multiple times" {
		panic(r)
	}
}

func (p *PluginParallel) root() testoreflect.TestingT {
	s := testo.Reflect(p).Suite

	root := &s

	for root.Parent != nil {
		root = root.Parent
	}

	return root.TestingT
}
