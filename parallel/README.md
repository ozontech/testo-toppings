# 🤹 Parallel

Testo plugin for making all tests parallel by default.

> Go has great support for parallel tests.
> But marking a test as parallel requires an explicit `t.Parallel()` call.
> This plugin does that for you!

## Quick Start

```bash
go get github.com/ozontech/testo-toppings
```

Add it to your `T`:

```go
package main

import (
	"github.com/ozontech/testo"
	"github.com/ozontech/testo-toppings/parallel"
)

type T struct {
	*testo.T
	*parallel.PluginParallel
}
```

All your tests are now parallel by default!

> [!NOTE]
> Explicit calls to `t.Parallel()` in top-level Testo tests are no-op.
> Explicit calls in ordinary Testo subtests still pass through; native Go tests
> retain their own `testing.T.Parallel()` behavior.

Use annotations to mark any test as synchronous:

```go
var _ = testo.For(Suite.TestFoo, parallel.WithSync())

func (Suite) TestFoo(t T) {
    // ...
}
```

To make all tests synchronous pass `-parallel.sync` flag to `go test`:

```bash
go test ./... -parallel.sync
```

## Scopes

Tests can be parallelized to a different extents.

This is configured through "scopes".

Available scopes:

```go
const (
	// SuiteTests is a [Scope] that covers suite tests.
	//
	// For example:
	//
	// 	func (Suite) TestA(t T) { ... }
	// 	func (Suite) TestB(t T) { ... }
	//
	// Tests A & B will be marked as parallel.
	//
	// This is a default value.
	SuiteTests Scope = 1 << iota

	// Suites is a [Scope] that covers suites but not their tests.
	//
	// For example:
	//
	// 	func Test(t *testing.T) {
	//		testo.RunSuite(t, new(Suite))
	//		testo.RunSuite(t, new(OtherSuite))
	// 	}
	//
	// Both of these suites will be run in parallel.
	Suites

	// Tests is a [Scope] that covers native tests.
	//
	// For example:
	//
	// 	func TestA(t *testing.T) {
	//		testo.RunSuite(t, new(Suite))
	// 	}
	//
	// 	func TestA(t *testing.T) {
	//		testo.RunSuite(t, new(OtherSuite))
	// 	}
	//
	// Tests A & B will be run in parallel.
	Tests
)
```

Scopes are configured with `WithScope` option:

```go
// WithScope sets a [Scope] of a plugin.
// Scope defines to what extent tests should become parallel.
//
// This plugin, by default, marks as parallel only suites' tests.
// You may want to extend its reach to suites as a whole, so that,
// say, multiple runs to [testo.RunSuite] will also run in parallel.
//
//	parallel.WithScope(parallel.SuiteTests | parallel.Suites)
//
// It's also possible to mark "native" tests as parallel by adding a [Tests] scope.
//
//	parallel.WithScope(parallel.SuiteTests | parallel.Suites | parallel.Tests)
func WithScope(scope Scope) testoplugin.Option
```

## Run resource setup after the parallel wait

Parallelism starts at the call to `Parallel()`, not at the beginning of a test.
Go pauses that call until the parent function returns and a slot is available
under `go test -parallel`. Code before the call has already run.

The plugin's `BeforeEach` and `BeforeAll` hooks both have priority **0** by
default. If another plugin allocates resources at priority `-100`, those
resources exist before Parallel pauses execution. The parent can then start
more tests or suites, each allocating resources before waiting. A small
`-parallel` value does not bound this waiting population.

Two independent options let you move the pause ahead of resource setup:

```go
testo.RunSuite(t, new(Suite),
    parallel.WithBeforeEachPriority(-200),
)
```

For resources allocated once per ordinary suite, enable suite parallelism and
configure its hook separately:

```go
testo.RunSuite(t, new(Suite),
    parallel.WithScope(parallel.Suites),
    parallel.WithBeforeAllPriority(-200),
)
```

Both can be configured together:

```go
testo.RunSuite(t, new(Suite),
    parallel.WithScope(parallel.SuiteTests | parallel.Suites),
    parallel.WithBeforeEachPriority(-200),
    parallel.WithBeforeAllPriority(-200),
)
```

Lower numbers run earlier. Choose a value lower than the resource plugin's
corresponding hook. `testoplugin.TryFirst` is the smallest numeric priority,
but does not guarantee absolute precedence: equal priorities follow Testo's
plugin declaration order, including its traversal of embedded plugins.

### Example: SQL connections held by waiting tests

Suppose a database plugin has this resource hook at priority `-100`.
This illustrative fragment uses a `database/sql/driver.Connector` supplied by
the caller; the actual SQL driver and database are application-specific.
The plugin can retain the DB handle for use by the test.

```go
func databaseHook(t *testo.T, connector driver.Connector) testoplugin.Hook {
    return testoplugin.Hook{
        Priority: -100,
        Func: func() {
            db := sql.OpenDB(connector)
            t.Cleanup(func() {
                if err := db.Close(); err != nil {
                    t.Errorf("close database: %v", err)
                }
            })

            db.SetMaxOpenConns(4)
            db.SetMaxIdleConns(1)
            if err := db.PingContext(t.Context()); err != nil {
                t.Fatalf("connect to database: %v", err)
            }
        },
    }
}
```

Install this hook as the database plugin's `BeforeEach` for per-test pools,
or as its `BeforeAll` for per-suite pools. Configure the matching Parallel
priority to `-200`, then run, for example:

```sh
go test ./... -parallel=2
```

With the default priority, each test can create a pool, open a connection via
`PingContext`, and then pause. The pool can retain that idle connection while
the test waits. SQL queries, migrations, or pool warm-up can retain further
resources before the test body starts. Across many waiting tests this can
exhaust database connection capacity even though only a few test bodies run.

With the earlier priority, the sequence becomes:

1. Parallel pauses the test.
2. Go admits the test to the parallel execution window.
3. The database hook creates its pool and opens a connection.
4. The test runs; its cleanup closes the pool when the owner and its subtests finish.

The same sequence applies to suite setup with `WithBeforeAllPriority` and
the `Suites` scope. This avoids opening connections solely to hold them while
waiting for the initial parallel slot. Setup after the pause may now run
concurrently, so the resource plugin must support concurrent setup.

For illustration, 100 leaf tests that each retain one connection before pausing
could hold 100 connections with `-parallel=2`. Moving that allocation after
the pause limits this particular isolated scenario to two live test pools,
provided each closes before releasing its execution slot. These are hypothetical
counts, not measurements or a general connection quota.

Creating a pool is not the same as opening all its connections.
`sql.Open` may not connect at all, and `SetMaxOpenConns(4)` is a ceiling,
not a request to open four connections. Actual connection counts depend on the
driver, queries, warm-up, idle retention, and other pool settings.

### Scope, inheritance, and limits

- Without these options, both priorities remain zero. An explicit zero restores
  the default. Repeated values use the last option; the two settings are independent.
- Options propagate along Testo's test hierarchy. Suite options precede test
  annotations, so `testo.For` and `testo.ForEach` can override BeforeEach
  priority for a test or each parametrized case.
- Pass BeforeAll priority to `testo.RunSuite` or `testo.Options`. A test
  annotation cannot reorder a suite's already executed BeforeAll.
- `testo.RunSubSuite` does not inherit parent options automatically. Pass the
  options explicitly, or use process-wide `testo.Options` intentionally.
- BeforeEach priority affects automatic parallelism only with `SuiteTests`.
  BeforeAll priority affects ordinary suites with `Suites` or `Tests`.
  Neither option enables a scope or overrides `WithSync` or `-parallel.sync`.
- Suite-less `testo.Test` and `testo.RunTest` retain Testo's routing of
  BeforeEach Parallel to the calling Go test. Their suite/root decisions in
  `Plan.Prepare` are unchanged; their Parallel BeforeAll hook remains a no-op.
- These options do not change plan or override priorities, and do not add a
  `BeforeEachSub` hook. Explicit Parallel calls in ordinary subtests remain
  the caller's responsibility. A third-party Parallel override that allocates
  before calling its next function, or suppresses it, can change the outcome.
- The ordinary `Suite.BeforeEach` and `Suite.BeforeAll` methods already
  run after the corresponding plugin hooks. The examples address earlier
  **plugin** hooks, not an ordering defect in those suite methods.
- Resources created in plugin initialization, case providers, plans, or native
  runner code are not moved by these settings.
- Suite resources live for the suite. Parents can retain resources while yielding
  execution slots to parallel subtests. Neither option guarantees that all live
  resources are bounded by `-parallel`.
- Register cleanup on the resource owner. `AfterEach` can run before parallel
  subtests finish; early Parallel does not repair incorrect cleanup or leaks.
- `-parallel` applies within one test binary. Concurrent packages, processes,
  CI jobs, and other database clients add their own load.
- Synchronous exceptions do not acquire a new parallel wait. They can still have
  parallel ancestors, and sync options do not disable explicit native Parallel
  calls or explicit Parallel in ordinary subtests.

See Go's [parallel subtest semantics](https://go.dev/blog/subtests#control-of-parallelism)
and the [database/sql connection pool contract](https://pkg.go.dev/database/sql#DB).
