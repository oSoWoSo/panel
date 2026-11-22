package panel

import (
	"io"
	"os"
	"strings"

	"charm.land/bubbles/v2/help"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/oSoWoSo/panel/v2/internal/exit"
	"github.com/oSoWoSo/panel/v2/internal/timeout"
	"github.com/oSoWoSo/panel/v2/internal/tty"
)

// stdinItems reads the items of panels that got none on the command line from
// stdin, one item per line. Items on the command line are free for a handful
// of items, but kong copies its whole remaining token slice for every token it
// reads, so a list of ten thousand items costs seconds before the panel draws
// its first frame. A pipe keeps that cost linear.
func stdinItems(o Options, r io.Reader) []string {
	if !o.FromStdin || r == nil {
		return nil
	}

	data, err := io.ReadAll(r)
	if err != nil {
		return nil
	}

	var items []string
	for _, line := range strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n") {
		if o.StripANSI {
			line = ansi.Strip(line)
		}
		if line = strings.TrimSpace(line); line == "" {
			continue
		}
		items = append(items, line)
	}

	return items
}

// Run provides a shell script interface for running multiple panels
// side by side, allowing users to navigate between them.
func (o Options) Run() error {
	if len(o.Panel) == 0 {
		return os.ErrInvalid
	}

	panels, err := parsePanelsFromArgs(o.Panel, stdinItems(o, os.Stdin))
	if err != nil {
		return err
	}

	height := o.Height
	if height <= 0 {
		height = 10
	}

	if o.HeaderText != "" {
		if headerLines := lipgloss.Height(o.HeaderText); height > headerLines {
			height -= headerLines
		} else {
			height = 1
		}
		o.Height = height
	}

	m := orchestrator{
		panels:      panels,
		activeIdx:   o.Active,
		height:      height,
		borderStyle: o.BorderStyle(),
		showHelp:    o.ShowHelp,
		customHelp:  o.CustomHelp,
		gap:         o.Gap,
		vertical:    o.Vertical,
		reverse:     o.Reverse,
		headerText:  o.HeaderText,
		stacked:     o.Stacked,
		delimiter:   o.Delimiter,
		debug:       o.Debug,
		all:         o.All,
		single:      o.Single,

		// Border styles
		activeBorderStyle:   o.ActiveBorderStyle.ToLipgloss(),
		inactiveBorderStyle: o.InactiveBorderStyle.ToLipgloss(),

		// Common styles
		matchStyle:        o.MatchStyle.ToLipgloss(),
		cursorStyle:       o.CursorStyle.ToLipgloss(),
		headerStyle:       o.HeaderStyle.ToLipgloss(),
		itemStyle:         o.ItemStyle.ToLipgloss(),
		selectedItemStyle: o.SelectedItemStyle.ToLipgloss(),
		indicatorStyle:    o.IndicatorStyle.ToLipgloss(),

		// Choose-specific styles
		selectedPrefixStyle:   o.SelectedPrefixStyle.ToLipgloss(),
		unselectedPrefixStyle: o.UnselectedPrefixStyle.ToLipgloss(),

		// Filter-specific styles
		textStyle:        o.TextStyle.ToLipgloss(),
		cursorTextStyle:  o.CursorTextStyle.ToLipgloss(),
		promptStyle:      o.PromptStyle.ToLipgloss(),
		placeholderStyle: o.PlaceholderStyle.ToLipgloss(),

		indicator: o.Indicator,
		keymap:    defaultPanelKeymap(),
		help:      help.New(),
	}

	if initErr := m.initModels(o); initErr != nil {
		return initErr
	}

	ctx, cancel := timeout.Context(o.Timeout)
	defer cancel()

	programOpts := []tea.ProgramOption{
		tea.WithOutput(os.Stderr),
		tea.WithContext(ctx),
	}
	if o.FromStdin {
		// stdin was the item pipe, the keyboard is on the terminal.
		if ttyFile, ttyErr := os.Open("/dev/tty"); ttyErr == nil {
			defer ttyFile.Close() //nolint:errcheck
			programOpts = append(programOpts, tea.WithInput(ttyFile))
		}
	}

	tm, err := tea.NewProgram(m, programOpts...).Run()
	if err != nil {
		return err
	}

	result := tm.(orchestrator)

	if !result.submitted {
		// The user quit without submitting (esc/ctrl+c).  Exit 1 like gum does,
		// but say nothing: the TUI has just been torn down and an
		// "invalid argument" line would be left behind on the terminal.
		return exit.ErrExit(1)
	}

	var finalOutput string
	if o.Single {
		finalOutput = result.getSingleResult()
	} else {
		output := result.getResults(o.OutputDelimiter)
		if o.Stacked {
			finalOutput = strings.Join(output, "\n"+o.Delimiter+"\n")
		} else {
			finalOutput = strings.Join(output, "\n")
		}
	}

	tty.Println(finalOutput)
	return nil
}
