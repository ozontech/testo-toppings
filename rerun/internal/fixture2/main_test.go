//go:build rerunfixture

// Package fixture2 mirrors the fixture package with the same boilerplate
// names (func Test, type Suite) but its own test methods. Used to check
// that packages sharing one cache directory do not interfere.
package fixture2

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

func maybeFail(t T) {
	t.Helper()

	for name := range strings.SplitSeq(os.Getenv("RERUN_FIXTURE_FAIL"), ",") {
		if t.Name() == name {
			t.Errorf("failing %q as requested", name)
		}
	}
}

type Suite struct{ testo.Suite[T] }

func (Suite) TestC(t T) { maybeFail(t) }
func (Suite) TestD(t T) { maybeFail(t) }

func Test(t *testing.T) { testo.RunSuite(t, new(Suite)) }
