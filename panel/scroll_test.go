package panel

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func scrollItems(n int) []string {
	items := make([]string, n)
	for i := range items {
		items[i] = fmt.Sprintf("pkg-%02d", i+1)
	}
	return items
}

// A list longer than the panel has to scroll: the cursor must stay visible,
// otherwise the user moves the selection into rows nobody can see.
func TestFilterScrollsCursorIntoView(t *testing.T) {
	fm := newMockFilterModel(scrollItems(30), 10, false)
	m := orchestrator{filterModels: []filterModel{*fm}}

	for range 25 {
		m.updateFilterModel(fm, tea.KeyPressMsg{Code: tea.KeyDown})
	}

	if fm.cursor != 25 {
		t.Fatalf("cursor = %d, want 25", fm.cursor)
	}
	if fm.offset == 0 {
		t.Error("the window did not follow the cursor")
	}

	view := stripANSI(m.renderFilterView(fm, true))
	if !strings.Contains(view, "pkg-26") {
		t.Errorf("the cursor row is not visible:\n%s", view)
	}
	if strings.Contains(view, "pkg-01") {
		t.Errorf("the window still shows items far above the cursor:\n%s", view)
	}

	// every visible row is rendered, no more: the rows are what costs time
	rows := 0
	for _, line := range strings.Split(strings.TrimRight(view, "\n"), "\n") {
		if strings.Contains(line, "pkg-") {
			rows++
		}
	}
	if rows != 10 {
		t.Errorf("rendered %d rows, want 10", rows)
	}
}

// --reverse puts the prompt at the bottom and prints the window bottom up.
func TestFilterReverseScrollsCursorIntoView(t *testing.T) {
	fm := newMockFilterModel(scrollItems(30), 10, false)
	fm.reverse = true
	m := orchestrator{filterModels: []filterModel{*fm}}

	m.updateFilterModel(fm, tea.KeyPressMsg{Code: tea.KeyDown})
	m.updateFilterModel(fm, tea.KeyPressMsg{Code: tea.KeyDown})

	view := stripANSI(m.renderFilterView(fm, true))
	if !strings.Contains(view, "pkg-03") {
		t.Errorf("the cursor row is not visible:\n%s", view)
	}
	// the prompt stays at the bottom
	if !strings.HasSuffix(view, "> ") && !strings.Contains(strings.Split(view, "\n")[len(strings.Split(view, "\n"))-1], ">") {
		t.Errorf("the prompt is not the last line:\n%s", view)
	}
}

func TestFilterReverseRendersOnlyTheWindow(t *testing.T) {
	fm := newMockFilterModel(scrollItems(5555), 12, false)
	fm.reverse = true
	m := orchestrator{filterModels: []filterModel{*fm}}

	view := m.renderFilterView(fm, true)

	rows := strings.Count(view, "\n")
	if rows > 12 {
		t.Errorf("rendered %d lines for a 12 row panel", rows)
	}
	if !strings.Contains(view, "pkg-0001") && !strings.Contains(view, "pkg-01") {
		t.Errorf("the first rows of the window are missing:\n%s", view)
	}
}

// The cursor of a choose panel belongs to the item, so it is visible on every
// page, not only on the first one.
func TestChooseCursorVisibleOnSecondPage(t *testing.T) {
	cm := newMockChooseModel(scrollItems(20), 10)
	o := orchestrator{chooseModels: []chooseModel{*cm}}

	for range 11 {
		o.updateChooseModel(cm, tea.KeyPressMsg{Code: tea.KeyDown})
	}

	view := stripANSI(o.renderChooseView(cm, true))
	if !strings.Contains(view, cm.cursor+cm.cursorPrefix+"pkg-12") {
		t.Errorf("cursor row missing on the second page:\n%s", view)
	}
}

// pgup and pgdown move the cursor a page, like in a scrolling list.
func TestFilterPageKeysMoveCursorAPage(t *testing.T) {
	fm := newMockFilterModel(scrollItems(40), 10, false)
	m := orchestrator{filterModels: []filterModel{*fm}}

	m.updateFilterModel(fm, tea.KeyPressMsg{Code: tea.KeyPgDown})
	if fm.cursor != 10 {
		t.Errorf("cursor = %d, want 10", fm.cursor)
	}
	m.updateFilterModel(fm, tea.KeyPressMsg{Code: tea.KeyPgUp})
	if fm.cursor != 0 {
		t.Errorf("cursor = %d, want 0", fm.cursor)
	}
}
