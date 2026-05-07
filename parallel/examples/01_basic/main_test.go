//go:build example

package main

import (
	"testing"

	"github.com/ozontech/testo"
	"github.com/ozontech/testo-toppings/parallel"
)

type T struct {
	*testo.T
	*parallel.PluginParallel
}

type Suite struct{ testo.Suite[T] }

// Tests A & B are parallel.

func (Suite) TestA(t T) {}

func (Suite) TestB(t T) {}

// Test C is not parallel.

var _ = testo.For(Suite.TestA, parallel.WithSync())

func (Suite) TestC(t T) {}

func Test(t *testing.T) {
	testo.RunSuite(t, new(Suite))
}
