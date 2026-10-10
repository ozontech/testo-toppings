package beatify

import "fmt"

func coverageLine(paint painter, pct float64) string {
	text := fmt.Sprintf("coverage: %.1f%%", pct)

	return colorCoverage(paint, pct, text)
}

func colorCoverage(paint painter, pct float64, text string) string {
	switch {
	case pct < 30:
		return paint.red(text)
	case pct <= 80:
		return paint.orange(text)
	default:
		return paint.green(text)
	}
}

// resultsCoverage is passed/total as a percentage (not go test -cover).
func resultsCoverage(c statusCounts) float64 {
	if c.total == 0 {
		return 0
	}

	return float64(c.passed) / float64(c.total) * 100
}

