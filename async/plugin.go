// Package async provides a test‑aware [sync.WaitGroup].
package async

import (
	"sync"

	"github.com/ozontech/testo"
	"github.com/ozontech/testo/testoplugin"
)

var (
	_ CommonT            = (*PluginAsync)(nil)
	_ testoplugin.Plugin = (*PluginAsync)(nil)
)

// CommonT is an interface common for all Ts with [PluginAsync] installed.
type CommonT interface {
	testo.CommonT

	Go(f func())
	Wait()

	unwrapWaitGroup() *PluginAsync
}

// PluginAsync simplifies goroutine spawning in tests.
type PluginAsync struct {
	*testo.T

	wg sync.WaitGroup
}

// Plugin implements [testoplugin.Plugin].
func (wg *PluginAsync) Plugin(testoplugin.Plugin, ...testoplugin.Option) testoplugin.Spec {
	return testoplugin.Spec{
		Hooks: testoplugin.Hooks{
			AfterAll:     wg.after(),
			AfterEach:    wg.after(),
			AfterEachSub: wg.after(),
		},
	}
}

func (wg *PluginAsync) after() testoplugin.Hook {
	return testoplugin.Hook{
		Priority: testoplugin.TryFirst,
		Func:     wg.Wait,
	}
}

// Wait blocks until all functions started from [PluginAsync.Go] (including [Run])
// are finished.
//
// Note, that calling this function is optional, as it will be called
// by the plugin after the current test or sub-test ends.
func (wg *PluginAsync) Wait() {
	wg.wg.Wait()
}

// Go calls f in a new goroutine and adds that task to the [sync.WaitGroup].
// When f returns, the task is removed from the WaitGroup.
//
// All tasks are awaited before test completion with [sync.WaitGroup.Wait].
// Use [PluginAsync.Wait] to manually await all running goroutines.
//
// The function f must not panic.
func (wg *PluginAsync) Go(f func()) {
	// TODO(metafates): use [sync.WaitGroup.Go] once available.

	wg.wg.Add(1)

	go func() {
		wg.Helper()

		defer func() {
			if x := recover(); x != nil {
				// f panicked, which will be fatal because
				// this is a new goroutine.
				//
				// Calling Done will unblock Wait in the main goroutine,
				// allowing it to race with the fatal panic and
				// possibly even exit the process (os.Exit(0))
				// before the panic completes.
				//
				// This is almost certainly undesirable,
				// so instead avoid calling Done and simply panic.
				panic(x)
			}

			wg.wg.Done()
		}()

		f()
	}()
}

func (wg *PluginAsync) unwrapWaitGroup() *PluginAsync {
	return wg
}

// Run calls [testo.Run] inside [PluginAsync.Go] and returns immediately.
//
// All tasks are awaited before test completion with [sync.WaitGroup.Wait].
// Use [PluginAsync.Wait] to manually await all running goroutines.
func Run[T CommonT](t T, name string, f func(t T), options ...testoplugin.Option) {
	t.unwrapWaitGroup().Go(func() {
		t.Helper()

		testo.Run(t, name, f, options...)
	})
}
