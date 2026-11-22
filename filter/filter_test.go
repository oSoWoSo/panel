package filter

import (
	"fmt"
	"reflect"
	"testing"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

func TestPasteRefreshesMatches(t *testing.T) {
	for _, reverse := range []bool{false, true} {
		for _, fuzzy := range []bool{false, true} {
			t.Run(fmt.Sprintf("reverse=%t/fuzzy=%t", reverse, fuzzy), func(t *testing.T) {
				input := textinput.New()
				input.Focus()
				choices := []string{"existing-template"}
				v := viewport.New(viewport.WithWidth(80), viewport.WithHeight(10))
				m := model{
					textinput:        input,
					viewport:         &v,
					keymap:           defaultKeymap(),
					filteringChoices: choices,
					matches:          matchAll(choices),
					limit:            1,
					reverse:          reverse,
					fuzzy:            fuzzy,
				}
				const url = "https://github.com/example/new-template"
				updated, _ := m.Update(tea.PasteMsg{Content: url})
				got := updated.(model)
				if got.textinput.Value() != url {
					t.Fatalf("input = %q, want %q", got.textinput.Value(), url)
				}
				if len(got.matches) != 1 || got.matches[0].Str != url {
					t.Fatalf("matches after paste = %v, want pasted URL", got.matches)
				}
				updated, _ = got.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
				got = updated.(model)
				if !got.submitted || got.matches[got.cursor].Str != url {
					t.Fatal("Enter should submit the pasted URL without another edit")
				}
			})
		}
	}
}

func TestMatchedRanges(t *testing.T) {
	for name, tt := range map[string]struct {
		in  []int
		out [][2]int
	}{
		"empty": {
			in:  []int{},
			out: [][2]int{},
		},
		"one char": {
			in:  []int{1},
			out: [][2]int{{1, 1}},
		},
		"2 char range": {
			in:  []int{1, 2},
			out: [][2]int{{1, 2}},
		},
		"multiple char range": {
			in:  []int{1, 2, 3, 4, 5, 6},
			out: [][2]int{{1, 6}},
		},
		"multiple char ranges": {
			in:  []int{1, 2, 3, 5, 6, 10, 11, 12, 13, 23, 24, 40, 42, 43, 45, 52},
			out: [][2]int{{1, 3}, {5, 6}, {10, 13}, {23, 24}, {40, 40}, {42, 43}, {45, 45}, {52, 52}},
		},
	} {
		t.Run(name, func(t *testing.T) {
			match := matchedRanges(tt.in)
			if !reflect.DeepEqual(match, tt.out) {
				t.Errorf("expected %v, got %v", tt.out, match)
			}
		})
	}
}

func TestByteToChar(t *testing.T) {
	stStr := "\x1b[90m\ue615\x1b[39m \x1b[3m\x1b[32mDow\x1b[0m\x1b[90m\x1b[39m\x1b[3wnloads"
	str := " Downloads"
	rng := [2]int{4, 7}
	expect := "Dow"

	if got := str[rng[0]:rng[1]]; got != expect {
		t.Errorf("expected %q, got %q", expect, got)
	}

	start, stop := bytePosToVisibleCharPos(str, rng)
	if got := ansi.Strip(ansi.Cut(stStr, start, stop)); got != expect {
		t.Errorf("expected %+q, got %+q", expect, got)
	}
}

func TestPasteUpdatesMatches(t *testing.T) {
	i := textinput.New()
	i.Focus()
	v := viewport.New(viewport.WithWidth(40), viewport.WithHeight(10))

	filteringChoices := []string{"alpha", "beta", "foo-bar"}
	m := model{
		choices: map[string]string{
			"alpha":   "alpha",
			"beta":    "beta",
			"foo-bar": "foo-bar",
		},
		filteringChoices: filteringChoices,
		matches:          matchAll(filteringChoices),
		textinput:        i,
		viewport:         &v,
		selected:         map[string]struct{}{},
		limit:            1,
		fuzzy:            true,
		sort:             true,
		strict:           true,
		keymap:           defaultKeymap(),
		help:             help.New(),
	}

	// Pasted text arrives as tea.PasteMsg; the query must be filtered just
	// like when the same characters are typed one by one.
	tm, _ := m.Update(tea.PasteMsg{Content: "foo"})
	m = tm.(model)

	if got := m.textinput.Value(); got != "foo" {
		t.Errorf("textinput value = %q, want %q", got, "foo")
	}
	if len(m.matches) != 1 {
		t.Fatalf("matches = %d, want 1", len(m.matches))
	}
	if m.matches[0].Str != "foo-bar" {
		t.Errorf("match = %q, want %q", m.matches[0].Str, "foo-bar")
	}
}
