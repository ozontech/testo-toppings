//go:build example

package main

import (
	"os"
	"testing"

	"github.com/ozontech/testo"
	"github.com/ozontech/testo-toppings/environment"
	"github.com/ozontech/testo/testoplugin"
)

type SuiteB struct{ testo.Suite[T] }

func (SuiteB) TestA(t T) {
	testo.Run(t, "KEY-VALUE-1", func(t T) {
		val := os.Getenv("KEY1")
		if val != "VALUE1" {
			t.FailNow()
		}
	})

	testo.Run(t, "KEY-VALUE-2", func(t T) {
		val := os.Getenv("KEY2")
		if val != "VALUE2" {
			t.FailNow()
		}
	})
}

func TestB(t *testing.T) {
	options := []testoplugin.Option{
		environment.WithEnvironments(),
	}

	testo.RunSuite(t, new(SuiteA), options...)
}
