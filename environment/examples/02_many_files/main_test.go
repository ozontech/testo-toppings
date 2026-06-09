//go:build example

package main

import (
	"os"
	"testing"

	"github.com/ozontech/testo"
	"github.com/ozontech/testo-toppings/environment"
	"github.com/ozontech/testo/testoplugin"
)

type T struct {
	*testo.T
	*environment.PluginEnvironment
}

type Suite struct{ testo.Suite[T] }

func (Suite) TestA(t T) {
	testo.Run(t, "KEY-VALUE-1", func(t T) {
		val := os.Getenv("KEY1")
		if val != "STAND_VALUE1" {
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

func Test(t *testing.T) {
	options := []testoplugin.Option{
		environment.WithEnvironments(
			"basic.env",
			"stand.env",
		),
	}

	testo.RunSuite(t, new(Suite), options...)
}
