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

	wg  sync.WaitGroup
	sem chan struct{}
}

// Plugin implements [testoplugin.Plugin].
func (pa *PluginAsync) Plugin(
	_ testoplugin.Plugin,
	options ...testoplugin.Option,
) testoplugin.Spec {
	for _, opt := range options {
		if o, ok := opt.Value.(option); ok {
			o(pa)
		}
	}

	return testoplugin.Spec{
		Hooks: testoplugin.Hooks{
			AfterAll:     pa.after(),
			AfterEach:    pa.after(),
			AfterEachSub: pa.after(),
		},
	}
}

// Wait blocks until all functions started from [PluginAsync.Go] (including [Run])
// are finished.
//
// Note, that calling this function is optional, as it will be called
// by the plugin after the current test or sub-test ends.
func (pa *PluginAsync) Wait() {
	pa.wg.Wait()
}

// Go calls f in a new goroutine and adds that task to the [sync.WaitGroup].
// When f returns, the task is removed from the WaitGroup.
//
// All tasks are awaited before test completion with [sync.WaitGroup.Wait].
// Use [PluginAsync.Wait] to manually await all running goroutines.
//
// The function f must not panic.
func (pa *PluginAsync) Go(f func()) {
	if pa.sem != nil {
		pa.sem <- struct{}{}
	}

	pa.wg.Add(1)

	go func() {
		pa.Helper()

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

			pa.done()
		}()

		f()
	}()
}

func (pa *PluginAsync) done() {
	if pa.sem != nil {
		<-pa.sem
	}

	pa.wg.Done()
}

func (pa *PluginAsync) after() testoplugin.Hook {
	return testoplugin.Hook{
		Priority: testoplugin.TryFirst,
		Func:     pa.Wait,
	}
}

func (pa *PluginAsync) unwrapWaitGroup() *PluginAsync {
	return pa
}

// Run calls [testo.Run] inside [PluginAsync.Go] and returns immediately.
//
// All tasks are awaited before test completion with [sync.WaitGroup.Wait].
// Use [PluginAsync.Wait] to manually await all running goroutines.
//
// # Difference from parallel tests
//
// When you call `t.Parallel()` it pauses current test until all other synchronous tests are completed.
// Sometimes it might be a problem.
//
// For example, when testing a concurrent component where you need to run several operations
// at the same time, then check its state inside the same test function:
//
//	type T struct{
//		*testo.T
//		*async.PluginAsync
//	}
//
//	func Test(t *testing.T) {
//		testo.RunTest(t, func(t T) {
//			const workers = 10
//			const incs = 100
//
//			var counter Counter
//
//			for i := range workers {
//				async.Run(t, fmt.Sprintf("worker %d", i), func(t T) {
//					for range incs {
//						counter.Inc()
//					}
//				})
//			}
//
//			t.Wait()
//
//			want := workers * incs
//			got := counter.Value()
//
//			if want != got {
//				t.Fatalf("counter = %d, want %d", got, want)
//			}
//		})
//	}
func Run[T CommonT](t T, name string, f func(t T), options ...testoplugin.Option) {
	t.unwrapWaitGroup().Go(func() {
		t.Helper()

		testo.Run(t, name, f, options...)
	})
}
