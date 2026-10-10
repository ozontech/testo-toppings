# Beatify

Testo plugin for colored per-test status lines and a suite summary.

The summary includes pass/fail/skip counts, durations, clickable source
locations (OSC 8 / `vscode://file`), and a pass-rate line
(`coverage: N%` = passed/total — not `go test -cover`).

## Quick Start

```bash
go get github.com/ozontech/testo-toppings
```

```go
package main

import (
	"github.com/ozontech/testo"
	"github.com/ozontech/testo-toppings/beatify"
)

type T struct {
	*testo.T
	*beatify.PluginBeatify
}
```

```bash
go test .
```

## Flags

| Flag | Effect |
| --- | --- |
| `-beatify.disable` | plain text, no ANSI colors / OSC 8 links |
| `-beatify.quiet` | suppress per-test status lines; keep the summary |
| `-beatify.coverage-only` | print only `coverage: N%` (test chatter silenced) |

`go test` still prints its own package `ok` / `FAIL` line after the run.

See also [examples](./examples).
