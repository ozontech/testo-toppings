// Package rerun provides plugin for pytest-like rerun functionality for failed tests.
package rerun

import (
	"flag"
	"sync"

	"github.com/ozontech/testo"
	"github.com/ozontech/testo/testocache"
	"github.com/ozontech/testo/testoplugin"
	"github.com/ozontech/testo/testoreflect"
)

var _ testoplugin.Plugin = (*PluginRerun)(nil)

var flagFailed = flag.Bool("rerun.failed", false, "only re-run the failures from last session")

// PluginRerun providers plugin for pytest-like rerun functionality for failed tests.
type PluginRerun struct {
	*testo.T
}

var (
	logFailedReadOnce sync.Once
	readCacheOnce     = sync.OnceValues(readCache)
)

// read cache as early as possible, but handle actual results later.
func init() {
	_, _ = readCacheOnce()
}

// Plugin implements [testoplugin.Plugin].
func (pr *PluginRerun) Plugin(testoplugin.Plugin, ...testoplugin.Option) testoplugin.Spec {
	return testoplugin.Spec{
		Hooks: pr.hooks(),
		Plan:  pr.plan(),
	}
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
					r := testo.Reflect(pr)

					s := suite{
						Name:   r.Suite.Caller + keySep + r.Suite.Name,
						Failed: pr.Failed(),
					}

					if err := s.Cache(); err != nil {
						pr.Logf("rerun: failed to cache suite: %v", err)
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

			suite := testo.Reflect(pr).Suite

			if !c.Suites[suite.Caller+keySep+suite.Name].Failed {
				pr.Skipf(
					"rerun: there are no known test failures for suite %q, skipping",
					suite.Name,
				)
			}
		},
	}
}

func (pr *PluginRerun) beforeEach() testoplugin.Hook {
	return testoplugin.Hook{
		Func: func() {
			pr.Helper()

			if testocache.Disabled() {
				return
			}

			pr.Cleanup(func() {
				t := test{
					Name:   pr.Name(),
					Failed: pr.Failed(),
					Suite:  testo.Reflect(pr).Suite.Name,
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

			failed := make([]testoplugin.PlannedTest, 0, len(*tests))

			for _, t := range *tests {
				cachedTest := c.Tests[t.Info().GetName()]

				if cachedTest.Failed {
					failed = append(failed, t)
				}
			}

			if len(failed) > 0 {
				*tests = failed

				return
			}

			// Suite failed, but no actual tests were failed.
			// It means suite failed in BeforeAll or/and AfterAll hooks.
			if c.Suites[suite.Caller+"/"+suite.Name].Failed {
				return
			}

			pr.Skipf(
				"rerun: there are no known test failures for suite %q, skipping",
				suite.Name,
			)
		},
	}
}
