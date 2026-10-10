package beatify

import (
	"fmt"
	"io"
	"strings"
)

// ANSI SGR codes.
const (
	ansiReset  = "\x1b[0m"
	ansiBold   = "\x1b[1m"
	ansiDim    = "\x1b[2m"
	ansiRed    = "\x1b[31m"
	ansiGreen  = "\x1b[32m"
	ansiYellow = "\x1b[33m"
	ansiOrange = "\x1b[38;5;208m"
	ansiCyan   = "\x1b[36m"
	ansiWhite  = "\x1b[37m"
)

type painter struct {
	enabled bool
}

func newPainter(enabled bool) painter {
	return painter{enabled: enabled}
}

func (p painter) wrap(codes, s string) string {
	if !p.enabled || s == "" {
		return s
	}

	return codes + s + ansiReset
}

func (p painter) bold(s string) string   { return p.wrap(ansiBold, s) }
func (p painter) dim(s string) string    { return p.wrap(ansiDim, s) }
func (p painter) red(s string) string    { return p.wrap(ansiRed, s) }
func (p painter) green(s string) string  { return p.wrap(ansiGreen, s) }
func (p painter) yellow(s string) string { return p.wrap(ansiYellow, s) }
func (p painter) orange(s string) string { return p.wrap(ansiOrange, s) }
func (p painter) cyan(s string) string   { return p.wrap(ansiCyan, s) }
func (p painter) white(s string) string  { return p.wrap(ansiWhite, s) }

func (p painter) boldCyan(s string) string   { return p.wrap(ansiBold+ansiCyan, s) }
func (p painter) boldGreen(s string) string  { return p.wrap(ansiBold+ansiGreen, s) }
func (p painter) boldRed(s string) string    { return p.wrap(ansiBold+ansiRed, s) }
func (p painter) boldYellow(s string) string { return p.wrap(ansiBold+ansiYellow, s) }

func writef(w io.Writer, format string, args ...any) {
	_, _ = fmt.Fprintf(w, format, args...)
}

func writeln(w io.Writer, s string) {
	_, _ = fmt.Fprintln(w, s)
}

func repeat(char string, n int) string {
	if n <= 0 {
		return ""
	}

	return strings.Repeat(char, n)
}
