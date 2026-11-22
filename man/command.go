// Package man the man command.
package man

import (
	"fmt"
	"io"

	"github.com/alecthomas/kong"
	mangokong "github.com/alecthomas/mango-kong"
	"github.com/muesli/roff"
)

// Copyright is the copyright section shared by every generated man page.
const Copyright = "(c) 2022-2024 Charmbracelet, Inc.\n" +
	"Released under MIT license."

// Man is the `panel man` sub-command that generates man pages.
type Man struct{}

// BeforeApply implements Kong BeforeApply hook.
func (m Man) BeforeApply(ctx *kong.Context) error {
	Page(ctx.Model, "A tool for glamorous shell scripts.", ctx.Stdout)
	ctx.Exit(0)
	return nil
}

// Page renders the man page for model in roff format and writes it to out.
// The description replaces the model's help text, so that the generated page
// carries a plain description instead of color escape sequences.
func Page(model *kong.Application, description string, out io.Writer) {
	model.Help = description
	page := mangokong.NewManPage(1, model)
	page = page.WithSection("Copyright", Copyright)
	_, _ = fmt.Fprint(out, page.Build(roff.NewDocument()))
}
