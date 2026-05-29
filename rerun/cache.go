package rerun

import (
	"encoding/json"
	"strings"

	"github.com/ozontech/testo/testocache"
)

const (
	keySep         = "-"
	keyPrefix      = "rerun" + keySep
	keyTestPrefix  = keyPrefix + "test" + keySep
	keySuitePrefix = keyPrefix + "suite" + keySep
)

func newCache() cache {
	return cache{
		Tests:  make(map[string]test),
		Suites: make(map[string]suite),
	}
}

func readCache() (cache, error) {
	if testocache.Disabled() {
		return newCache(), nil
	}

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
			var s suite

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
	Suites map[string]suite
}

// test is a cached test.
type test struct {
	Name   string `json:"n"`
	Failed bool   `json:"f"`
	Suite  string `json:"s"`
}

func (t test) Cache() error {
	marshalled, err := json.Marshal(t)
	if err != nil {
		return err
	}

	return testocache.Set(keyTestPrefix+normalize(t.Name), marshalled)
}

// suite is a cached suite.
type suite struct {
	Name   string `json:"n"`
	Failed bool   `json:"f"`
}

func (s suite) Cache() error {
	marshalled, err := json.Marshal(s)
	if err != nil {
		return err
	}

	return testocache.Set(keySuitePrefix+normalize(s.Name), marshalled)
}

func normalize(s string) string {
	return strings.ReplaceAll(s, "/", keySep)
}
