//go:build example

package main

import (
	"testing"

	"github.com/ozontech/testo"
	"github.com/ozontech/testo-toppings/xfail"
)

type T struct {
	*testo.T
	*xfail.PluginXFail
}

type Suite struct{ testo.Suite[T] }

// This test is expected to fail - and it will.
// Thus, on the first Error, Fatal, Fail, FailNow, etc. xfail will mark
// this test as skipped.
func (Suite) TestTopLevel(t T) {
	t.XFail()

	t.Error("oops")
}

// The same applies for (possibly deeply) nested tests.
func (Suite) TestInner(t T) {
	t.XFail()

	testo.Run(t, "inner 1", func(t T) {
		testo.Run(t, "inner 2", func(t T) {
			t.Error("oops")
		})
	})
}

var _ = testo.For(Suite.TestStrict, xfail.WithStrict())

// This test will not fail, but we annotated it as strict ^
// thus, it will be marked as failed explicitly by xfail plugin.
func (Suite) TestStrict(t T) {
	t.XFail()

	// nothing to do
}

func Test(t *testing.T) {
	testo.RunSuite(t, new(Suite))
}
