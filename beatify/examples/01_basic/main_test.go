//go:build example

package main

import (
	"testing"
	"time"

	"github.com/ozontech/testo"
	"github.com/ozontech/testo-toppings/beatify"
)

type T struct {
	*testo.T
	*beatify.PluginBeatify
}

type Suite struct{ testo.Suite[T] }

func Test(t *testing.T) {
	testo.RunSuite(t, new(Suite))
}

func (Suite) TestPass(t T) {
	time.Sleep(5 * time.Millisecond)
}

func (Suite) TestFail(t T) {
	time.Sleep(5 * time.Millisecond)
	t.Error("expected failure for the demo summary")
}

func (Suite) TestSkip(t T) {
	time.Sleep(3 * time.Millisecond)
	t.Skip("skipped on purpose")
}
