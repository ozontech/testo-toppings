package beatify

import "testing"

func TestDisplayName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "strips suite and testo wrapper",
			in:   "Test/Suite/testo!/TestAuthLogin",
			want: "TestAuthLogin",
		},
		{
			name: "strips run index",
			in:   "Test/#00/testo!/TestFoo",
			want: "TestFoo",
		},
		{
			name: "keeps plain name",
			in:   "PlainName",
			want: "PlainName",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := displayName(tt.in); got != tt.want {
				t.Fatalf("displayName(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestProgressBarLength(t *testing.T) {
	t.Parallel()

	paint := painter{enabled: false}
	counts := statusCounts{total: 10, passed: 6, failed: 2, skipped: 2}
	bar := progressBar(paint, counts, 24)

	// "[", 24 cells, "]"
	if got := len([]rune(bar)); got != 26 {
		t.Fatalf("progress bar rune length = %d, want 26 (%q)", got, bar)
	}
}

func TestTruncate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		in     string
		maxLen int
		want   string
	}{
		{name: "short", in: "abc", maxLen: 10, want: "abc"},
		{name: "exact", in: "abcdef", maxLen: 6, want: "abcdef"},
		{name: "long", in: "abcdefghij", maxLen: 6, want: "abc..."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := truncate(tt.in, tt.maxLen); got != tt.want {
				t.Fatalf("truncate(%q, %d) = %q, want %q", tt.in, tt.maxLen, got, tt.want)
			}
		})
	}
}
