# Void Linux package

The template builds Panel from its published source tag, using Void's Go build
style. It installs the binary, MIT license, Bash/Fish/Zsh completions and man page.
Panel 2.0.2.001 requires Go 1.26.7 or newer; use an up-to-date Void build environment.

From this repository, copy the template into an existing `void-packages` checkout:

```sh
mkdir -p /path/to/void-packages/srcpkgs/panel
cp packaging/xbps-src/template /path/to/void-packages/srcpkgs/panel/template
cd /path/to/void-packages
./xbps-src binary-bootstrap
./xbps-src pkg panel
```

Install the resulting package (adjust the repository path for your architecture
and libc, for example `hostdir/binpkgs/musl`):

```sh
sudo xbps-install -R "$PWD/hostdir/binpkgs" panel
```

The template version and checksum refer to the published release, not the working
tree. When packaging a new release, update `version`, reset `revision` to `1`, and
update the archive checksum with `xgensum -i srcpkgs/panel/template` in the
`void-packages` checkout. Increase only `revision` for packaging-only changes to
the same release. `make version ...` does not update this template.
