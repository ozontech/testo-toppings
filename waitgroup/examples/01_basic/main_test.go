//go:build example

package main

import (
	"testing"
	"time"

	"github.com/ozontech/testo"
	"github.com/ozontech/testo-toppings/waitgroup"
)

type T struct {
	*testo.T
	*waitgroup.PluginWaitGroup
}

func test(t T) {
	waitgroup.Run(t, "first sub-test", func(t T) {
		time.Sleep(5 * time.Second)

		t.Logf("finished %q", t.Name())
	})

	waitgroup.Run(t, "second sub-test", func(t T) {
		time.Sleep(5 * time.Second)

		t.Logf("finished %q", t.Name())
	})

	t.Log("waiting for sub-tests")
	t.Wait()
	t.Log("done waiting")

	for i := range 5 {
		t.Go(func() {
			time.Sleep(time.Second)

			t.Logf("finished goroutine #%d", i)
		})
	}

	// goroutines spawned above will be awaited after the test end
	// without explicit call to t.Wait()
}

func Test(t *testing.T) {
	testo.RunTest(t, test)
}
