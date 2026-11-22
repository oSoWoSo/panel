package panel

import (
	"bytes"
	"strings"
	"testing"

	"github.com/oSoWoSo/panel/v2/input"
	"github.com/oSoWoSo/panel/v2/internal/exit"
	"github.com/oSoWoSo/panel/v2/write"
)

func TestRunMan(t *testing.T) {
	parser, err := newParser("panel version test", func(int) {})
	if err != nil {
		t.Fatalf("newParser: %v", err)
	}
	var out bytes.Buffer
	parser.Stdout = &out

	if code := runMan(parser, nil); code != 0 {
		t.Fatalf("runMan() = %d, want 0", code)
	}
	for _, want := range []string{
		`.TH PANEL 1`,
		`panel - Glamorous panels for your shell scripts\&.`,
		".SH COPYRIGHT",
		"Charmbracelet, Inc\\&.",
	} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("man page missing %q\n%s", want, out.String())
		}
	}
	// The rendered page must describe the panel command line, not itself.
	if strings.Contains(out.String(), "\fBman\fP") {
		t.Error("man page lists a man subcommand, want only the root command")
	}
}

func TestRunManRejectsArguments(t *testing.T) {
	parser, err := newParser("panel version test", func(int) {})
	if err != nil {
		t.Fatalf("newParser: %v", err)
	}
	var out bytes.Buffer
	parser.Stdout = &out

	if code := runMan(parser, []string{"extra"}); code != 1 {
		t.Fatalf("runMan(parser, [extra]) = %d, want 1", code)
	}
}

func TestRunCompletionGenerators(t *testing.T) {
	parser, err := newParser("panel version test", func(int) {})
	if err != nil {
		t.Fatalf("newParser: %v", err)
	}
	for _, tc := range []struct {
		sub  string
		want string
	}{
		{"bash", `has_completion_function="yes"`},
		{"zsh", "#compdef"},
		{"fish", "complete "},
	} {
		t.Run(tc.sub, func(t *testing.T) {
			var out bytes.Buffer
			parser.Stdout = &out
			if code := runCompletion(parser, []string{tc.sub}); code != 0 {
				t.Fatalf("runCompletion(%q) = %d, want 0", tc.sub, code)
			}
			if !strings.Contains(out.String(), tc.want) {
				t.Errorf("%s script missing %q", tc.sub, tc.want)
			}
		})
	}
}

func TestRunCompletionRejectsUnknown(t *testing.T) {
	parser, err := newParser("panel version test", func(int) {})
	if err != nil {
		t.Fatalf("newParser: %v", err)
	}
	for _, args := range [][]string{nil, {"nope"}} {
		if code := runCompletion(parser, args); code != 1 {
			t.Errorf("runCompletion(%v) = %d, want 1", args, code)
		}
	}
}

func TestRunPromptHelp(t *testing.T) {
	for _, tc := range []struct {
		name        string
		description string
		model       any
		want        []string
	}{
		{"input", "Prompt for some input", &input.Options{}, []string{"Usage: panel input", "Prompt for some input", "--placeholder", "--char-limit"}},
		{"write", "Prompt for long-form text", &write.Options{}, []string{"Usage: panel write", "Prompt for long-form text", "--show-line-numbers"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var out bytes.Buffer
			if code := runPromptSubcommand("panel version test", tc.name, tc.description, tc.model, []string{"--help"}, &out); code != 0 {
				t.Fatalf("runPromptSubcommand(%s, --help) = %d, want 0", tc.name, code)
			}
			for _, want := range tc.want {
				if !strings.Contains(out.String(), want) {
					t.Errorf("%s help missing %q", tc.name, want)
				}
			}
		})
	}
}

func TestRunPromptRejectsUnknown(t *testing.T) {
	for _, args := range [][]string{nil, {"nope"}} {
		if code := runPrompt("panel version test", args); code != 1 {
			t.Errorf("runPrompt(%v) = %d, want 1", args, code)
		}
	}
}

func TestExitCodeFromCanceled(t *testing.T) {
	// A canceled run must exit 1 and, unlike a plain error, print nothing:
	// the TUI is already torn down when Run() returns.
	if code := exitCodeFromError(exit.ErrExit(1)); code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
}
