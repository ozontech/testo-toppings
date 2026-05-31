//go:build example

package main

import (
	"fmt"
	"sync/atomic"
	"testing"

	"github.com/ozontech/testo"
	"github.com/ozontech/testo-toppings/async"
)

type T struct {
	*testo.T
	*async.PluginAsync
}

type Counter struct{ n atomic.Int64 }

func (c *Counter) Inc() {
	c.n.Add(1)
}

func (c *Counter) Value() int {
	return int(c.n.Load())
}

func test(t T) {
	const workers = 10
	const incs = 100

	var counter Counter

	for i := range workers {
		async.Run(t, fmt.Sprintf("worker %d", i), func(t T) {
			for range incs {
				counter.Inc()
			}
		})
	}

	t.Wait()

	want := workers * incs
	got := counter.Value()

	if want != got {
		t.Fatalf("counter = %d, want %d", got, want)
	}
}

func Test(t *testing.T) {
	testo.RunTest(t, func(t T) {
		const workers = 10
		const incs = 100

		var counter Counter

		for i := range workers {
			async.Run(t, fmt.Sprintf("worker %d", i), func(t T) {
				for range incs {
					counter.Inc()
				}
			})
		}

		t.Wait()

		want := workers * incs
		got := counter.Value()

		if want != got {
			t.Fatalf("counter = %d, want %d", got, want)
		}
	})
}
