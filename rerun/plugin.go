// Package rerun re-runs tests that failed in the previous session,
// like pytest's --last-failed.
package rerun

import (
	"flag"
	"path"
	"slices"
	"strings"
	"sync"

	"github.com/ozontech/testo"
	"github.com/ozontech/testo/testocache"
	"github.com/ozontech/testo/testoplugin"
	"github.com/ozontech/testo/testoreflect"
)

var _ testoplugin.Plugin = (*PluginRerun)(nil)

var flagFailed = flag.Bool("rerun.failed", false, "only re-run the failures from last session")

// PluginRerun re-runs tests that failed in the previous session,
// like pytest's --last-failed.
type PluginRerun struct {
	*testo.T
}

var (
	logFailedReadOnce sync.Once

	// The first call must happen after flag.Parse (-cache.dir,
	// -cache.disable) and before any cache writes. The first hook
	// satisfies both, since writes only happen in cleanups.
	readCacheOnce = sync.OnceValues(readCache)

	// globalMu makes suite cleanups atomic: testocache synchronizes
	// per operation only, so without it a green suite finishing
	// concurrently with a failed one could re-scan the namespace
	// before the failure lands and record the package as green.
	globalMu sync.Mutex
)

// Plugin implements [testoplugin.Plugin].
func (pr *PluginRerun) Plugin(testoplugin.Plugin, ...testoplugin.Option) testoplugin.Spec {
	return testoplugin.Spec{
		Hooks: pr.hooks(),
		Plan:  pr.plan(),
	}
}

func suiteKey(s testoreflect.SuiteInfo) string {
	return s.Caller + keySep + s.Name
}

// suiteTestPrefix returns the prefix of full test names belonging to s.
// Cached names are logical, but for sub-suites s.Caller is the real
// testing.T name, so strip the "testo!" wrapper segments from it.
//
// For suiteless tests (Name == "") the prefix is just the caller and may
// also match sibling suites under it: better to rerun too much than
// to lose a failure.
func suiteTestPrefix(s testoreflect.SuiteInfo) string {
	segments := slices.DeleteFunc(
		strings.Split(s.Caller, "/"),
		func(seg string) bool { return seg == "testo!" },
	)

	prefix := strings.Join(segments, "/") + "/"

	if s.Name != "" {
		prefix += s.Name + "/"
	}

	return prefix
}

// failedIn reports whether s or any cached test belonging to s failed.
// The suite flag alone is not enough: a partial session (e.g. -test.run)
// can record the suite as passed while a failed test entry still exists.
func (c cache) failedIn(s testoreflect.SuiteInfo) bool {
	return c.Suites[suiteKey(s)].Failed || c.anyFailedWithPrefix(suiteTestPrefix(s))
}

// failedUnder reports whether the named test or any cached test below it
// (e.g. inside a sub-suite it spawns) failed.
func (c cache) failedUnder(name string) bool {
	return c.Tests[name].Failed || c.anyFailedWithPrefix(name+"/")
}

func (c cache) anyFailedWithPrefix(prefix string) bool {
	for name, t := range c.Tests {
		if t.Failed && strings.HasPrefix(name, prefix) {
			return true
		}
	}

	return false
}

func (pr *PluginRerun) hooks() testoplugin.Hooks {
	return testoplugin.Hooks{
		BeforeAll:  pr.beforeAll(),
		BeforeEach: pr.beforeEach(),
	}
}

func (pr *PluginRerun) beforeAll() testoplugin.Hook {
	return testoplugin.Hook{
		Priority: testoplugin.TryFirst,
		Func: func() {
			pr.Helper()

			if !testocache.Disabled() {
				pr.Cleanup(func() {
					globalMu.Lock()
					defer globalMu.Unlock()

					r := testo.Reflect(pr)

					s := suiteEntry{
						Name:   suiteKey(r.Suite),
						Failed: pr.Failed(),
					}

					if err := s.Cache(); err != nil {
						pr.Logf("rerun: failed to cache suite: %v", err)
					}

					// Test entries are already persisted: test cleanups run
					// before the suite's, so a fresh re-scan reflects the
					// whole session up to this suite.
					c, err := readOwnCache()
					if err != nil {
						pr.Logf("rerun: failed to refresh package status: %v", err)

						return
					}

					p := pkgEntry{
						Name:   currentPkg,
						Failed: c.anyFailure(),
					}

					if err := p.Cache(); err != nil {
						pr.Logf("rerun: failed to cache package status: %v", err)
					}
				})
			}

			if !*flagFailed {
				return
			}

			c, err := readCacheOnce()
			if err != nil {
				logFailedReadOnce.Do(func() {
					pr.Helper()

					pr.Logf("rerun: failed to read test statuses: %v", err)
				})

				return
			}

			// No failure is known in this or any other package sharing
			// the cache dir: run everything instead of skipping.
			if !c.anyKnownFailure() {
				return
			}

			suite := testo.Reflect(pr).Suite

			// This skip runs before the suite's own BeforeAll and, being
			// TryFirst, before other plugins' hooks. It avoids suite
			// setup and stays out of reports: Allure registers its
			// writers in its own BeforeAll, which never runs.
			// Individual tests are never skipped, see plan().
			if !c.failedIn(suite) {
				// inside a suiteless test
				if suite.Name == "" {
					pr.Skipf(
						"rerun: there is no known test failure for test %q, skipping",
						path.Base(suite.Caller),
					)
				} else {
					pr.Skipf(
						"rerun: there are no known test failures for suite %q, skipping",
						suite.Name,
					)
				}
			}
		},
	}
}

func (pr *PluginRerun) beforeEach() testoplugin.Hook {
	return testoplugin.Hook{
		// TryFirst, so a failing BeforeEach hook in another plugin is
		// less likely to leave this session's result unrecorded.
		Priority: testoplugin.TryFirst,
		Func: func() {
			pr.Helper()

			if testocache.Disabled() {
				return
			}

			pr.Cleanup(func() {
				t := test{
					Name:   pr.Name(),
					Failed: pr.Failed(),
				}

				if err := t.Cache(); err != nil {
					pr.Logf("rerun: failed to cache test: %v", err)
				}
			})
		},
	}
}

func (pr *PluginRerun) plan() testoplugin.Plan {
	return testoplugin.Plan{
		Prepare: func(suite testoreflect.SuiteInfo, tests *[]testoplugin.PlannedTest) {
			pr.Helper()

			if !*flagFailed {
				return
			}

			c, err := readCacheOnce()
			if err != nil {
				logFailedReadOnce.Do(func() {
					pr.Helper()

					pr.Logf("rerun: failed to read test statuses: %v", err)
				})

				return
			}

			// No failure is known anywhere: keep the full plan,
			// mirroring the no-skip in beforeAll.
			if !c.anyKnownFailure() {
				return
			}

			failed := make([]testoplugin.PlannedTest, 0, len(*tests))

			for _, t := range *tests {
				if c.failedUnder(t.Info().GetName()) {
					failed = append(failed, t)
				}
			}

			if len(failed) > 0 {
				*tests = failed

				return
			}

			// Suite failed, but no planned test has a cached failure:
			// it failed in BeforeAll/AfterAll hooks, or the failed test
			// is no longer planned. Re-run everything.
			if c.Suites[suiteKey(suite)].Failed {
				return
			}

			// A cached failure exists (beforeAll did not skip), but no
			// planned test matches it: the test was renamed or removed.
			// Empty the plan rather than skip, since t.Skip here would
			// suppress AfterAll hooks and mark tests skipped in reports.
			//
			// inside a suiteless test
			if suite.Name == "" {
				pr.Logf(
					"rerun: known test failure for %q matches no planned test, running nothing",
					path.Base(suite.Caller),
				)
			} else {
				pr.Logf(
					"rerun: known test failures for suite %q match no planned test, running nothing",
					suite.Name,
				)
			}

			*tests = nil
		},
	}
}
