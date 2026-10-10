package beatify

import (
	"fmt"
	"os"
	"strings"
	"sync"
)

var (
	silenceOnce sync.Once
	realStdout  *os.File
)

func init() {
	// Silence before testing prints === RUN; flags are not parsed yet.
	if coverageOnlyFromArgs(os.Args[1:]) {
		silenceTestOutput()
	}
}

func coverageOnlyFromArgs(args []string) bool {
	const name = "-beatify.coverage-only"

	for _, arg := range args {
		switch {
		case arg == name, arg == name+"=true":
			return true
		case arg == name+"=false":
			return false
		case strings.HasPrefix(arg, name+"="):
			return arg[len(name)+1:] == "true"
		}
	}

	return false
}

// silenceTestOutput redirects stdio to os.DevNull so coverage-only mode
// can print its line without testing package chatter.
func silenceTestOutput() {
	silenceOnce.Do(func() {
		saved, err := redirectStdio()
		if err != nil {
			return
		}

		realStdout = saved
	})
}

func writeRealStdout(line string) {
	w := realStdout
	if w == nil {
		w = os.Stdout
	}

	_, _ = fmt.Fprintln(w, line)
}
