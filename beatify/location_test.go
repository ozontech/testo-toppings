package beatify

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestShortPathUsesRelative(t *testing.T) {
	t.Parallel()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}

	got := shortPath(file)
	if strings.Contains(got, "..") {
		t.Fatalf("shortPath(%q) = %q, want path without ..", file, got)
	}
	if !strings.HasSuffix(got, "location_test.go") && got != "location_test.go" {
		t.Fatalf("shortPath(%q) = %q, want *location_test.go", file, got)
	}
}

func TestFormatLocation(t *testing.T) {
	t.Parallel()

	t.Run("empty", func(t *testing.T) {
		t.Parallel()

		got := formatLocation(painter{enabled: false}, sourceLoc{})
		if got != "-" {
			t.Fatalf("got %q, want -", got)
		}
	})

	t.Run("plain absolute when colors off", func(t *testing.T) {
		t.Parallel()

		loc := sourceLoc{File: "/tmp/example/main_test.go", Line: 42}
		got := formatLocation(painter{enabled: false}, loc)
		want := "/tmp/example/main_test.go:42"
		if got != want {
			t.Fatalf("got %q, want %q", got, want)
		}
	})

	t.Run("osc8 when colors on", func(t *testing.T) {
		t.Parallel()

		loc := sourceLoc{File: "/tmp/example/main_test.go", Line: 42}
		got := formatLocation(painter{enabled: true}, loc)
		if !strings.Contains(got, "vscode://file") {
			t.Fatalf("expected vscode:// URI in %q", got)
		}
		if !strings.Contains(got, "\x1b]8;;") {
			t.Fatalf("expected OSC 8 wrapper in %q", got)
		}
		if !strings.Contains(got, ":42") {
			t.Fatalf("expected :42 in %q", got)
		}
	})
}

func TestEditorURI(t *testing.T) {
	t.Parallel()

	t.Run("unix", func(t *testing.T) {
		t.Parallel()

		got := editorURI("/tmp/foo_test.go", 10)
		want := "vscode://file/tmp/foo_test.go:10"
		if got != want {
			t.Fatalf("got %q, want %q", got, want)
		}
	})

	t.Run("windows drive", func(t *testing.T) {
		t.Parallel()

		got := editorURI("C:/Users/x/main_test.go", 25)
		want := "vscode://file/C:/Users/x/main_test.go:25"
		if got != want {
			t.Fatalf("got %q, want %q", got, want)
		}
	})
}

func TestNormalizePath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in, want string
	}{
		{in: "/mnt/c/Users/x/main_test.go", want: "C:/Users/x/main_test.go"},
		{in: "/mnt/d/work/a.go", want: "D:/work/a.go"},
		{in: "/home/rict/a.go", want: "/home/rict/a.go"},
		{in: `C:\Users\x\a.go`, want: `C:\Users\x\a.go`},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			t.Parallel()

			if got := normalizePath(tt.in); got != tt.want {
				t.Fatalf("normalizePath(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

type locSuite struct{}

func (locSuite) TestSample() {}

func TestLocFromSuiteMethod(t *testing.T) {
	t.Parallel()

	loc := locFromSuiteMethod(locSuite{}, "TestSample")
	if loc.File == "" {
		t.Fatal("expected source file")
	}
	if filepath.Base(loc.File) != "location_test.go" {
		t.Fatalf("file=%q", loc.File)
	}
	if loc.Line <= 0 {
		t.Fatalf("line=%d", loc.Line)
	}
}
