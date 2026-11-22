package panel

import (
	"github.com/oSoWoSo/panel/v2/choose"
	"github.com/oSoWoSo/panel/v2/filter"
	"github.com/oSoWoSo/panel/v2/input"
)

// PanelType represents the type of panel (choose, filter or input).
//
//revive:disable:var-naming
type PanelType string

//revive:enable:var-naming

const (
	// PanelChoose is a choose-style panel where users select from a list.
	PanelChoose PanelType = "choose"
	// PanelFilter is a filter-style panel with fuzzy search.
	PanelFilter PanelType = "filter"
	// PanelInput is an editable text panel where users type a value.
	PanelInput PanelType = "input"
)

// Panel represents a single panel configuration.
type Panel struct {
	Type       PanelType
	ModelIdx   int
	ChooseOpts *choose.Options // set when Type == PanelChoose
	FilterOpts *filter.Options // set when Type == PanelFilter
	InputOpts  *input.Options  // set when Type == PanelInput
	// InputMultiline is set when Type == PanelInput and the block passed
	// --multiline, which makes the panel edit a textarea instead of a
	// single-line input.
	InputMultiline bool
}
