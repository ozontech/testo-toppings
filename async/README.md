# 🕰️ Async

Provides a test‑aware [`sync.WaitGroup`] with configurable goroutine limits.

## Quick Start

```bash
go get github.com/ozontech/testo-toppings
```

```go
type T struct{
	*testo.T
	*async.PluginAsync
}

func Test(t *testing.T) {
	testo.RunTest(t, func(t T) {
		t.Go(func() { t.Log("hello from goroutine") })
		t.Go(func() { t.Log("hello from other") })

		t.Wait() // optional, will be called by the plugin after test completion

		t.Log("goroutines are completed")
	})
}
```

## How it works

It creates a [`sync.WaitGroup`] for each test (and sub-test) and provides a `T.Go(f func())`
method for `T`.

Wait group is awaited through `.Wait()` in `AfterAll` hook by the plugin itself.
But it's also possible to call (potentially multiple times) a `T.Wait()` method to wait for
current goroutines to finish:

```go
t.Go(func() { ... })
t.Go(func() { ... })
t.Go(func() { ... })

t.Wait() // wait for these ^3 goroutines to finish

t.Go(func() { ... })
t.Go(func() { ... })

t.Wait() // wait for these ^2 goroutines to finish

t.Go(func() { ... }) // test will wait for this goroutine to finish in the end.
```

## Difference from parallel tests

When you call `t.Parallel()` it pauses current test until all other synchronous tests are completed.
Sometimes it might be a problem.

For example, when testing a concurrent component where you need to run several operations
at the same time, then check its state inside the same test function:

```go
type T struct{
	*testo.T
	*async.PluginAsync
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
```

> `async.Run` is just a syntax sugar over calling `testo.Run` inside `t.Go`.

[`sync.WaitGroup`]: https://pkg.go.dev/sync#WaitGroup
