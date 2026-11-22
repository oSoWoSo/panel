package completion

import (
	"fmt"
	"strings"

	"github.com/alecthomas/kong"
)

// panelType is a top-level panel block type. Each maps 1:1 to a subcommand
// name (kong generates the matching flags from the same model the run.go
// dispatcher uses).
type panelType string

const (
	panelChoose panelType = "choose"
	panelFilter panelType = "filter"
)

// completionContext is the position of the cursor within the completion
// command line, derived from the committed words (everything before the
// current word).
type completionContext struct {
	// block is the enclosing "" | "choose" | "filter" block. It is non-empty
	// only when a "--" block boundary followed by a panel-type token was seen.
	block string
	// boundary is true when the last committed word was the "--" boundary
	// marker, i.e. the cursor sits immediately after the "--" that separates
	// committed words from the current word. In that state only panel types
	// are offered (or nothing when the current word looks like a flag).
	boundary bool
}

// normalize splits the raw completion argv (ALL words after the binary, as
// delivered by the shell's custom completion request) into the current word
// and the committed words. Bash appends a trailing "" when the cursor is on
// an empty word (i.e. after a trailing space), so a value of "" for the
// cursor is represented as one or more trailing empty tokens; committed is
// everything before them.
//
// kong's generated bash/fish/zsh custom completion always send the words
// verbatim; when cur is empty (trailing space) the shell appends an extra ""
// word (bash also appends "" when the previous char is a space and cur is
// empty; fish and zsh append one "" for the empty current word). We therefore
// normalize by:
//
//	if the last argv element is "" → cur="" and committed=argv[:len-1]
//	                                            (multiple trailing "" are merged
//	                                             into a single empty cur)
//	else → cur=argv[len-1] and committed=argv[:len-1]
//
// RULE: strip ALL trailing "". If any were stripped, cur="". Otherwise cur is
// the last argument (which may itself be a "" that acts as a literal token
// only if it's not trailing — impossible per the contract, since "" is only
// ever a trailing current-word marker from the shell).
//
// A `--` can appear:
//   - as the committed-word separator before the subcommand words (when the
//     user typed `--` at the top level and then a panel type), in which case
//     we detect it during committed-walk,
//   - as a trailing token when the user typed `--` + SPACE (empty current
//     word): this is then the immediately-preceding committed word, and it
//     signals `boundary` (the user is about to start a new block).
func normalize(argv []string) (cur string, committed []string) {
	n := len(argv)
	for n > 0 && argv[n-1] == "" {
		n--
	}
	if n < len(argv) {
		return "", append([]string(nil), argv[:n]...)
	}
	if n == 0 {
		return "", nil
	}
	return argv[n-1], append([]string(nil), argv[:n-1]...)
}

// contextFor walks the committed words (the text before the current word) to
// determine the completion context.
//
// The custom-completion contract (mirroring kong v1.15.0's bash/fish/zsh
// templates) treats `--` as the kong separator that introduces a subcommand's
// argument block: `--` followed by a panel type (choose|filter) starts that
// block; a `--` in the final committed position marks the boundary after
// which the user is about to type a new panel type.
//
// Walking rules for a committed token at index i:
//   - if it equals "--":
//   - if i+2 <= len(committed) and committed[i+1] is a panel type
//     (choose|filter, case-insensitive): the block becomes that type and
//     both tokens are consumed (i += 2).
//   - else if i+1 == len(committed): this `--` is the LAST committed word
//     → boundary becomes true and walking stops (the cursor is right after
//     the `--`, so only panel types can follow).
//   - else: the `--` is a plain literal token (e.g. followed by a
//     non-panel-type word); consume just it (i += 1).
//   - otherwise: consume the token (i += 1).
func contextFor(committed []string) completionContext {
	ctx := completionContext{}
	for i := 0; i < len(committed); {
		if committed[i] == "--" {
			if i+1 < len(committed) {
				next := strings.ToLower(committed[i+1])
				if isPanelType(next) {
					ctx.block = next
					i += 2
					continue
				}
				i++ // literal "--" followed by a non-panel-type word
				continue
			}
			// "--" is the last committed word: the cursor sits right after
			// the block separator boundary.
			ctx.boundary = true
			break
		}
		i++
	}
	return ctx
}

func isPanelType(s string) bool {
	switch s {
	case string(panelChoose), string(panelFilter):
		return true
	}
	return false
}

// Model carries the kong models Serve needs to answer a completion request: the
// root panel model plus one model per panel block type. The caller owns the
// parsers (panel/run.go builds them with the same kong.Vars the run path uses),
// so the flag names offered here are the very same ones kong accepts.
type Model struct {
	Root   *kong.Node
	Choose *kong.Node
	Filter *kong.Node
}

// blockNode returns the kong model backing the given panel block. The empty
// block is the root panel model.
func (m Model) blockNode(block string) *kong.Node {
	switch block {
	case string(panelChoose):
		return m.Choose
	case string(panelFilter):
		return m.Filter
	default:
		return m.Root
	}
}

// flagCandidates returns the long flags of node that start with cur, in the
// order kong declares them.
func flagCandidates(node *kong.Node, cur string) []string {
	if node == nil {
		return nil
	}
	candidates := make([]string, 0, len(node.Flags))
	for _, flag := range node.Flags {
		if nonCompletableFlag(flag) {
			continue
		}
		name := "--" + flag.Name
		if strings.HasPrefix(name, cur) {
			candidates = append(candidates, name)
		}
	}
	return candidates
}

// Serve implements the custom completion server for the panel command. It
// receives the raw words the shell sends (everything after the binary name,
// plus an empty word marker where applicable) and returns the candidates, one
// per line, followed by a final line carrying the shell completion directive.
//
// The generated bash script (see bash.go) parses the reply as
// `directive=${out##*:}` / `out=${out%%:*}`, so the last line must be the
// directive introduced by a colon — `:4` for ShellCompDirectiveNoFileComp:
// print the candidates, offer no file completion. The request itself is
// `${words[0]} completion completeNoDesc ${args[*]}`, so panel/run.go strips
// both `completion` and `completeNoDesc` before calling Serve.
//
// The candidate set depends on the context:
//
//	boundary (right after `--`):
//	    if cur starts with "-" → (nothing; "--" already typed, flags don't
//	                              autocomplete after the boundary)
//	    else → [choose filter]
//	block "" (top level):
//	    if cur starts with "-" → the root panel flags starting with cur
//	    else → [choose filter]
//	block "choose" / "filter":
//	    if cur starts with "-" → that panel's flags starting with cur
//	    else → (nothing; the block collects items, not more words)
func Serve(model Model, argv []string) string {
	cur, committed := normalize(argv)
	ctx := contextFor(committed)

	var candidates []string
	switch {
	case ctx.boundary:
		if !strings.HasPrefix(cur, "-") {
			candidates = []string{string(panelChoose), string(panelFilter)}
		}
	case ctx.block == "":
		if strings.HasPrefix(cur, "-") {
			candidates = flagCandidates(model.Root, cur)
		} else {
			candidates = []string{string(panelChoose), string(panelFilter)}
		}
	default:
		if strings.HasPrefix(cur, "-") {
			candidates = flagCandidates(model.blockNode(ctx.block), cur)
		}
	}

	b := strings.Builder{}
	for _, c := range candidates {
		b.WriteString(c)
		b.WriteString("\n")
	}
	fmt.Fprintf(&b, ":%d\n", ShellCompDirectiveNoFileComp)
	return b.String()
}
