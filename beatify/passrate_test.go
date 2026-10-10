package beatify

import (
	"strings"
	"testing"
)

func TestCoverageOnlyFromArgs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args []string
		want bool
	}{
		{name: "absent", args: nil, want: false},
		{name: "plain", args: []string{"-beatify.coverage-only"}, want: true},
		{name: "true", args: []string{"-beatify.coverage-only=true"}, want: true},
		{name: "false", args: []string{"-beatify.coverage-only=false"}, want: false},
		{
			name: "among others",
			args: []string{"-count=1", "-beatify.coverage-only", "-v"},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := coverageOnlyFromArgs(tt.args); got != tt.want {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCoverageLine(t *testing.T) {
	t.Parallel()

	paint := painter{enabled: false}
	got := coverageLine(paint, 42.5)
	if got != "coverage: 42.5%" {
		t.Fatalf("got %q", got)
	}
}

func TestColorCoverage(t *testing.T) {
	t.Parallel()

	paint := painter{enabled: true}
	text := "coverage: 50.0%"

	tests := []struct {
		name string
		pct  float64
		code string
	}{
		{name: "red below 30", pct: 29.9, code: ansiRed},
		{name: "orange at 30", pct: 30, code: ansiOrange},
		{name: "orange at 80", pct: 80, code: ansiOrange},
		{name: "green above 80", pct: 80.1, code: ansiGreen},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := colorCoverage(paint, tt.pct, text)
			if !strings.Contains(got, tt.code) {
				t.Fatalf("pct=%v got %q, want ANSI %q", tt.pct, got, tt.code)
			}
		})
	}
}

func TestResultsCoverage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		c    statusCounts
		want float64
	}{
		{name: "empty", c: statusCounts{}, want: 0},
		{
			name: "example suite",
			c:    statusCounts{total: 10, passed: 6, failed: 2, skipped: 2},
			want: 60,
		},
		{
			name: "all passed",
			c:    statusCounts{total: 4, passed: 4},
			want: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := resultsCoverage(tt.c); got != tt.want {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}
