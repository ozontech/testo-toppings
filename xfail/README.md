# 🫥 XFail

Testo plugin for marking tests as "expected to fail".
Similar to [Pytest xfail functionality](https://docs.pytest.org/en/stable/how-to/skipping.html#xfail-mark-test-functions-as-expected-to-fail)

Such tests will not affect exit code of `go test` command in case of a failure.

## Quick Start

```bash
go get github.com/ozontech/testo-toppings
```

```go
type T struct {
    *testo.T
    *xfail.PluginXFail
}

type Suite struct { testo.Suite[T] }

func (Suite) Test(t T) {
    t.XFail()

    t.Fail("oops")
}
```

```txt
=== RUN   Test
=== RUN   Test/Suite
=== RUN   Test/Suite/testo!
=== RUN   Test/Suite/testo!/TestExample
    example_test.go:19: oops
    example_test.go:19: xfail: test "Test/Suite/TestExample" failed as expected, skipping
--- PASS: Test (0.00s)
    --- PASS: Test/Suite (0.00s)
        --- PASS: Test/Suite/testo! (0.00s)
            --- SKIP: Test/Suite/testo!/TestExample (0.00s)
```

See also [examples](./examples).
