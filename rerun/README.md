# 🪃 Rerun

Testo plugin for rerunning failed tests.

[Similar `--last-failed` flag from Pytest](https://docs.pytest.org/en/stable/how-to/cache.html)

## Quick Start

```bash
go get github.com/ozontech/testo-toppings
```

Add it to your `T`:

```go
package main

import (
	"github.com/ozontech/testo"
	"github.com/ozontech/testo-toppings/rerun"
)

type T struct {
	*testo.T
	*rerun.PluginRerun
}
```

Run tests as usual:

```bash
go test .
```

To rerun only failed tests run tests again with flag `-rerun.failed`:

```bash
go test . -rerun.failed
```

This flag will instruct plugin to execute tests failed in the previous run.

If there are no previous runs or no failed tests
since last run, suite will be skipped by calling `t.Skip()` underneath.
