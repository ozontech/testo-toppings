# Testo Environment Plugin

A testo plugin that loads environment variables from `.env` files for testing.

## Usage

```go
import (
    "github.com/ozontech/testo"
    "github.com/ozontech/testo-toppings/environment"
    "github.com/ozontech/testo/testoplugin"
)

func Test(t *testing.T) {
    options := []testoplugin.Option{
        environment.WithEnvironments(".env", "config.env"),
    }

    testo.RunSuite(t, new(Suite), options...)
}
```

## Features

- Loads environment variables from `.env` files
- Sets environment variables in `BeforeAll` hook
- Supports multiple `.env` files (later files overwrite earlier ones)
- Handles comments (`#`) and empty lines
- Defaults to loading `.env` from the current directory if no files specified

## Default File Loading

By default, the plugin loads `.env` from the current directory when no files are specified:
```go
func Test(t *testing.T) {
    options := []testoplugin.Option{
        environment.WithEnvironments(),  // Loads .env automatically
    }
    testo.RunSuite(t, new(Suite), options...)
}
```

## .env File Format

```
# This is a comment
KEY1=VALUE1
KEY2="quoted value"
KEY3=VALUE3
```

## Multiple Files

When multiple files are provided, values from later files overwrite earlier ones:
```go
environment.WithEnvironments("base.env", "override.env")
```

In this case, `override.env` values take precedence over `base.env` values.

## Custom Default Files

You can customize the default files loaded when no explicit files are specified:
```go
package environment

var DefaultEnviroments = []string{".env", "config.env"}
```

This will load both `.env` and `config.env` when `WithEnvironments()` is called without arguments.
