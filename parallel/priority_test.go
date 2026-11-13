package parallel_test

import (
	"flag"
	"sync"
	"testing"

	"github.com/ozontech/testo"
	"github.com/ozontech/testo-toppings/parallel"
	"github.com/ozontech/testo/testoplugin"
)

func TestHookPriorityContract(t *testing.T) {
	cases := []struct {
		name string
		opts []testoplugin.Option
		each testoplugin.Priority
		all  testoplugin.Priority
	}{
		{name: "defaults"},
		{
			name: "each only",
			opts: []testoplugin.Option{parallel.WithBeforeEachPriority(-200)},
			each: -200,
		},
		{name: "all only", opts: []testoplugin.Option{parallel.WithBeforeAllPriority(50)}, all: 50},
		{name: "independent", opts: []testoplugin.Option{
			parallel.WithBeforeEachPriority(testoplugin.TryFirst),
			parallel.WithBeforeAllPriority(testoplugin.TryLast),
		}, each: testoplugin.TryFirst, all: testoplugin.TryLast},
		{name: "last wins", opts: []testoplugin.Option{
			parallel.WithBeforeEachPriority(-200), parallel.WithBeforeEachPriority(30),
			parallel.WithBeforeAllPriority(-300), parallel.WithBeforeAllPriority(40),
		}, each: 30, all: 40},
		{name: "explicit reset", opts: []testoplugin.Option{
			parallel.WithBeforeEachPriority(-200), parallel.WithBeforeEachPriority(0),
			parallel.WithBeforeAllPriority(-300), parallel.WithBeforeAllPriority(0),
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var p parallel.PluginParallel

			spec := p.Plugin(nil, tc.opts...)

			if spec.Hooks.BeforeEach.Priority != tc.each ||
				spec.Hooks.BeforeAll.Priority != tc.all {
				t.Fatalf("priorities = (%d, %d), want (%d, %d)",
					spec.Hooks.BeforeEach.Priority, spec.Hooks.BeforeAll.Priority, tc.each, tc.all)
			}

			if spec.Plan.Priority != testoplugin.TryLast || spec.Overrides.Priority != 0 ||
				spec.Hooks.BeforeEachSub.Func != nil {
				t.Fatal("unrelated plugin components changed")
			}

			for _, opt := range tc.opts {
				if !opt.Propagate {
					t.Error("priority option must propagate")
				}
			}

			reset := p.Plugin(nil)

			if reset.Hooks.BeforeEach.Priority != 0 || reset.Hooks.BeforeAll.Priority != 0 {
				t.Fatal("priorities leaked into a later Plugin call")
			}
		})
	}
}

func TestBeforeEachPriorityResources(t *testing.T) {
	requireAutomaticParallel(t)

	cases := []struct {
		name     string
		opts     []testoplugin.Option
		prepared int
	}{
		{name: "default", prepared: 4},
		{name: "early", opts: []testoplugin.Option{parallel.WithBeforeEachPriority(-200)}},
		{
			name:     "late",
			opts:     []testoplugin.Option{parallel.WithBeforeEachPriority(100)},
			prepared: 4,
		},
		{
			name:     "other hook unchanged",
			opts:     []testoplugin.Option{parallel.WithBeforeAllPriority(-200)},
			prepared: 4,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			state := new(resourceState)
			opts := append([]testoplugin.Option{resources(state)}, tc.opts...)
			testo.RunSuite(t, new(eachResourceSuite), opts...)

			state.check(t, 4, tc.prepared)

			if tc.prepared == 4 && state.peak != 4 {
				t.Errorf("peak resources = %d, want 4 before any test resumes", state.peak)
			}

			if tc.prepared == 0 && state.peak > parallelLimit() {
				t.Errorf(
					"peak resources = %d, exceeds parallel limit %d",
					state.peak,
					parallelLimit(),
				)
			}
		})
	}
}

func TestBeforeAllPriorityResources(t *testing.T) {
	requireAutomaticParallel(t)

	cases := []struct {
		name     string
		opts     []testoplugin.Option
		prepared int
	}{
		{name: "default", prepared: 4},
		{name: "early", opts: []testoplugin.Option{parallel.WithBeforeAllPriority(-200)}},
		{
			name:     "late",
			opts:     []testoplugin.Option{parallel.WithBeforeAllPriority(100)},
			prepared: 4,
		},
		{
			name:     "other hook unchanged",
			opts:     []testoplugin.Option{parallel.WithBeforeEachPriority(-200)},
			prepared: 4,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			state := new(resourceState)
			opts := append(
				[]testoplugin.Option{resources(state), parallel.WithScope(parallel.Suites)},
				tc.opts...,
			)
			t.Run("suites", func(t *testing.T) {
				for range 4 {
					testo.RunSuite(t, new(allResourceSuite), opts...)
				}
			})

			state.check(t, 4, tc.prepared)

			if tc.prepared == 4 && state.peak != 4 {
				t.Errorf("peak resources = %d, want 4 before any suite resumes", state.peak)
			}

			if tc.prepared == 0 && state.peak > parallelLimit() {
				t.Errorf(
					"peak resources = %d, exceeds parallel limit %d",
					state.peak,
					parallelLimit(),
				)
			}
		})
	}
}

func TestPrioritySyncAndScope(t *testing.T) {
	cases := []struct {
		name string
		opts []testoplugin.Option
	}{
		{name: "sync", opts: []testoplugin.Option{parallel.WithSync()}},
		{name: "no scope", opts: []testoplugin.Option{parallel.WithScope(0)}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			each := new(resourceState)
			all := new(resourceState)
			opts := append([]testoplugin.Option{
				parallel.WithBeforeEachPriority(-200), parallel.WithBeforeAllPriority(-200),
			}, tc.opts...)
			testo.RunSuite(t, new(eachResourceSuite), append(opts, resources(each))...)
			testo.RunSuite(t, new(allResourceSuite), append(opts, resources(all))...)
			each.check(t, 0, 0)
			all.check(t, 0, 0)
		})
	}
}

func TestPrioritySyncFlag(t *testing.T) {
	if !flag.Lookup("parallel.sync").Value.(flag.Getter).Get().(bool) {
		t.Skip("requires -parallel.sync")
	}

	each := new(resourceState)
	all := new(resourceState)
	testo.RunSuite(
		t,
		new(eachResourceSuite),
		resources(each),
		parallel.WithBeforeEachPriority(-200),
	)
	testo.RunSuite(t, new(allResourceSuite), resources(all),
		parallel.WithScope(parallel.Suites), parallel.WithBeforeAllPriority(-200))
	each.check(t, 0, 0)
	all.check(t, 0, 0)
}

func requireAutomaticParallel(t *testing.T) {
	t.Helper()

	if flag.Lookup("parallel.sync").Value.(flag.Getter).Get().(bool) {
		t.Skip("requires automatic parallelism")
	}
}

func parallelLimit() int {
	return flag.Lookup("test.parallel").Value.(flag.Getter).Get().(int)
}

func resources(state *resourceState) testoplugin.Option {
	return testoplugin.Option{Value: state, Propagate: true}
}

func resourceOptions(opts []testoplugin.Option) *resourceState {
	for _, opt := range opts {
		if state, ok := opt.Value.(*resourceState); ok {
			return state
		}
	}

	return nil
}

type resourceState struct {
	mu       sync.Mutex
	live     int
	peak     int
	calls    int
	prepared int
	resumed  int
}

func (s *resourceState) acquire(t *testo.T) {
	s.mu.Lock()
	s.live++

	if s.live > s.peak {
		s.peak = s.live
	}
	s.mu.Unlock()

	t.Cleanup(func() {
		s.mu.Lock()
		s.live--
		s.mu.Unlock()
	})
}

func (s *resourceState) observe(allocated *bool) testoplugin.Overrides {
	return testoplugin.Overrides{
		Parallel: func(next testoplugin.FuncParallel) testoplugin.FuncParallel {
			return func() {
				s.mu.Lock()
				s.calls++

				if *allocated {
					s.prepared++
				}
				s.mu.Unlock()

				next()

				s.mu.Lock()
				s.resumed++
				s.mu.Unlock()
			}
		},
	}
}

func (s *resourceState) check(t *testing.T, calls, prepared int) {
	t.Helper()

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.calls != calls || s.prepared != prepared || s.resumed != calls || s.live != 0 {
		t.Errorf("calls/prepared/resumed/live = %d/%d/%d/%d, want %d/%d/%d/0",
			s.calls, s.prepared, s.resumed, s.live, calls, prepared, calls)
	}
}

// EachResourcePlugin allocates a resource for each suite test.
type EachResourcePlugin struct{ *testo.T }

// Plugin installs per-test resource tracking.
func (p *EachResourcePlugin) Plugin(
	_ testoplugin.Plugin,
	opts ...testoplugin.Option,
) testoplugin.Spec {
	state := resourceOptions(opts)

	if state == nil {
		// Suite-less test options arrive only on the test, not its wrapper suite.

		return testoplugin.Spec{}
	}

	allocated := false

	return testoplugin.Spec{
		Hooks: testoplugin.Hooks{
			BeforeEach: testoplugin.Hook{Priority: -100, Func: func() {
				state.acquire(p.T)
				allocated = true
			}},
		},
		Overrides: state.observe(&allocated),
	}
}

// AllResourcePlugin allocates a resource for each suite.
type AllResourcePlugin struct{ *testo.T }

// Plugin installs suite resource tracking.
func (p *AllResourcePlugin) Plugin(
	_ testoplugin.Plugin,
	opts ...testoplugin.Option,
) testoplugin.Spec {
	state := resourceOptions(opts)

	if state == nil {
		// Suite-less test options arrive only on the test, not its wrapper suite.

		return testoplugin.Spec{}
	}

	allocated := false

	return testoplugin.Spec{
		Hooks: testoplugin.Hooks{
			BeforeAll: testoplugin.Hook{Priority: -100, Func: func() {
				state.acquire(p.T)
				allocated = true
			}},
		},
		Overrides: state.observe(&allocated),
	}
}

type eachResourceT struct {
	*testo.T
	*parallel.PluginParallel
	*EachResourcePlugin
}

type allResourceT struct {
	*testo.T
	*parallel.PluginParallel
	*AllResourcePlugin
}

type eachResourceSuite struct{ testo.Suite[eachResourceT] }

func (*eachResourceSuite) TestA(t eachResourceT) {}
func (*eachResourceSuite) TestB(t eachResourceT) {}
func (*eachResourceSuite) TestC(t eachResourceT) {}
func (*eachResourceSuite) TestD(t eachResourceT) {}

type allResourceSuite struct{ testo.Suite[allResourceT] }

func (*allResourceSuite) Test(t allResourceT) {}
