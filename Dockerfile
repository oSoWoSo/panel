# The build context must contain the compiled `panel` binary: goreleaser puts
# it there itself, for a manual build run `make build && cp bin/panel .`.
FROM gcr.io/distroless/static
COPY panel /usr/local/bin/panel
ENTRYPOINT [ "/usr/local/bin/panel" ]
