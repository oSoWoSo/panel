package panel

import (
	"testing"

	"github.com/alecthomas/kong"
	"github.com/oSoWoSo/panel/v2/choose"
	"github.com/oSoWoSo/panel/v2/filter"
	"github.com/oSoWoSo/panel/v2/input"
	"github.com/oSoWoSo/panel/v2/write"
)

func TestPanelEnvironmentVariables(t *testing.T) {
	root := &Options{}
	chooseOpts := &choose.Options{}
	filterOpts := &filter.Options{}
	inputOpts := &input.Options{}
	writeOpts := &write.Options{}

	for _, tc := range []struct {
		name  string
		env   string
		model any
		args  []string
		got   func() string
	}{
		{"root", "PANEL_HEADER_TEXT", root, []string{"--", "choose", "item"}, func() string { return root.HeaderText }},
		{"choose", "PANEL_CHOOSE_HEADER", chooseOpts, nil, func() string { return chooseOpts.Header }},
		{"filter", "PANEL_FILTER_HEADER", filterOpts, nil, func() string { return filterOpts.Header }},
		{"input", "PANEL_INPUT_HEADER", inputOpts, nil, func() string { return inputOpts.Header }},
		{"write", "PANEL_WRITE_HEADER", writeOpts, nil, func() string { return writeOpts.Header }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			const want = "configured through PANEL environment"
			t.Setenv(tc.env, want)
			parser, err := kong.New(tc.model, kongVars)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := parser.Parse(tc.args); err != nil {
				t.Fatal(err)
			}
			if got := tc.got(); got != want {
				t.Fatalf("header = %q, want %q from %s", got, want, tc.env)
			}
		})
	}
}
