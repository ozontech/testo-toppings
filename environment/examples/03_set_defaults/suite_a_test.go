//go:build example

package main

import (
	"os"
	"testing"

	"github.com/ozontech/testo"
	"github.com/ozontech/testo-toppings/environment"
	"github.com/ozontech/testo/testoplugin"
)

func init() {
	environment.DefaultEnviroments = []string{"default.env"}
}

type T struct {
	*testo.T
	*environment.PluginEnvironment
}

type SuiteA struct{ testo.Suite[T] }

func (SuiteA) TestA(t T) {
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

func TestA(t *testing.T) {
	options := []testoplugin.Option{
		environment.WithEnvironments(),
	}

	testo.RunSuite(t, new(SuiteA), options...)
}
