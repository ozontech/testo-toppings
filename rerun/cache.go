package rerun

import (
	"encoding/json"
	"net/url"
	"runtime/debug"
	"strings"

	"github.com/ozontech/testo/testocache"
)

const (
	keySep         = "-"
	keyTestPrefix  = "test" + keySep
	keySuitePrefix = "suite" + keySep
	keyPkgPrefix   = "pkg" + keySep
)

var currentPkg = packagePath()

// The cache is namespaced per package: boilerplate names like Test/Suite
// repeat across packages, and with a shared TESTO_CACHE_DIR their entries
// would mix. The namespace also hides entries of older plugin versions,
// which used lossy keys in the shared keyspace.
var cacheNS = testocache.Namespace("rerun" + keySep + currentPkg)

// globalNS is the cross-package failure registry: one entry per package,
// recording whether that package's own namespace contains any failure.
// It lets -rerun.failed distinguish "no failures anywhere" (run all)
// from "failures in another package" (skip). Disjoint from the
// per-package namespaces above and from legacy pre-1.3.0 flat keys.
var globalNS = testocache.Namespace("rerun")

func packagePath() string {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return ""
	}

	return strings.TrimSuffix(bi.Path, ".test")
}

func newCache() cache {
	return cache{
		Tests:  make(map[string]test),
		Suites: make(map[string]suiteEntry),
	}
}

// readCache loads all cached entries, own-package and cross-package.
// With caching disabled it returns testocache.ErrDisabled, so callers
// run everything instead of treating the empty cache as "nothing failed".
func readCache() (cache, error) {
	c, err := readOwnCache()
	if err != nil {
		return cache{}, err
	}

	keys, err := globalNS.Keys(keyPkgPrefix + "*")
	if err != nil {
		return cache{}, err
	}

	for _, k := range keys {
		var p pkgEntry

		err = getJSON(globalNS, k, &p)
		if err != nil {
			return cache{}, err
		}

		// Own entry is skipped: the own namespace just read is fresher.
		if p.Failed && p.Name != currentPkg {
			c.otherPkgFailed = true
		}
	}

	return c, nil
}

// readOwnCache loads the cached entries of this package's namespace.
func readOwnCache() (cache, error) {
	keys, err := cacheNS.Keys("*")
	if err != nil {
		return cache{}, err
	}

	c := newCache()

	for _, k := range keys {
		switch {
		case strings.HasPrefix(k, keyTestPrefix):
			var t test

			err = getJSON(cacheNS, k, &t)
			if err != nil {
				return cache{}, err
			}

			c.Tests[t.Name] = t

		case strings.HasPrefix(k, keySuitePrefix):
			var s suiteEntry

			err = getJSON(cacheNS, k, &s)
			if err != nil {
				return cache{}, err
			}

			c.Suites[s.Name] = s
		}
	}

	return c, nil
}

func getJSON(ns testocache.Cache, key string, v any) error {
	value, err := ns.Get(key)
	if err != nil {
		return err
	}

	return json.Unmarshal(value, v)
}

type cache struct {
	// Tests holds data about cached tests.
	// Key is full test name, as returned by t.Name().
	Tests map[string]test

	// Suites holds data about cached suites.
	// Key is the suite key, see suiteKey.
	Suites map[string]suiteEntry

	// otherPkgFailed reports whether another package's registry entry
	// records a failure, see globalNS.
	otherPkgFailed bool
}

// anyFailure reports whether this package's namespace has any failure.
func (c cache) anyFailure() bool {
	for _, t := range c.Tests {
		if t.Failed {
			return true
		}
	}

	for _, s := range c.Suites {
		if s.Failed {
			return true
		}
	}

	return false
}

// anyKnownFailure reports whether any failure is known in any package.
func (c cache) anyKnownFailure() bool {
	return c.anyFailure() || c.otherPkgFailed
}

// test is a cached test.
type test struct {
	Name   string `json:"n"`
	Failed bool   `json:"f"`
}

func (t test) Cache() error {
	marshalled, err := json.Marshal(t)
	if err != nil {
		return err
	}

	return cacheNS.Set(keyTestPrefix+normalize(t.Name), marshalled)
}

// suiteEntry is a cached suite.
type suiteEntry struct {
	Name   string `json:"n"`
	Failed bool   `json:"f"`
}

func (s suiteEntry) Cache() error {
	marshalled, err := json.Marshal(s)
	if err != nil {
		return err
	}

	return cacheNS.Set(keySuitePrefix+normalize(s.Name), marshalled)
}

// pkgEntry is a cached per-package failure status in globalNS.
type pkgEntry struct {
	Name   string `json:"n"`
	Failed bool   `json:"f"`
}

func (p pkgEntry) Cache() error {
	marshalled, err := json.Marshal(p)
	if err != nil {
		return err
	}

	return globalNS.Set(keyPkgPrefix+normalize(p.Name), marshalled)
}

// normalize escapes s without collisions and keeps "/" out of keys:
// testocache.Keys matches with path.Match, whose "*" does not cross "/".
func normalize(s string) string {
	return url.PathEscape(s)
}
