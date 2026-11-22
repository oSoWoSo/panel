package choose

import (
	"time"

	"github.com/oSoWoSo/panel/v2/style"
)

// Options is the customization options for the choose command.
type Options struct {
	Options          []string      `arg:"" optional:"" help:"Options to choose from."`
	Limit            int           `help:"Maximum number of options to pick" default:"1" group:"Selection"`
	NoLimit          bool          `help:"Pick unlimited number of options (ignores limit)" group:"Selection"`
	Ordered          bool          `help:"Maintain the order of the selected options" env:"PANEL_CHOOSE_ORDERED"`
	Height           int           `help:"Height of the list" default:"10" env:"PANEL_CHOOSE_HEIGHT"`
	Cursor           string        `help:"Prefix to show on item that corresponds to the cursor position" default:"> " env:"PANEL_CHOOSE_CURSOR"`
	ShowHelp         bool          `help:"Show help keybinds" default:"true" negatable:"" env:"PANEL_CHOOSE_SHOW_HELP"`
	Timeout          time.Duration `help:"Timeout until choose returns selected element" default:"0s" env:"PANEL_CHOOSE_TIMEOUT"` // including timeout command options [Timeout,...]
	Header           string        `help:"Header value" default:"Choose:" env:"PANEL_CHOOSE_HEADER"`
	CursorPrefix     string        `help:"Prefix to show on the cursor item (hidden if limit is 1)" default:"• " env:"PANEL_CHOOSE_CURSOR_PREFIX"`
	SelectedPrefix   string        `help:"Prefix to show on selected items (hidden if limit is 1)" default:"✓ " env:"PANEL_CHOOSE_SELECTED_PREFIX"`
	UnselectedPrefix string        `help:"Prefix to show on unselected items (hidden if limit is 1)" default:"• " env:"PANEL_CHOOSE_UNSELECTED_PREFIX"`
	Selected         []string      `help:"Options that should start as selected (selects all if given *)" default:"" env:"PANEL_CHOOSE_SELECTED"`
	SelectIfOne      bool          `help:"Select the given option if there is only one" group:"Selection"`
	ExecHelp         string        `help:"Command to run on Ctrl+H (stdout printed above TUI)" default:"" env:"PANEL_CHOOSE_EXEC_HELP" group:"Selection"`
	InputDelimiter   string        `help:"Option delimiter when reading from STDIN" default:"\n" env:"PANEL_CHOOSE_INPUT_DELIMITER"`
	OutputDelimiter  string        `help:"Option delimiter when writing to STDOUT" default:"\n" env:"PANEL_CHOOSE_OUTPUT_DELIMITER"`
	LabelDelimiter   string        `help:"Allows to set a delimiter, so options can be set as label:value" default:"" env:"PANEL_CHOOSE_LABEL_DELIMITER"`
	StripANSI        bool          `help:"Strip ANSI sequences when reading from STDIN" default:"true" negatable:"" env:"PANEL_CHOOSE_STRIP_ANSI"`
	Padding          string        `help:"Padding" default:"${defaultPadding}" group:"Style Flags" env:"PANEL_CHOOSE_PADDING"`

	CursorStyle       style.Styles `embed:"" prefix:"cursor." set:"defaultForeground=212" envprefix:"PANEL_CHOOSE_CURSOR_"`
	HeaderStyle       style.Styles `embed:"" prefix:"header." set:"defaultForeground=99" envprefix:"PANEL_CHOOSE_HEADER_"`
	ItemStyle         style.Styles `embed:"" prefix:"item." hidden:"" envprefix:"PANEL_CHOOSE_ITEM_"`
	SelectedItemStyle style.Styles `embed:"" prefix:"selected." set:"defaultForeground=212" envprefix:"PANEL_CHOOSE_SELECTED_"`
}
