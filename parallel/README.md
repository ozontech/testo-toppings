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
