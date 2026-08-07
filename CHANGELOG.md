# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Fixed

- With `-rerun.failed`, a suite that failed only in its `BeforeAll`/`AfterAll` hooks
  (without failed tests) is now re-run instead of being skipped: the rerun plugin
  looked it up in the cache under a wrong key.

## [1.2.0] - 2026-06-15

### Added

- Plugin `async` for test-aware `sync.WaitGroup`.

## [1.1.1] - 2026-05-29

### Fixed

- Fixed a bug when long test names in rerun plugin could cause invalid reruns.

## [1.1.0] - 2026-05-24

### Added

- Introduce "scopes" for `parallel` plugin.

## [1.0.0] - 2026-05-13

### Added

- Initial stable version.

[1.2.0]: https://github.com/ozontech/testo/compare/v1.1.1...v1.2.0
[1.1.1]: https://github.com/ozontech/testo/compare/v1.1.0...v1.1.1
[1.1.0]: https://github.com/ozontech/testo/compare/v1.0.0...v1.1.0
[1.0.0]: https://github.com/ozontech/testo-toppings/releases/tag/v1.0.0
