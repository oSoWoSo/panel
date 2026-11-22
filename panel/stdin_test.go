package panel

import (
	"strconv"
	"strings"
	"testing"
)

func TestStdinItems(t *testing.T) {
	t.Run("one item per line", func(t *testing.T) {
		got := stdinItems(Options{FromStdin: true}, strings.NewReader("a\nb\nc\n"))
		want := []string{"a", "b", "c"}
		if len(got) != len(want) {
			t.Fatalf("got %v, want %v", got, want)
		}
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("got %v, want %v", got, want)
			}
		}
	})

	t.Run("blank lines and crlf are dropped", func(t *testing.T) {
		got := stdinItems(Options{FromStdin: true}, strings.NewReader("a\r\n\n \nb\n\n"))
		if len(got) != 2 || got[0] != "a" || got[1] != "b" {
			t.Fatalf("got %v, want [a b]", got)
		}
	})

	t.Run("ansi is stripped when asked for", func(t *testing.T) {
		got := stdinItems(Options{FromStdin: true, StripANSI: true}, strings.NewReader("\x1b[31mred\x1b[0m\n"))
		if len(got) != 1 || got[0] != "red" {
			t.Fatalf("got %v, want [red]", got)
		}
	})

	t.Run("ansi kept when strip is off", func(t *testing.T) {
		got := stdinItems(Options{FromStdin: true, StripANSI: false}, strings.NewReader("\x1b[31mred\x1b[0m\n"))
		if len(got) != 1 || got[0] != "\x1b[31mred\x1b[0m" {
			t.Fatalf("got %q, want the escape sequence kept", got[0])
		}
	})

	t.Run("nothing is read without the flag", func(t *testing.T) {
		if got := stdinItems(Options{}, strings.NewReader("a\nb\n")); got != nil {
			t.Fatalf("got %v, want nil", got)
		}
	})
}

// A panel without items on the command line takes them from the pipe, the
// items given on the command line still win.
func TestPanelsFromStdinItems(t *testing.T) {
	t.Run("panel without items takes the piped ones", func(t *testing.T) {
		panels, err := parsePanelsFromArgs([]string{"filter", "--no-strict", "--header", "T"}, []string{"a", "b"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(panels) != 1 || panels[0].FilterOpts == nil {
			t.Fatalf("got %+v", panels)
		}
		if got := panels[0].FilterOpts.Options; len(got) != 2 || got[0] != "a" || got[1] != "b" {
			t.Fatalf("items = %v, want [a b]", got)
		}
	})

	t.Run("items on the command line win", func(t *testing.T) {
		panels, err := parsePanelsFromArgs([]string{"choose", "x", "y"}, []string{"a", "b"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got := panels[0].ChooseOpts.Options; len(got) != 2 || got[0] != "x" {
			t.Fatalf("items = %v, want [x y]", got)
		}
	})

	t.Run("without piped items a bare panel is an error", func(t *testing.T) {
		if _, err := parsePanelsFromArgs([]string{"choose"}, nil); err == nil {
			t.Fatal("expected an error when there are no items at all")
		}
	})
}

// The items must not cost time proportional to their number squared: passing
// them on the command line made kong copy its remaining token slice for every
// single one of them. Run with -bench to watch the numbers.
func BenchmarkPanelsFromStdin(b *testing.B) {
	args := []string{"filter", "--no-strict", "--header", "T"}
	for _, n := range []int{100, 1000, 15000} {
		b.Run(strconv.Itoa(n), func(b *testing.B) {
			for b.Loop() {
				if _, err := parsePanelsFromArgs(args, make([]string, n)); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
