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
> With this plugin enabled, explicit calls to `t.Parallel()` are no-op.

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
