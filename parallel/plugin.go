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
// Options passed to those calls never reach the wrapper suite that testo
// builds around the test, so BeforeAll can't see [WithSync] there. The
// options do surface in Prepare as the planned test's annotations, which
// is why the decision waits until then. The wrapper has no user hooks;
// the wait only delays other plugins' BeforeAll hooks.
//
// Regular suites decide in BeforeAll: pausing before the suite's own
// BeforeAll lets its setup run in the parallel phase.
func (p *PluginParallel) plan() testoplugin.Plan {
	return testoplugin.Plan{
		// Run after filtering plugins (like rerun): a plan they emptied
		// has nothing left to parallelize.
		Priority: testoplugin.TryLast,
		Prepare: func(suite testoreflect.SuiteInfo, tests *[]testoplugin.PlannedTest) {
			p.Helper()

			// testo runs Prepare bare, unlike hooks: a panic here (say,
			// Parallel on a root test that used t.Setenv) would crash the
			// whole binary, so turn it into a test failure.
			defer func() {
				if r := recover(); r != nil {
					p.Fatalf("parallel: %v", r)
				}
			}()

			if suite.Name != "" {
				return
			}

			// Fold the tests' options into the plugin config: one test
			// for testo.Test and testo.RunTest; for anonymous-struct
			// suites lumped in here, any sync test makes the whole suite
			// sync. Nothing reads this config after Prepare, so mutating
			// it in place is fine.
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
				// Suite-less tests are handled in Prepare, see plan.
				//
				// ponytail: anonymous-struct suites also have empty names
				// and get lumped in; detect testo's singleton type if
				// that ever matters.
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

// parallelizeSuite marks, within [Suites] scope, the current suite and,
// within [Tests] scope, the native test it runs under as parallel.
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

	// Only the first suite under a native test marks it parallel; this
	// also keeps sibling suites from calling Parallel concurrently.
	// Keyed by the testing.T itself because -count reruns create a fresh
	// testing.T that has to be marked again.
	if _, ok := parallelTests.LoadOrStore(t, struct{}{}); ok {
		return
	}

	// The entry has done its job once the root finishes; drop it so long
	// -count runs don't pin every past testing.T in memory.
	t.Cleanup(func() { parallelTests.Delete(t) })

	// The map only tracks this plugin's calls; user code may have marked
	// the test parallel too.
	defer swallowRepeatedParallel()

	// No T wraps the root native test, so this call has to be direct.
	t.Parallel()
}

// parallel marks the current test parallel through T.Parallel, so other
// plugins' overrides and testo's own routing apply.
func (p *PluginParallel) parallel() {
	p.Helper()

	p.allowParallel = true
	defer func() { p.allowParallel = false }()

	// For suite-less tests testo routes Parallel to the calling test,
	// which this plugin may have marked already within [Tests] scope.
	defer swallowRepeatedParallel()

	p.Parallel()
}

// swallowRepeatedParallel recovers the panic testing throws when a test
// is marked parallel twice. Must be deferred directly.
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
