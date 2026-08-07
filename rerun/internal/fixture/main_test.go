//go:build rerunfixture

// Package fixture is a test fixture driven by rerun/plugin_test.go.
// It is compiled and run as a child `go test` process, with failures
// injected via the RERUN_FIXTURE_FAIL environment variable.
package fixture

import (
	"os"
	"strings"
	"testing"

	"github.com/ozontech/testo"
	"github.com/ozontech/testo-toppings/rerun"
)

type T struct {
	*testo.T
	*rerun.PluginRerun
}

// maybeFail fails the test if its full name is listed
// in the comma-separated RERUN_FIXTURE_FAIL env variable.
func maybeFail(t T) {
	t.Helper()

	for name := range strings.SplitSeq(os.Getenv("RERUN_FIXTURE_FAIL"), ",") {
		if t.Name() == name {
			t.Errorf("failing %q as requested", name)
		}
	}
}

type Suite struct{ testo.Suite[T] }

func (Suite) TestA(t T) { maybeFail(t) }
func (Suite) TestB(t T) { maybeFail(t) }

func Test(t *testing.T) { testo.RunSuite(t, new(Suite)) }

func TestSolo(t *testing.T) { testo.RunTest(t, maybeFail) }

// Hooked can fail in BeforeAll: in its scope t.Name() is the suite
// root name "TestHooks/Hooked".
type Hooked struct{ testo.Suite[T] }

func (Hooked) BeforeAll(t T) {
	t.Log("fixture: Hooked.BeforeAll executed")
	maybeFail(t)
}

func (Hooked) TestH(t T) { maybeFail(t) }

func TestHooks(t *testing.T) { testo.RunSuite(t, new(Hooked)) }

type Outer struct{ testo.Suite[T] }

func (Outer) TestInner(t T) { testo.RunSubSuite(t, new(Inner)) }

type Inner struct{ testo.Suite[T] }

func (Inner) TestX(t T) { maybeFail(t) }
func (Inner) TestY(t T) { maybeFail(t) }

func TestSub(t *testing.T) { testo.RunSuite(t, new(Outer)) }

type Pair struct{ testo.Suite[T] }

func (Pair) TestP(t T) { maybeFail(t) }

// TestCollide runs two suites whose test and suite names differ
// only by "-" vs "/", so a lossy cache key normalization makes
// their cache entries clobber each other.
func TestCollide(t *testing.T) {
	t.Run("a-b", func(t *testing.T) { testo.RunSuite(t, new(Pair)) })
	t.Run("a/b", func(t *testing.T) { testo.RunSuite(t, new(Pair)) })
}
