# panel

<p>
    <a href="https://github.com/oSoWoSo/panel/releases"><img src="https://img.shields.io/github/release/oSoWoSo/panel.svg" alt="Latest Release"></a>
    <a href="https://github.com/oSoWoSo/panel/actions/workflows/build.yml"><img src="https://img.shields.io/github/actions/workflow/status/oSoWoSo/panel/build.yml?branch=main" alt="Build Status"></a>
    <a href="https://pkg.go.dev/github.com/oSoWoSo/panel/v2?tab=doc"><img src="https://img.shields.io/badge/go-docs-blue?logo=go" alt="Go Docs"></a>
</p>

Glamorous panels for your shell scripts: put **choose**, **filter** and
**input** panels side by side, collect an answer from all of them, and print
the results to stdout. Built with [Bubbles](https://github.com/charmbracelet/bubbles)
and [Lip Gloss](https://github.com/charmbracelet/lipgloss), driven from your
aliases and scripts — no Go code required.

## Installation

```bash
# from source
git clone https://github.com/oSoWoSo/panel
cd panel && make install

# or with the Go toolchain
go install github.com/oSoWoSo/panel/v2/cmd/panel@latest
```

Prebuilt archives for Linux, macOS and Windows (amd64 + arm64) are attached to
every [release](https://github.com/oSoWoSo/panel/releases).

## Build version

The `VERSION` file records the upstream Gum version followed by the Panel
revision: `2.0.2.001` means Gum 2.0.2, Panel revision 001.

```bash
make version 2.0.2.002
make build
./bin/panel --version
```

Commit the updated `VERSION` file when changing the version. This target does
not create tags or publish releases. For a one-off build without changing the
file, use `make build VERSION=2.0.2.002`.

This setting controls Makefile builds; GoReleaser still uses its release or
snapshot version. Four-part Git tags require a separate release-pipeline change.

## Usage

A panel run is a sequence of blocks, each introduced by `--` and a panel type.
The items of a block are the things you choose from.

```bash
panel -- choose "Deploy" "Rollback" -- filter "prod" "staging" -- input "reason"
```

The three panels open next to each other, and their answers are printed to
stdout, separated by `---`:

```
Deploy
---
prod
---
reason
```

### Key bindings

| Key            | Action                       |
| -------------- | ---------------------------- |
| `tab`          | Next panel                   |
| `shift+tab`    | Previous panel               |
| `←` `→`        | Previous/next item           |
| `enter`        | Select in this panel, then move on; on the last panel it submits |
| `ctrl+s`       | Submit every panel and exit  |
| `ctrl+h`       | Toggle the help footer       |
| `esc`          | Quit                         |

### Panels

| Type     | What it does                                                        |
| -------- | ------------------------------------------------------------------- |
| `choose` | Pick one item from the block                                        |
| `filter` | Fuzzy-find among the items and mark any number of them              |
| `input`  | Type a free-form value; the first positional is the initial content |

Input panels take their own flags, for example:

```bash
# pre-filled, masked, multiline, with a header
panel -- input "https://example.com" --multiline --password --header "URL:"
```

In a multiline input panel `ctrl+j` inserts a newline; `enter` still moves to
the next panel, so a value can never be submitted by accident.

### Output

```bash
# one line per panel value, no separator
panel --no-stacked -- choose a b -- input c

# custom separator between panels
panel --stacked --delimiter "===" -- choose a b -- input c

# several filter selections joined by a pipe, whole run on one line
panel --output-delimiter "|" -- filter a b c
```

Use `panel --help` for the full list of layout, output and style flags. Every
configuration can also be set through `PANEL_*` environment variables.

## Shell integration

`panel` can generate its own completion scripts, and a man page:

```bash
panel completion bash > /etc/bash_completion.d/panel
panel completion zsh  > "${fpath[1]}/_panel"
panel completion fish > ~/.config/fish/completions/panel.fish

panel man | man -l -
```

For multi-panel completion, `panel` answers the shell itself: inside a block
(`--` after the first one) you get the flags of that panel's type, and before
it the root flags and the `choose`/`filter`/`input` keywords.

## License

[MIT](LICENSE) — forked from [gum](https://github.com/charmbracelet/gum) by
Charm, © 2022-2024 Charmbracelet, Inc.
