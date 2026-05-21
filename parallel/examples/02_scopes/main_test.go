//go:build example

package main

import (
	"testing"
	"time"

	"github.com/ozontech/testo"
	"github.com/ozontech/testo-toppings/parallel"
)

func init() {
	testo.Options(
		parallel.WithScope(parallel.SuiteTests | parallel.Suites | parallel.Tests),
	)
}

type T struct {
	*testo.T
	*parallel.PluginParallel
}

type Suite struct{ testo.Suite[T] }

func (Suite) TestA(T) { time.Sleep(5 * time.Second) }
func (Suite) TestB(T) { time.Sleep(5 * time.Second) }

func TestA(t *testing.T) {
	testo.RunSuite(t, new(Suite))
	testo.RunSuite(t, new(Suite))
}

func TestB(t *testing.T) {
	testo.RunSuite(t, new(Suite))
	testo.RunSuite(t, new(Suite))
}
