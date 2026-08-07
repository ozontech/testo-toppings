package rerun_test

import (
	"errors"
	"flag"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ozontech/testo/testocache"
)

// runFixture runs `go test` on the internal/fixture package.
// cacheDir is passed via TESTO_CACHE_DIR (empty means unset),
// failNames are full test names to fail, args are extra `go test` arguments
// (testo flags like -cache.dir and -rerun.failed go after "-args").
func runFixture(
	t *testing.T,
	cacheDir string,
	failNames []string,
	args ...string,
) (out string, pass bool) {
	t.Helper()

	cmd := exec.Command(
		"go",
		append([]string{"test", "-count=1", "-v", "-tags", "rerunfixture"}, args...)...)
	cmd.Dir = "internal/fixture"
	cmd.Env = append(os.Environ(),
		"TESTO_CACHE_DIR="+cacheDir,
		"RERUN_FIXTURE_FAIL="+strings.Join(failNames, ","),
	)

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

// TestHappyPath verifies the documented flow: fail, rerun only failures,
// pass, then skip when nothing is failed.
func TestHappyPath(t *testing.T) {
	dir := t.TempDir()
	failA := []string{"Test/Suite/TestA"}
	run := []string{"-run", "Test/Suite"}
	rerun := append(run, "-args", "-rerun.failed")

	out, pass := runFixture(t, dir, failA, run...)
	assert(t, !pass, "run 1: expected failure, got pass:\n%s", out)
	assert(
		t,
		strings.Contains(out, "--- FAIL: Test/Suite/testo!/TestA"),
		"run 1: TestA did not fail:\n%s",
		out,
	)
	assert(
		t,
		strings.Contains(out, "--- PASS: Test/Suite/testo!/TestB"),
		"run 1: TestB did not pass:\n%s",
		out,
	)

	out, pass = runFixture(t, dir, failA, rerun...)
	assert(t, !pass, "run 2: expected failure, got pass:\n%s", out)
	assert(
		t,
		strings.Contains(out, "--- FAIL: Test/Suite/testo!/TestA"),
		"run 2: TestA was not rerun:\n%s",
		out,
	)
	assert(
		t,
		!strings.Contains(out, "--- PASS: Test/Suite/testo!/TestB"),
		"run 2: TestB ran, but only TestA failed before:\n%s",
		out,
	)

	out, pass = runFixture(t, dir, nil, rerun...)
	assert(t, pass, "run 3: expected pass:\n%s", out)
	assert(
		t,
		strings.Contains(out, "--- PASS: Test/Suite/testo!/TestA"),
		"run 3: TestA was not rerun:\n%s",
		out,
	)

	out, pass = runFixture(t, dir, nil, rerun...)
	assert(t, pass, "run 4: expected pass:\n%s", out)
	assert(
		t,
		strings.Contains(out, "no known test failures for suite"),
		"run 4: no message about skipped suite:\n%s",
		out,
	)
	assert(
		t,
		!strings.Contains(out, "--- PASS: Test/Suite/testo!/TestA"),
		"run 4: TestA ran, but nothing failed before:\n%s",
		out,
	)
	assert(
		t,
		strings.Contains(out, "--- SKIP: Test/Suite "),
		"run 4: suite was not visibly skipped:\n%s",
		out,
	)
	assert(
		t,
		!strings.Contains(out, "--- SKIP: Test/Suite/testo!"),
		"run 4: individual tests were marked skipped:\n%s",
		out,
	)
}

// TestSuiteless verifies the same flow for a suiteless test.
func TestSuiteless(t *testing.T) {
	dir := t.TempDir()
	failSolo := []string{"TestSolo/#00/TestSolo"}
	run := []string{"-run", "TestSolo"}
	rerun := append(run, "-args", "-rerun.failed")

	out, pass := runFixture(t, dir, failSolo, run...)
	assert(t, !pass, "run 1: expected failure, got pass:\n%s", out)

	out, pass = runFixture(t, dir, failSolo, rerun...)
	assert(t, !pass, "run 2: expected failure, got pass:\n%s", out)
	assert(
		t,
		strings.Contains(out, "--- FAIL: TestSolo/#00/testo!/TestSolo"),
		"run 2: TestSolo was not rerun:\n%s",
		out,
	)

	out, pass = runFixture(t, dir, nil, rerun...)
	assert(t, pass, "run 3: expected pass:\n%s", out)

	out, pass = runFixture(t, dir, nil, rerun...)
	assert(t, pass, "run 4: expected pass:\n%s", out)
	assert(
		t,
		strings.Contains(out, "no known test failure for test"),
		"run 4: no message about skipped test:\n%s",
		out,
	)
	assert(
		t,
		!strings.Contains(out, "--- PASS: TestSolo/#00/testo!/TestSolo"),
		"run 4: TestSolo ran, but nothing failed before:\n%s",
		out,
	)
	assert(
		t,
		strings.Contains(out, "--- SKIP: TestSolo/#00 "),
		"run 4: test was not visibly skipped:\n%s",
		out,
	)
	assert(
		t,
		!strings.Contains(out, "--- SKIP: TestSolo/#00/testo!"),
		"run 4: individual tests were marked skipped:\n%s",
		out,
	)
}

// TestCacheDirFlag verifies that -cache.dir passed as a flag (not env)
// is honored by the cache read path: the cache must not be read from
// the default directory before flags are parsed.
func TestCacheDirFlag(t *testing.T) {
	defaultDir := filepath.Join("internal", "fixture", ".testo_cache")

	if err := os.RemoveAll(defaultDir); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = os.RemoveAll(defaultDir) })

	dir := t.TempDir()
	failA := []string{"Test/Suite/TestA"}

	out, pass := runFixture(t, "", failA, "-run", "Test/Suite", "-args", "-cache.dir="+dir)
	assert(t, !pass, "run 1: expected failure, got pass:\n%s", out)

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	assert(t, len(entries) > 0, "run 1: nothing was written to -cache.dir")

	out, pass = runFixture(
		t,
		"",
		failA,
		"-run",
		"Test/Suite",
		"-args",
		"-cache.dir="+dir,
		"-rerun.failed",
	)
	assert(t, !pass, "run 2: expected failure, got pass:\n%s", out)
	assert(
		t,
		strings.Contains(out, "--- FAIL: Test/Suite/testo!/TestA"),
		"run 2: TestA was not rerun:\n%s",
		out,
	)

	_, err = os.Stat(defaultDir)
	assert(t, os.IsNotExist(err), "default cache dir was created despite -cache.dir")
}

// TestPartialRunKeepsFailures verifies that a -run filtered session
// does not erase knowledge of previously failed tests: the suite must
// not be skipped while a cached failed test still belongs to it.
func TestPartialRunKeepsFailures(t *testing.T) {
	dir := t.TempDir()
	failA := []string{"Test/Suite/TestA"}

	out, pass := runFixture(t, dir, failA, "-run", "Test/Suite")
	assert(t, !pass, "run 1: expected failure, got pass:\n%s", out)

	out, pass = runFixture(t, dir, nil, "-run", "Test/Suite/testo!/TestB")
	assert(t, pass, "run 2: expected pass:\n%s", out)
	assert(
		t,
		strings.Contains(out, "--- PASS: Test/Suite/testo!/TestB"),
		"run 2: TestB did not run:\n%s",
		out,
	)

	out, pass = runFixture(t, dir, failA, "-run", "Test/Suite", "-args", "-rerun.failed")
	assert(t, !pass, "run 3: expected failure, got pass:\n%s", out)
	assert(
		t,
		strings.Contains(out, "--- FAIL: Test/Suite/testo!/TestA"),
		"run 3: TestA was not rerun:\n%s",
		out,
	)
}

// TestSubSuitePartialRun verifies that a failed sub-suite test survives
// a -run filtered session: the outer suite's test that spawns the sub-suite
// must stay planned while any cached test below it is failed.
func TestSubSuitePartialRun(t *testing.T) {
	dir := t.TempDir()
	failX := []string{"TestSub/Outer/TestInner/Inner/TestX"}
	realX := "TestSub/Outer/testo!/TestInner/Inner/testo!/TestX"
	realY := "TestSub/Outer/testo!/TestInner/Inner/testo!/TestY"

	out, pass := runFixture(t, dir, failX, "-run", "TestSub")
	assert(t, !pass, "run 1: expected failure, got pass:\n%s", out)
	assert(t, strings.Contains(out, "--- FAIL: "+realX), "run 1: TestX did not fail:\n%s", out)

	out, pass = runFixture(t, dir, nil, "-run", realY)
	assert(t, pass, "run 2: expected pass:\n%s", out)
	assert(t, strings.Contains(out, "--- PASS: "+realY), "run 2: TestY did not run:\n%s", out)

	out, pass = runFixture(t, dir, failX, "-run", "TestSub", "-args", "-rerun.failed")
	assert(t, !pass, "run 3: expected failure, got pass:\n%s", out)
	assert(t, strings.Contains(out, "--- FAIL: "+realX), "run 3: TestX was not rerun:\n%s", out)
	assert(
		t,
		!strings.Contains(out, "--- PASS: "+realY),
		"run 3: TestY ran, but only TestX failed before:\n%s",
		out,
	)
	assert(t, !strings.Contains(out, "--- SKIP"), "run 3: something was marked skipped:\n%s", out)
}

// TestHookFailureRerun verifies that a suite that failed only in its
// BeforeAll hook (no failed tests) is fully re-run, not skipped.
func TestHookFailureRerun(t *testing.T) {
	dir := t.TempDir()
	failHook := []string{"TestHooks/Hooked"}
	run := []string{"-run", "TestHooks"}
	rerun := append(run, "-args", "-rerun.failed")

	out, pass := runFixture(t, dir, failHook, run...)
	assert(t, !pass, "run 1: expected failure, got pass:\n%s", out)

	out, pass = runFixture(t, dir, nil, rerun...)
	assert(t, pass, "run 2: expected pass:\n%s", out)
	assert(
		t,
		strings.Contains(out, "--- PASS: TestHooks/Hooked/testo!/TestH"),
		"run 2: suite was not re-run:\n%s",
		out,
	)
	assert(
		t,
		strings.Contains(out, "fixture: Hooked.BeforeAll executed"),
		"run 2: suite BeforeAll did not run:\n%s",
		out,
	)

	out, pass = runFixture(t, dir, nil, rerun...)
	assert(t, pass, "run 3: expected pass:\n%s", out)
	assert(
		t,
		strings.Contains(out, "no known test failures for suite"),
		"run 3: no message about skipped suite:\n%s",
		out,
	)
	assert(
		t,
		!strings.Contains(out, "fixture: Hooked.BeforeAll executed"),
		"run 3: suite BeforeAll ran despite the skip:\n%s",
		out,
	)
	assert(
		t,
		strings.Contains(out, "--- SKIP: TestHooks/Hooked "),
		"run 3: suite was not visibly skipped:\n%s",
		out,
	)
	assert(
		t,
		!strings.Contains(out, "--- SKIP: TestHooks/Hooked/testo!"),
		"run 3: individual tests were marked skipped:\n%s",
		out,
	)
}

// TestStaleFailureExcludesPlan verifies the fallback for a cached failure
// whose test no longer exists (e.g. renamed): the suite is not skipped
// (its hooks run), but nothing is executed and nothing is marked skipped.
func TestStaleFailureExcludesPlan(t *testing.T) {
	dir := t.TempDir()

	old := flag.Lookup("cache.dir").Value.String()
	if err := flag.Set("cache.dir", dir); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = flag.Set("cache.dir", old) })

	// A failed entry for a test that is not planned anymore,
	// written in the plugin's on-disk format.
	err := testocache.Set(
		"rerun-v2-test-"+url.PathEscape("Test/Suite/TestGone"),
		[]byte(`{"n":"Test/Suite/TestGone","f":true}`),
	)
	if err != nil {
		t.Fatal(err)
	}

	out, pass := runFixture(t, dir, nil, "-run", "Test/Suite", "-args", "-rerun.failed")
	assert(t, pass, "expected pass:\n%s", out)
	assert(
		t,
		strings.Contains(out, "match no planned test, running nothing"),
		"no message about the stale failure:\n%s",
		out,
	)
	assert(
		t,
		!strings.Contains(out, "--- PASS: Test/Suite/testo!/TestA"),
		"TestA ran, but only a stale failure exists:\n%s",
		out,
	)
	assert(t, !strings.Contains(out, "--- SKIP"), "something was marked skipped:\n%s", out)
}

// TestCacheDisabled verifies that -cache.disable with -rerun.failed
// degrades to running everything instead of silently skipping all suites.
func TestCacheDisabled(t *testing.T) {
	dir := t.TempDir()
	failA := []string{"Test/Suite/TestA"}
	run := []string{"-run", "Test/Suite"}

	out, pass := runFixture(t, dir, failA, run...)
	assert(t, !pass, "run 1: expected failure, got pass:\n%s", out)

	out, pass = runFixture(
		t,
		dir,
		failA,
		append(run, "-args", "-cache.disable", "-rerun.failed")...)
	assert(t, !pass, "run 2: expected failure, got pass:\n%s", out)
	assert(
		t,
		strings.Contains(out, "--- FAIL: Test/Suite/testo!/TestA"),
		"run 2: TestA did not run:\n%s",
		out,
	)
	assert(
		t,
		strings.Contains(out, "rerun: failed to read test statuses"),
		"run 2: no warning about unreadable statuses:\n%s",
		out,
	)
}

// TestNameCollision verifies that test names differing only by "-" vs "/"
// do not clobber each other's cache entries.
func TestNameCollision(t *testing.T) {
	dir := t.TempDir()
	failP := []string{"TestCollide/a-b/Pair/TestP"}
	run := []string{"-run", "TestCollide"}

	out, pass := runFixture(t, dir, failP, run...)
	assert(t, !pass, "run 1: expected failure, got pass:\n%s", out)
	assert(
		t,
		strings.Contains(out, "--- FAIL: TestCollide/a-b/Pair/testo!/TestP"),
		"run 1: a-b TestP did not fail:\n%s",
		out,
	)
	assert(
		t,
		strings.Contains(out, "--- PASS: TestCollide/a/b/Pair/testo!/TestP"),
		"run 1: a/b TestP did not pass:\n%s",
		out,
	)

	out, pass = runFixture(t, dir, failP, append(run, "-args", "-rerun.failed")...)
	assert(t, !pass, "run 2: expected failure, got pass:\n%s", out)
	assert(
		t,
		strings.Contains(out, "--- FAIL: TestCollide/a-b/Pair/testo!/TestP"),
		"run 2: failed a-b TestP was not rerun:\n%s",
		out,
	)
	assert(
		t,
		!strings.Contains(out, "--- PASS: TestCollide/a/b/Pair/testo!/TestP"),
		"run 2: passing a/b suite was rerun:\n%s",
		out,
	)
	assert(
		t,
		strings.Contains(out, "--- SKIP: TestCollide/a/b/Pair "),
		"run 2: passing a/b suite was not visibly skipped:\n%s",
		out,
	)
	assert(
		t,
		!strings.Contains(out, "--- SKIP: TestCollide/a/b/Pair/testo!"),
		"run 2: individual tests were marked skipped:\n%s",
		out,
	)
}
