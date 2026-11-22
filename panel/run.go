package panel

import (
	"errors"
	"fmt"
	"io"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/alecthomas/kong"
	"github.com/oSoWoSo/panel/v2/choose"
	"github.com/oSoWoSo/panel/v2/completion"
	"github.com/oSoWoSo/panel/v2/filter"
	"github.com/oSoWoSo/panel/v2/input"
	"github.com/oSoWoSo/panel/v2/internal/exit"
	"github.com/oSoWoSo/panel/v2/man"
	"github.com/oSoWoSo/panel/v2/write"
)

const (
	boolFalse = "false"
	shaLen    = 7

	// inputCommand and writeCommand are the prompt subcommand names, shared by
	// the dispatch in RunProgram and by runPrompt. versionVar is the kong
	// interpolation variable backing --version.
	inputCommand = "input"
	writeCommand = "write"
	versionVar   = "version"
)

// kongVars are the default variables required by choose/filter Options structs.
// "version" is used by the panel's --version flag.
var kongVars = kong.Vars{
	"version":                 "",
	"defaultHeight":           "0",
	"defaultWidth":            "0",
	"defaultAlign":            "left",
	"defaultBorder":           "none",
	"defaultBorderForeground": "",
	"defaultBorderBackground": "",
	"defaultBackground":       "",
	"defaultForeground":       "",
	"defaultMargin":           "0 0",
	"defaultPadding":          "0 0",
	"defaultUnderline":        boolFalse,
	"defaultBold":             boolFalse,
	"defaultFaint":            boolFalse,
	"defaultItalic":           boolFalse,
	"defaultStrikethrough":    boolFalse,
}

// newParser builds the kong parser for the panel command line. exit is invoked
// instead of os.Exit when kong handles --help/--version itself.
func newParser(version string, exit func(int)) (*kong.Kong, error) {
	return kong.New(
		&Options{},
		kong.Name("panel"),
		kong.Description("Run multiple choose/filter/input panels side by side"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact:             true,
			Summary:             false,
			NoExpandSubcommands: true,
		}),
		kong.Exit(exit),
		kongVars,
		kong.Vars{versionVar: version},
	)
}

// runCompletion dispatches the `panel completion <subcommand>` invocations that
// the panel command line exposes outside kong: kong v1.15.0 has no completion
// machinery of its own, so panel ships the generators (bash/zsh/fish) and the
// dynamic server (completeNoDesc) itself.
//
//	bash|zsh|fish    → print the generated shell script
//	completeNoDesc   → answer the shell's dynamic candidate request
func runCompletion(parser *kong.Kong, args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "panel completion: expected bash, fish, zsh or completeNoDesc")
		return 1
	}
	switch args[0] {
	case "bash":
		return completionExit(completion.Bash{}.Run(&kong.Context{Kong: parser}))
	case "zsh":
		return completionExit(completion.Zsh{}.Run(&kong.Context{Kong: parser}))
	case "fish":
		return completionExit(completion.Fish{}.Run(&kong.Context{Kong: parser}))
	case "completeNoDesc":
		fmt.Print(completion.Serve(completionModels(parser), args[1:]))
		return 0
	default:
		fmt.Fprintf(os.Stderr, "panel completion: unknown completion command %q\n", args[0])
		return 1
	}
}

func completionExit(err error) int {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}

// completionModels builds the kong models the dynamic server completes
// against: the root panel model plus the choose/filter block models, parsed
// with the same kong.Vars the panel run path uses so the offered flags match
// the ones kong actually accepts.
func completionModels(parser *kong.Kong) completion.Model {
	model := completion.Model{Root: parser.Model.Node}
	if chooseParser, err := kong.New(&choose.Options{}, kong.Exit(func(int) {}), kongVars); err == nil {
		model.Choose = chooseParser.Model.Node
	}
	if filterParser, err := kong.New(&filter.Options{}, kong.Exit(func(int) {}), kongVars); err == nil {
		model.Filter = filterParser.Model.Node
	}
	return model
}

// The input and write prompts are dispatched here for the same reason: the
// root model carries a positional panel argument, so kong would treat a
// leading "input"/"write" as the first panel block instead of a subcommand.
//
//	input [flags]    → single line prompt on stdout
//	write [flags]    → multi-line textarea on stdout
func runPrompt(version string, args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "panel: expected input or write")
		return 1
	}
	var (
		model       any
		name        string
		description string
	)
	switch args[0] {
	case inputCommand:
		model, name, description = &input.Options{}, inputCommand, "Prompt for some input"
	case writeCommand:
		model, name, description = &write.Options{}, writeCommand, "Prompt for long-form text"
	default:
		fmt.Fprintf(os.Stderr, "panel: unknown command %q\n", args[0])
		return 1
	}
	return runPromptSubcommand(version, name, description, model, args[1:], os.Stdout)
}

// runPromptSubcommand parses args against its own kong application — built
// with the same kongVars as the panel root so ${defaultPadding}-style defaults
// resolve — and runs the model. out receives kong's own output (--help).
func runPromptSubcommand(version, name, description string, model any, args []string, out io.Writer) int {
	var (
		exited   bool
		exitCode int
	)
	parser, err := kong.New(
		model,
		kong.Name("panel "+name),
		kong.Description(description),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact:             true,
			Summary:             false,
			NoExpandSubcommands: true,
		}),
		kong.Writers(out, os.Stderr),
		kong.Exit(func(code int) {
			exited = true
			exitCode = code
		}),
		kongVars,
		kong.Vars{versionVar: version},
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	ctx, err := parser.Parse(args)
	if err != nil {
		if exited {
			// --help was handled by kong and already returned.
			return exitCode
		}
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if exited {
		// --help was handled by kong and already returned.
		return exitCode
	}
	if err := ctx.Run(); err != nil {
		return exitCodeFromError(err)
	}
	return 0
}

// manDescription is the plain description printed on the generated man page,
// in place of the colorized panel help.
const manDescription = "Glamorous panels for your shell scripts."

// runMan renders the roff man page for the panel command line. Like the
// completion dispatch, it happens outside kong: the root model carries the
// positional panel argument, so a kong `man` subcommand would be swallowed by
// it, and rendering from the parser model keeps the page free of a `man`
// entry describing itself.
func runMan(parser *kong.Kong, args []string) int {
	if len(args) > 0 {
		fmt.Fprintf(os.Stderr, "panel man: unexpected argument %q\n", args[0])
		return 1
	}
	man.Page(parser.Model, manDescription, parser.Stdout)
	return 0
}

// VersionString builds the version string printed by --version.
func VersionString(version, commitSHA string) string {
	s := "panel version " + version
	if len(commitSHA) >= shaLen {
		s += " (" + commitSHA[:shaLen] + ")"
	}
	return s
}

// RunProgram parses args against the panel Options and runs the panel TUI,
// returning the process exit code.
//
// The completion, man, input and write command lines are dispatched here,
// outside kong: kong v1.15.0 has no completion machinery of its own, so panel's
// own completion package owns both the shell script generators and the dynamic
// candidate server (see completion.Serve), and the root model carries a
// positional argument that a kong subcommand would collide with.
func RunProgram(args []string, version string) int {
	var (
		exited   bool
		exitCode int
	)
	parser, err := newParser(version, func(code int) {
		exited = true
		exitCode = code
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if len(args) > 0 {
		switch args[0] {
		case "completion":
			return runCompletion(parser, args[1:])
		case "man":
			return runMan(parser, args[1:])
		case inputCommand, writeCommand:
			return runPrompt(version, args)
		}
	}
	ctx, err := parser.Parse(args)
	if err != nil {
		if exited {
			// --help / --version already handled the invocation and exited.
			return exitCode
		}
		parser.FatalIfErrorf(err)
		if !exited {
			return 1
		}
		return exitCode
	}
	if exited {
		// --help / --version were handled by kong and already returned.
		return exitCode
	}
	if err := ctx.Run(); err != nil {
		return exitCodeFromError(err)
	}
	return 0
}

// exitCodeFromError translates a Run() error into a process exit code,
// mirroring the error handling the panel command line uses elsewhere.
func exitCodeFromError(err error) int {
	var ex exit.ErrExit
	if errors.As(err, &ex) {
		return int(ex)
	}
	if errors.Is(err, tea.ErrInterrupted) {
		return exit.StatusAborted
	}
	if errors.Is(err, tea.ErrProgramKilled) {
		fmt.Fprintln(os.Stderr, "timed out")
		return exit.StatusTimeout
	}
	fmt.Fprintln(os.Stderr, err)
	return 1
}
