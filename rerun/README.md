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

If there are no previous runs or no failed tests since last run, the
whole suite is skipped with `t.Skip()` before any suite hooks run.
The skip is visible in test output, but invisible to reporting plugins
(e.g. Allure), whose hooks never get a chance to record anything.

In a suite that does have failures, non-failed tests are excluded
from the plan entirely rather than skipped, so reporting plugins do
not record previously passed tests as skipped.

The cache is kept per package, so many packages can share a single
cache directory (e.g. via `TESTO_CACHE_DIR`) without interfering.
