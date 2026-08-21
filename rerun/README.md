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

If no failure is known in **any** package sharing the cache directory
(no previous runs, or everything passed since), `-rerun.failed` runs
all tests as usual instead of skipping everything.

While some failure is known — in this package or another one — a suite
without failures of its own is skipped with `t.Skip()` before any suite
hooks run. The skip is visible in test output, but invisible to
reporting plugins (e.g. Allure), whose hooks never get a chance to
record anything.

In a suite that does have failures, non-failed tests are excluded
from the plan entirely rather than skipped, so reporting plugins do
not record previously passed tests as skipped.

The cache is kept per package, so many packages can share a single
cache directory (e.g. via `TESTO_CACHE_DIR`) without interfering.
Cross-package awareness comes from a shared registry in that directory
where each package records whether it has known failures. Two caveats:

- Packages running in parallel (`go test ./...`) update the registry
  as they finish, so a package starting late in the same session may
  observe state written earlier in that session. This only makes it
  run extra tests or skip as before — a failure is never lost.
- The registry entry of a package that was deleted or renamed lingers;
  if it recorded failures, other packages keep skipping until the
  cache directory is cleared.
