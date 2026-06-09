# 🍕 Testo Toppings

[![Go Reference](https://pkg.go.dev/badge/github.com/ozontech/testo-toppings.svg)](https://pkg.go.dev/github.com/ozontech/testo-toppings)

> Add some flavor to your tests!

A collection of small, miscellaneous plugins for [Testo framework](https://github.com/ozontech/testo).

## Plugins

- [Rerun](./rerun) - adds `--last-failed`-like behaviour from Pytest to Testo. Makes it possible to rerun only failed tests.
- [XFail](./xfail) - adds `t.XFail()` method to mark a test as "expected to fail".
- [Parallel](./parallel) - marks all tests as parallel by default.
- [Environment](./environment) - loads environment variables from `.env` files for testing.

See also [Allure plugin for Testo](https://github.com/ozontech/testo-allure) - visualize results of a test run with [Allure Report](https://allurereport.org/).
