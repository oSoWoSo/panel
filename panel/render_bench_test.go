package panel

import (
	"fmt"
	"testing"

	"charm.land/lipgloss/v2"
)

func benchItems(n int) []string {
	items := make([]string, n)
	for i := range items {
		items[i] = fmt.Sprintf("pkg-%03d", i+1)
	}
	return items
}

// BenchmarkRenderFilterReverse measures one frame of a filter panel that shows
// a long list in reverse mode - how vote filters the package list.
func BenchmarkRenderFilterReverse(b *testing.B) {
	fm := newMockFilterModel(benchItems(555), 12, false)
	fm.reverse = true
	m := orchestrator{filterModels: []filterModel{*fm}, activeIdx: 0, height: 12}
	b.ResetTimer()
	for b.Loop() {
		_ = m.renderFilterView(fm, true)
	}
}

// BenchmarkRenderFilterForward is the same list without --reverse.
func BenchmarkRenderFilterForward(b *testing.B) {
	fm := newMockFilterModel(benchItems(555), 12, false)
	m := orchestrator{filterModels: []filterModel{*fm}, activeIdx: 0, height: 12}
	b.ResetTimer()
	for b.Loop() {
		_ = m.renderFilterView(fm, true)
	}
}

// BenchmarkViewReverse measures a whole frame of the vote case: one filter
// panel with a long list, border included.
func BenchmarkViewReverse(b *testing.B) {
	fm := newMockFilterModel(benchItems(555), 12, false)
	fm.reverse = true
	m := orchestrator{
		panels:       []Panel{{Type: PanelFilter, ModelIdx: 0}},
		filterModels: []filterModel{*fm},
		activeIdx:    0,
		height:       12,
		gap:          0,
		reverse:      true,
		borderStyle:  lipgloss.NewStyle(),
	}
	b.ResetTimer()
	for b.Loop() {
		_ = m.View()
	}
}

// BenchmarkViewReverseBig is a 5555 item list - a whole masterdir.
func BenchmarkViewReverseBig(b *testing.B) {
	fm := newMockFilterModel(benchItems(5555), 12, false)
	fm.reverse = true
	m := orchestrator{
		panels:       []Panel{{Type: PanelFilter, ModelIdx: 0}},
		filterModels: []filterModel{*fm},
		activeIdx:    0,
		height:       12,
		gap:          0,
		reverse:      true,
		borderStyle:  lipgloss.NewStyle(),
	}
	b.ResetTimer()
	for b.Loop() {
		_ = m.View()
	}
}
