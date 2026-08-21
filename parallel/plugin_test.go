package parallel_test

import (
	"errors"
	"os/exec"
	"strings"
	"testing"
)

// runFixture runs `go test` on the internal/fixture package and returns its
// verbose output. The tests below assert on "=== PAUSE" lines: they appear
// exactly when a test calls t.Parallel, so their presence or absence shows
// what the plugin decided.
func runFixture(t *testing.T, args ...string) (out string, pass bool) {
	t.Helper()

	cmd := exec.Command(
		"go",
		append(
			[]string{"test", "-count=1", "-v", "-parallel", "4", "-tags", "parallelfixture"},
			args...,
		)...)
	cmd.Dir = "internal/fixture"

	b, err := cmd.CombinedOutput()
	if err != nil {
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) {
			t.Fatalf("running fixture: %v\n%s", err, b)
		}
	}

	return string(b), err == nil
}

func assert(t *testing.T, cond bool, format string, args ...any) {
	t.Helper()

	if !cond {
		t.Errorf(format, args...)
	}
}

// TestSyncOption verifies that WithSync passed to testo.Test keeps the
// whole chain sequential: neither the test, nor the wrapper suite around
// it, nor the native test above may go parallel.
func TestSyncOption(t *testing.T) {
	out, pass := runFixture(t, "-run", "TestSyncTable")
	assert(t, pass, "expected pass:\n%s", out)
	assert(t, strings.Contains(out, "--- PASS: TestSyncTable"), "test did not run:\n%s", out)
	assert(
		t,
		!strings.Contains(out, "=== PAUSE TestSyncTable"),
		"a sync test went parallel:\n%s",
		out,
	)
}

// TestParallelByDefault verifies that the same table without WithSync
// still runs in parallel (the fixture tests block until they overlap).
func TestParallelByDefault(t *testing.T) {
	out, pass := runFixture(t, "-run", "TestParallelTable")
	assert(t, pass, "tests did not overlap:\n%s", out)
	assert(
		t,
		strings.Contains(out, "=== PAUSE TestParallelTable"),
		"tests were not marked parallel:\n%s",
		out,
	)
}

// TestBeforeAllStaysParallel verifies that regular suites still pause
// before their BeforeAll hooks, so expensive setup runs concurrently.
func TestBeforeAllStaysParallel(t *testing.T) {
	out, pass := runFixture(t, "-run", "TestParallelHooks")
	assert(t, pass, "BeforeAll hooks did not overlap:\n%s", out)
	assert(t, strings.Contains(out, "--- PASS: TestParallelHooks"), "test did not run:\n%s", out)
}

// TestMixedTable verifies that sync and default tests can coexist:
// the sync one stays sequential while the other goes parallel.
func TestMixedTable(t *testing.T) {
	out, pass := runFixture(t, "-run", "TestMixed")
	assert(t, pass, "expected pass:\n%s", out)
	assert(
		t,
		!strings.Contains(out, "=== PAUSE TestMixed/seq"),
		"the sync test went parallel:\n%s",
		out,
	)
	assert(
		t,
		strings.Contains(out, "=== PAUSE TestMixed/par"),
		"the default test was not marked parallel:\n%s",
		out,
	)
}

// TestGlobalSyncFlag verifies that -parallel.sync disables parallelism
// for everything, including tests without an explicit WithSync.
func TestGlobalSyncFlag(t *testing.T) {
	out, pass := runFixture(t, "-run", "TestMixed", "-args", "-parallel.sync")
	assert(t, pass, "expected pass:\n%s", out)
	assert(t, strings.Contains(out, "--- PASS: TestMixed"), "test did not run:\n%s", out)
	assert(t, !strings.Contains(out, "=== PAUSE"), "a test went parallel:\n%s", out)
}

// TestRunTestSync verifies WithSync passed directly to testo.RunTest.
func TestRunTestSync(t *testing.T) {
	out, pass := runFixture(t, "-run", "TestSyncSingle")
	assert(t, pass, "expected pass:\n%s", out)
	assert(t, strings.Contains(out, "--- PASS: TestSyncSingle"), "test did not run:\n%s", out)
	assert(t, !strings.Contains(out, "=== PAUSE"), "the test went parallel:\n%s", out)
}

// TestSuiteSyncOption verifies WithSync passed to testo.RunSuite keeps
// a regular suite sequential.
func TestSuiteSyncOption(t *testing.T) {
	out, pass := runFixture(t, "-run", "TestSyncSuite")
	assert(t, pass, "expected pass:\n%s", out)
	assert(t, strings.Contains(out, "--- PASS: TestSyncSuite"), "test did not run:\n%s", out)
	assert(t, !strings.Contains(out, "=== PAUSE"), "the sync suite went parallel:\n%s", out)
}

// TestSetenvRootFails verifies that a root test using t.Setenv, which
// cannot be marked parallel, fails cleanly instead of crashing the
// test binary.
func TestSetenvRootFails(t *testing.T) {
	out, pass := runFixture(t, "-run", "TestSetenvRoot")
	assert(t, !pass, "expected failure:\n%s", out)
	assert(t, strings.Contains(out, "--- FAIL: TestSetenvRoot"), "test did not fail:\n%s", out)
	assert(
		t,
		strings.Contains(out, "can not use t.Parallel"),
		"test failed for an unexpected reason:\n%s",
		out,
	)
	assert(t, !strings.Contains(out, "\npanic:"), "the test binary crashed:\n%s", out)
}

// TestDefaultScope verifies that a suite-less table runs in parallel even
// with the default SuiteTests scope: testo routes each test's Parallel
// call to its calling test.
func TestDefaultScope(t *testing.T) {
	out, pass := runFixture(t, "-run", "TestDefaultScopeTable")
	assert(t, pass, "tests did not overlap:\n%s", out)
	assert(
		t,
		strings.Contains(out, "=== PAUSE TestDefaultScopeTable"),
		"tests were not marked parallel:\n%s",
		out,
	)
}

// TestRepeatedRuns verifies -count reruns stay parallel: every iteration
// gets a fresh testing.T that must be marked parallel again. The fixture
// test's scope leaves no other source of parallelism, so a stale mark
// from the first run would make the second run sequential.
func TestRepeatedRuns(t *testing.T) {
	out, pass := runFixture(t, "-run", "TestRootScopeTable", "-count=2")
	assert(t, pass, "tests did not overlap on some iteration:\n%s", out)
	assert(
		t,
		strings.Count(out, "--- PASS: TestRootScopeTable ") == 2,
		"expected two iterations:\n%s",
		out,
	)
}
