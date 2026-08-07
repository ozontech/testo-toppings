//go:build parallelfixture

// Package fixture is a test fixture driven by parallel/plugin_test.go.
// It is compiled and run as a child `go test` process so that the driver
// can assert on "=== PAUSE" lines and the exit status.
package fixture

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ozontech/testo"
	"github.com/ozontech/testo-toppings/parallel"
)

type T struct {
	*testo.T
	*parallel.PluginParallel
}

var _ = testo.Options(
	parallel.WithScope(parallel.SuiteTests | parallel.Suites | parallel.Tests),
)

// running counts tests that are inside their bodies at the same time.
var running atomic.Int32

// meet blocks until every party arrives. Callers can only meet if they
// run concurrently, so a successful meet proves parallel execution.
func meet(t T, wg *sync.WaitGroup) {
	t.Helper()

	wg.Done()

	done := make(chan struct{})

	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("expected to run in parallel with the other test, but it never arrived")
	}
}

// TestSyncTable is a table of suite-less tests, each passing WithSync
// directly to testo.Test. They must not overlap.
func TestSyncTable(t *testing.T) {
	for _, name := range []string{"1", "2"} {
		t.Run(name, testo.Test(func(t T) {
			if n := running.Add(1); n > 1 {
				t.Errorf("%d tests are running at once, want them sequential", n)
			}
			defer running.Add(-1)

			time.Sleep(100 * time.Millisecond)
		}, parallel.WithSync()))
	}
}

// TestParallelTable is the same table without WithSync:
// the tests must still run in parallel.
func TestParallelTable(t *testing.T) {
	var wg sync.WaitGroup

	wg.Add(2)

	for _, name := range []string{"1", "2"} {
		t.Run(name, testo.Test(func(t T) {
			meet(t, &wg)
		}))
	}
}

var hooksMeet sync.WaitGroup

type SuiteA struct{ testo.Suite[T] }

func (SuiteA) BeforeAll(t T) { meet(t, &hooksMeet) }

func (SuiteA) TestA(t T) {}

type SuiteB struct{ testo.Suite[T] }

func (SuiteB) BeforeAll(t T) { meet(t, &hooksMeet) }

func (SuiteB) TestB(t T) {}

// TestParallelHooks proves suites still pause before their BeforeAll
// hooks: both hooks must be running at the same time.
func TestParallelHooks(t *testing.T) {
	hooksMeet.Add(2)

	testo.RunSuite(t, new(SuiteA))
	testo.RunSuite(t, new(SuiteB))
}

// TestMixed mixes a sync and a default test:
// only the default one may go parallel.
func TestMixed(t *testing.T) {
	t.Run("seq", testo.Test(func(t T) {}, parallel.WithSync()))
	t.Run("par", testo.Test(func(t T) {}))
}

// TestSyncSingle passes WithSync straight to testo.RunTest.
func TestSyncSingle(t *testing.T) {
	testo.RunTest(t, func(t T) {}, parallel.WithSync())
}

// TestSetenvRoot calls t.Setenv on the native test, which makes marking
// it parallel illegal. The plugin must fail the test, not crash the
// binary.
func TestSetenvRoot(t *testing.T) {
	t.Setenv("PARALLEL_FIXTURE_ENV", "1")

	testo.RunTest(t, func(t T) {})
}

type SyncSuite struct{ testo.Suite[T] }

func (SyncSuite) TestS(t T) {}

// TestSyncSuite passes WithSync to a regular suite.
func TestSyncSuite(t *testing.T) {
	testo.RunSuite(t, new(SyncSuite), parallel.WithSync())
}

// TestRootScopeTable keeps SuiteTests out of scope, so marking the native
// tests parallel is the only mechanism left.
func TestRootScopeTable(t *testing.T) {
	var wg sync.WaitGroup

	wg.Add(2)

	for _, name := range []string{"1", "2"} {
		t.Run(name, testo.Test(func(t T) {
			meet(t, &wg)
		}, parallel.WithScope(parallel.Suites|parallel.Tests)))
	}
}

// TestDefaultScopeTable runs a table with the default SuiteTests scope:
// testo routes each suite-less test's Parallel to its calling test,
// so the table entries still run in parallel.
func TestDefaultScopeTable(t *testing.T) {
	var wg sync.WaitGroup

	wg.Add(2)

	for _, name := range []string{"1", "2"} {
		t.Run(name, testo.Test(func(t T) {
			meet(t, &wg)
		}, parallel.WithScope(parallel.SuiteTests)))
	}
}
