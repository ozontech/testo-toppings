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
