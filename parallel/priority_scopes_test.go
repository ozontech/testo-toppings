package parallel_test

import (
	"testing"

	"github.com/ozontech/testo"
	"github.com/ozontech/testo-toppings/parallel"
	"github.com/ozontech/testo/testoplugin"
)

func TestPriorityEqualOrder(t *testing.T) {
	requireAutomaticParallel(t)

	t.Run("each parallel declared first", func(t *testing.T) {
		state := new(resourceState)
		testo.RunSuite(
			t,
			new(eachResourceSuite),
			resources(state),
			parallel.WithBeforeEachPriority(-100),
		)
		state.check(t, 4, 0)
	})
	t.Run("each resource declared first", func(t *testing.T) {
		state := new(resourceState)
		testo.RunSuite(
			t,
			new(eachResourceFirstSuite),
			resources(state),
			parallel.WithBeforeEachPriority(-100),
		)
		state.check(t, 1, 1)
	})
	t.Run("all parallel declared first", func(t *testing.T) {
		state := new(resourceState)
		t.Run("suites", func(t *testing.T) {
			testo.RunSuite(t, new(allResourceSuite), resources(state),
				parallel.WithScope(parallel.Suites), parallel.WithBeforeAllPriority(-100))
		})
		state.check(t, 1, 0)
	})
	t.Run("all resource declared first", func(t *testing.T) {
		state := new(resourceState)
		t.Run("suites", func(t *testing.T) {
			testo.RunSuite(t, new(allResourceFirstSuite), resources(state),
				parallel.WithScope(parallel.Suites), parallel.WithBeforeAllPriority(-100))
		})
		state.check(t, 1, 1)
	})
}

func TestPrioritySuiteLess(t *testing.T) {
	requireAutomaticParallel(t)

	t.Run("Test", func(t *testing.T) {
		state := new(resourceState)
		t.Run("owner", testo.Test(
			func(t eachResourceT) {},
			resources(
				state,
			),
			parallel.WithBeforeEachPriority(-200),
			parallel.WithBeforeAllPriority(-300),
		))
		// The owner's Parallel waits for this parent to return.
		t.Cleanup(func() { state.check(t, 1, 0) })
	})
	t.Run("RunTest", func(t *testing.T) {
		state := new(resourceState)
		t.Run("owner", func(t *testing.T) {
			testo.RunTest(
				t,
				func(t eachResourceT) {},
				resources(
					state,
				),
				parallel.WithBeforeEachPriority(-200),
				parallel.WithBeforeAllPriority(-300),
			)
		})
		t.Cleanup(func() { state.check(t, 1, 0) })
	})
}

func TestPriorityRepeatedSuites(t *testing.T) {
	requireAutomaticParallel(t)

	early := new(resourceState)
	testo.RunSuite(t, new(eachResourceSuite), resources(early),
		parallel.WithBeforeEachPriority(-200), parallel.WithBeforeAllPriority(-300))
	early.check(t, 4, 0)

	defaults := new(resourceState)
	testo.RunSuite(t, new(eachResourceSuite), resources(defaults))
	defaults.check(t, 4, 4)
}

func TestPriorityNativeRoots(t *testing.T) {
	requireAutomaticParallel(t)

	cases := []struct {
		name     string
		priority testoplugin.Priority
		waiting  int
	}{
		{name: "default", waiting: 4},
		{name: "early", priority: -200},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			state := new(resourceState)
			t.Run("owners", func(t *testing.T) {
				for range 4 {
					t.Run("native", func(t *testing.T) {
						testo.RunSuite(t, new(allResourceSuite), resources(state),
							parallel.WithScope(parallel.Tests),
							parallel.WithBeforeAllPriority(tc.priority))
					})
				}
				state.mu.Lock()
				waiting := state.live
				state.mu.Unlock()

				if waiting != tc.waiting {
					t.Errorf(
						"resources before native roots resume = %d, want %d",
						waiting,
						tc.waiting,
					)
				}
			})

			// Tests scope calls the native root directly, outside Testo overrides.
			state.check(t, 0, 0)

			if tc.waiting == 0 && state.peak > parallelLimit() {
				t.Errorf(
					"peak resources = %d, exceeds parallel limit %d",
					state.peak,
					parallelLimit(),
				)
			}
		})
	}
}

func TestPriorityPropagation(t *testing.T) {
	testo.RunSuite(t, new(priorityPropagationSuite), parallel.WithSync(),
		parallel.WithBeforeEachPriority(-200), parallel.WithBeforeAllPriority(-300))
}

type eachResourceFirstT struct {
	*testo.T
	*EachResourcePlugin
	*parallel.PluginParallel
}

type allResourceFirstT struct {
	*testo.T
	*AllResourcePlugin
	*parallel.PluginParallel
}

type eachResourceFirstSuite struct {
	testo.Suite[eachResourceFirstT]
}

func (*eachResourceFirstSuite) Test(t eachResourceFirstT) {}

type allResourceFirstSuite struct{ testo.Suite[allResourceFirstT] }

func (*allResourceFirstSuite) Test(t allResourceFirstT) {}

// PriorityProbe observes the options delivered by Testo to each plugin instance.
type PriorityProbe struct {
	*testo.T
	each testoplugin.Priority
	all  testoplugin.Priority
}

// Plugin records effective priorities without changing the active Parallel plugin.
func (p *PriorityProbe) Plugin(_ testoplugin.Plugin, opts ...testoplugin.Option) testoplugin.Spec {
	var independent parallel.PluginParallel

	spec := independent.Plugin(nil, opts...)
	p.each = spec.Hooks.BeforeEach.Priority
	p.all = spec.Hooks.BeforeAll.Priority

	return testoplugin.Spec{}
}

func (p *PriorityProbe) expect(each, all testoplugin.Priority) {
	p.Helper()

	if p.each != each || p.all != all {
		p.Errorf("effective priorities = (%d, %d), want (%d, %d)", p.each, p.all, each, all)
	}
}

type priorityProbeT struct {
	*testo.T
	*parallel.PluginParallel
	*PriorityProbe
}

type priorityPropagationSuite struct{ testo.Suite[priorityProbeT] }

func (*priorityPropagationSuite) BeforeAll(t priorityProbeT) {
	t.expect(-200, -300)
}

func (*priorityPropagationSuite) TestInherited(t priorityProbeT) {
	t.expect(-200, -300)

	testo.Run(t, "inherited", func(t priorityProbeT) {
		t.expect(-200, -300)
	})
	testo.Run(t, "local reset", func(t priorityProbeT) {
		t.expect(0, 0)
	}, parallel.WithBeforeEachPriority(0), parallel.WithBeforeAllPriority(0))

	testo.RunSubSuite(t, new(priorityFreshSuite), parallel.WithSync())
	testo.RunSubSuite(t, new(priorityExplicitSuite), parallel.WithSync(),
		parallel.WithBeforeEachPriority(-500), parallel.WithBeforeAllPriority(-600))
}

var _ = testo.For((*priorityPropagationSuite).TestAnnotated,
	parallel.WithBeforeEachPriority(0), parallel.WithBeforeAllPriority(50))

func (*priorityPropagationSuite) TestAnnotated(t priorityProbeT) {
	t.expect(0, 50)
}

type priorityCase struct{ value int }

func (*priorityPropagationSuite) CasesInput() []priorityCase {
	return []priorityCase{{value: 1}, {value: 2}}
}

var _ = testo.ForEach((*priorityPropagationSuite).TestCases,
	parallel.WithBeforeEachPriority(-400))

func (*priorityPropagationSuite) TestCases(t priorityProbeT, p struct{ Input priorityCase }) {
	t.expect(-400, -300)

	if p.Input.value < 1 || p.Input.value > 2 {
		t.Fatalf("unexpected case: %v", p.Input)
	}
}

type priorityFreshSuite struct{ testo.Suite[priorityProbeT] }

func (*priorityFreshSuite) BeforeAll(t priorityProbeT) { t.expect(0, 0) }
func (*priorityFreshSuite) Test(t priorityProbeT)      { t.expect(0, 0) }

type priorityExplicitSuite struct{ testo.Suite[priorityProbeT] }

func (*priorityExplicitSuite) BeforeAll(t priorityProbeT) { t.expect(-500, -600) }
func (*priorityExplicitSuite) Test(t priorityProbeT)      { t.expect(-500, -600) }
