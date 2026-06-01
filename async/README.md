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
		async.Run(t, "foo", func(t T) { t.Log("hello from goroutine") })
		async.Run(t, "bar", func(t T) { t.Log("hello from other") })

		t.Wait() // optional, will be called by the plugin after test completion

		t.Log("goroutines are completed")
	})
}
```

## How it works

It creates a [`sync.WaitGroup`] for each test (and sub-test) and provides a `Run` function to run sub-tests in independent goroutines.

Wait group is awaited through `Wait` in `AfterAll` hook by the plugin itself.
If at least one test called [`t.FailNow`] inside `Run`, `Wait` will propagate it.

It's also possible to call (potentially multiple times) a `Wait` method to wait for
current goroutines to finish:

```go
async.Run(t, "a", func(t T) { ... })
async.Run(t, "b", func(t T) { ... })
async.Run(t, "c", func(t T) { ... })

t.Wait() // wait for these ^3 goroutines to finish

async.Run(t, "a", func(t T) { ... })
async.Run(t, "b", func(t T) { ... })

t.Wait() // wait for these ^2 goroutines to finish

// test will wait for this goroutine to finish in the end.
async.Run(t, "a", func(t T) { ... }) 
```

## Difference from parallel tests

When you call [`t.Parallel`] it pauses current test until all other synchronous tests are completed.
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

[`sync.WaitGroup`]: https://pkg.go.dev/sync#WaitGroup
[`t.Parallel`]: https://pkg.go.dev/github.com/ozontech/testo#T.Parallel
[`t.FailNow`]: https://pkg.go.dev/github.com/ozontech/testo#T.FailNow
