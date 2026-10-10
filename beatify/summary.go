package beatify

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	statusPass status = iota
	statusFail
	statusSkip
)

type status int

type testResult struct {
	Name     string
	Status   status
	Duration time.Duration
	Loc      sourceLoc
}

type runStats struct {
	mu sync.Mutex

	suiteStart time.Time
	results    []testResult
}

type statusCounts struct {
	total   int
	passed  int
	failed  int
	skipped int
}

func (s *runStats) markStart(now time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.suiteStart.IsZero() {
		s.suiteStart = now
	}
}

func (s *runStats) add(r testResult) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.results = append(s.results, r)
}

func (s *runStats) snapshot() (time.Time, []testResult) {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make([]testResult, len(s.results))
	copy(out, s.results)

	return s.suiteStart, out
}

func (p *PluginBeatify) printSummary() {
	p.Helper()

	start, results := p.stats.snapshot()
	paint := p.painter
	w := os.Stdout
	counts := countStatuses(results)
	pct := resultsCoverage(counts)

	if len(results) == 0 {
		writeln(w, "  "+coverageLine(paint, pct))

		return
	}

	elapsed := suiteElapsed(start, results)

	const rule = "─"

	nameWidth := maxNameWidth(results, 36)
	width := max(52, summaryWidth(results, nameWidth))
	title := " Test Summary "
	side := max((width-len(title))/2, 2)
	header := repeat(rule, side) + title + repeat(rule, width-side-len(title))

	writeln(w, "")
	writeln(w, paint.boldCyan(header))

	for _, row := range formatResultRows(paint, results, nameWidth) {
		writeln(w, row)
	}

	writeln(w, paint.dim(repeat(rule, width)))
	writeln(w, "  "+progressBar(paint, counts, 24))
	writeln(w, "  "+formatTotals(paint, counts, elapsed))
	writeln(w, "  "+coverageLine(paint, pct))
	writeln(w, "  "+formatVerdict(paint, counts))
	writeln(w, paint.boldCyan(repeat(rule, width)))
	writeln(w, "")
}

func (p *PluginBeatify) printCoverageOnly() {
	p.Helper()

	_, results := p.stats.snapshot()
	pct := resultsCoverage(countStatuses(results))
	writeRealStdout(coverageLine(p.painter, pct))
}

func countStatuses(results []testResult) statusCounts {
	var c statusCounts

	c.total = len(results)
	for _, r := range results {
		switch r.Status {
		case statusPass:
			c.passed++
		case statusFail:
			c.failed++
		case statusSkip:
			c.skipped++
		}
	}

	return c
}

func suiteElapsed(start time.Time, results []testResult) time.Duration {
	if !start.IsZero() {
		return time.Since(start)
	}

	var elapsed time.Duration

	for _, r := range results {
		elapsed += r.Duration
	}

	return elapsed
}

func formatResultRows(paint painter, results []testResult, nameWidth int) []string {
	rows := make([]string, 0, len(results))

	for _, r := range results {
		icon, coloredStatus := statusGlyph(paint, r.Status)
		name := padRight(truncate(displayName(r.Name), nameWidth), nameWidth)
		dur := paint.dim(fmt.Sprintf("%8s", formatDuration(r.Duration)))
		loc := formatLocation(paint, r.Loc)

		rows = append(
			rows,
			fmt.Sprintf("  %s  %s  %s  %s  %s", icon, coloredStatus, name, dur, loc),
		)
	}

	return rows
}

func summaryWidth(results []testResult, nameWidth int) int {
	width := 0

	for _, r := range results {
		loc := "-"
		if r.Loc.File != "" && r.Loc.Line > 0 {
			loc = fmt.Sprintf("%s:%d", shortPath(normalizePath(r.Loc.File)), r.Loc.Line)
		}

		// "  X  STATUS  " + name + "  " + duration(8) + "  " + loc
		n := 2 + 1 + 2 + 4 + 2 + nameWidth + 2 + 8 + 2 + len(loc)
		if n > width {
			width = n
		}
	}

	return width
}

func maxNameWidth(results []testResult, limit int) int {
	nameWidth := 0
	for _, r := range results {
		if n := len(displayName(r.Name)); n > nameWidth {
			nameWidth = n
		}
	}

	return min(nameWidth, limit)
}

func formatTotals(paint painter, c statusCounts, elapsed time.Duration) string {
	parts := []string{
		paint.bold(fmt.Sprintf("%d tests", c.total)),
		paint.green(fmt.Sprintf("%d passed", c.passed)),
	}

	if c.failed > 0 {
		parts = append(parts, paint.red(fmt.Sprintf("%d failed", c.failed)))
	} else {
		parts = append(parts, paint.dim("0 failed"))
	}

	if c.skipped > 0 {
		parts = append(parts, paint.yellow(fmt.Sprintf("%d skipped", c.skipped)))
	} else {
		parts = append(parts, paint.dim("0 skipped"))
	}

	parts = append(parts, paint.cyan(formatDuration(elapsed)))

	return strings.Join(parts, paint.dim(" · "))
}

func formatVerdict(paint painter, c statusCounts) string {
	switch {
	case c.failed > 0:
		return paint.boldRed("✗ FAILED")
	case c.skipped > 0 && c.passed == 0:
		return paint.boldYellow("↷ ALL SKIPPED")
	case c.skipped > 0:
		return paint.boldGreen("✓ PASSED") + paint.dim(" (with skips)")
	default:
		return paint.boldGreen("✓ ALL PASSED")
	}
}

func (p *PluginBeatify) printTestLine(r testResult) {
	p.Helper()

	if p.quiet {
		return
	}

	paint := p.painter
	icon, coloredStatus := statusGlyph(paint, r.Status)
	name := paint.white(displayName(r.Name))
	dur := paint.dim(formatDuration(r.Duration))

	writef(os.Stdout, "  %s %s %s %s\n", icon, coloredStatus, name, dur)
}

func statusGlyph(paint painter, s status) (icon, label string) {
	switch s {
	case statusPass:
		return paint.green("✓"), paint.green("PASS")
	case statusFail:
		return paint.red("✗"), paint.red("FAIL")
	case statusSkip:
		return paint.yellow("↷"), paint.yellow("SKIP")
	default:
		return "?", "????"
	}
}

func progressBar(paint painter, c statusCounts, width int) string {
	if c.total == 0 || width <= 0 {
		return ""
	}

	passW := c.passed * width / c.total
	failW := c.failed * width / c.total
	skipW := c.skipped * width / c.total
	rest := max(width-passW-failW-skipW, 0)

	var b strings.Builder

	b.WriteString(paint.dim("["))
	b.WriteString(paint.green(repeat("█", passW)))
	b.WriteString(paint.red(repeat("█", failW)))
	b.WriteString(paint.yellow(repeat("█", skipW)))
	b.WriteString(paint.dim(repeat("░", rest)))
	b.WriteString(paint.dim("]"))

	return b.String()
}

func displayName(full string) string {
	parts := strings.Split(full, "/")
	filtered := make([]string, 0, len(parts))

	for _, part := range parts {
		if part == "testo!" || strings.HasPrefix(part, "#") {
			continue
		}

		filtered = append(filtered, part)
	}

	if len(filtered) == 0 {
		return full
	}

	return filtered[len(filtered)-1]
}

func formatDuration(d time.Duration) string {
	switch {
	case d < time.Millisecond:
		return fmt.Sprintf("%dµs", d.Microseconds())
	case d < time.Second:
		return fmt.Sprintf("%dms", d.Milliseconds())
	default:
		return fmt.Sprintf("%.2fs", d.Seconds())
	}
}

func truncate(s string, maxLen int) string {
	if maxLen <= 0 || len(s) <= maxLen {
		return s
	}

	if maxLen <= 3 {
		return s[:maxLen]
	}

	return s[:maxLen-3] + "..."
}

func padRight(s string, width int) string {
	if len(s) >= width {
		return s
	}

	return s + strings.Repeat(" ", width-len(s))
}
