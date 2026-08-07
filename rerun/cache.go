package rerun

import (
	"encoding/json"
	"net/url"
	"strings"

	"github.com/ozontech/testo/testocache"
)

const (
	keySep = "-"
	// v2: names are URL-escaped. v1 replaced "/" with "-", so "a-b" and
	// "a/b" clobbered each other; the bump also hides stale v1 entries.
	keyPrefix      = "rerun" + keySep + "v2" + keySep
	keyTestPrefix  = keyPrefix + "test" + keySep
	keySuitePrefix = keyPrefix + "suite" + keySep
)

func newCache() cache {
	return cache{
		Tests:  make(map[string]test),
		Suites: make(map[string]suiteEntry),
	}
}

// readCache loads all cached entries. With caching disabled it returns
// testocache.ErrDisabled, so callers run everything instead of treating
// the empty cache as "nothing failed".
func readCache() (cache, error) {
	keys, err := testocache.Keys(keyPrefix + "*")
	if err != nil {
		return cache{}, err
	}

	c := newCache()

	for _, k := range keys {
		switch {
		case strings.HasPrefix(k, keyTestPrefix):
			var t test

			err = cacheGetJSON(k, &t)
			if err != nil {
				return cache{}, err
			}

			c.Tests[t.Name] = t

		case strings.HasPrefix(k, keySuitePrefix):
			var s suiteEntry

			err = cacheGetJSON(k, &s)
			if err != nil {
				return cache{}, err
			}

			c.Suites[s.Name] = s
		}
	}

	return c, nil
}

func cacheGetJSON(key string, v any) error {
	value, err := testocache.Get(key)
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
	// Key is suite name.
	Suites map[string]suiteEntry
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

	return testocache.Set(keyTestPrefix+normalize(t.Name), marshalled)
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

	return testocache.Set(keySuitePrefix+normalize(s.Name), marshalled)
}

// normalize escapes s without collisions and keeps "/" out of keys:
// testocache.Keys matches with path.Match, whose "*" does not cross "/".
func normalize(s string) string {
	return url.PathEscape(s)
}
