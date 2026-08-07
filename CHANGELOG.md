# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.3.0] - 2026-08-07

### Changed

- Update `testo` to v1.7.0.

### Fixed

- Rerun plugin looked up suites under a wrong cache key, so a suite that failed
  only in `BeforeAll`/`AfterAll` hooks was skipped instead of re-run.
- Rerun plugin read the cache in `init`, ignoring the `-cache.dir` and
  `-cache.disable` flags.
- Test and suite names that differ only by `-` vs `/` shared one cache entry,
  which could leave a failed test out of the rerun. Cache keys changed format,
  so the first `-rerun.failed` after the upgrade finds no previous failures.
- Runs filtered with `-test.run` erased previous failures. Now a suite is kept
  whenever it or any test under it failed, including in sub-suites.
- Packages sharing one `TESTO_CACHE_DIR` interfered with each other:
  boilerplate names like `Test/Suite` repeat across packages, so a failure
  in one could block another's skip. The cache is now namespaced per package.
- `-rerun.failed` with `-cache.disable` now warns and runs all tests instead
  of skipping everything as if nothing had failed.

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

[1.3.0]: https://github.com/ozontech/testo/compare/v1.2.0...v1.3.0
[1.2.0]: https://github.com/ozontech/testo/compare/v1.1.1...v1.2.0
[1.1.1]: https://github.com/ozontech/testo/compare/v1.1.0...v1.1.1
[1.1.0]: https://github.com/ozontech/testo/compare/v1.0.0...v1.1.0
[1.0.0]: https://github.com/ozontech/testo-toppings/releases/tag/v1.0.0
