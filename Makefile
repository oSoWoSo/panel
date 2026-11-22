.DEFAULT_GOAL := usage

.PHONY: help build version test test-all lint lint-all clean install uninstall run-panel run-input run-write fmt usage

usage: help
	@exit 1

BINARY_NAME=panel
BIN_DIR=bin
INSTALL_PATH=$(DESTDIR)/usr/bin
GO=go
GOFLAGS=-v
# A static binary is what the distroless Docker image needs; every dependency
# is pure Go, so cgo only costs portability here.
CGO_ENABLED ?= 0
VERSION ?= $(strip $(file <VERSION))
export VERSION

# Support the shorthand: make version 2.0.2.001
ifeq ($(firstword $(MAKECMDGOALS)),version)
ifneq ($(words $(MAKECMDGOALS)),2)
$(error Usage: make version 2.0.2.001)
endif
VERSION := $(word 2,$(MAKECMDGOALS))
# Consume the positional version so Make does not treat it as another target.
.PHONY: $(VERSION)
$(VERSION): version
	@:
endif

COMMIT  ?= $(shell git rev-parse HEAD 2>/dev/null || echo "unknown")

LDFLAGS = -X main.Version=$(VERSION) -X main.CommitSHA=$(COMMIT)

build:
	mkdir -p $(BIN_DIR)
	CGO_ENABLED=$(CGO_ENABLED) $(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$(BINARY_NAME) ./cmd/panel

# Persist a version without creating a Git tag or publishing a release.
version:
	@printf '%s\n' "$$VERSION" | grep -Eq '^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]{3,}$$' || { \
		echo 'Usage: make version 2.0.2.001 (upstream major.minor.patch + Panel revision)' >&2; \
		exit 1; \
	}
	@printf '%s\n' "$$VERSION" > VERSION
	@printf 'Version set to %s\n' "$$VERSION"

test:
	$(GO) test -v ./panel/

test-all:
	$(GO) test -v ./...

lint:
	golangci-lint run ./panel/

lint-all:
	golangci-lint run ./...

clean:
	rm -rf $(BIN_DIR)

install: build
	install -m 755 $(BIN_DIR)/$(BINARY_NAME) $(INSTALL_PATH)/$(BINARY_NAME)

uninstall:
	rm -f $(INSTALL_PATH)/$(BINARY_NAME)

run-panel:
	$(GO) run ./cmd/panel -- choose a b c filter x y z

run-input:
	$(GO) run ./cmd/panel input

run-write:
	$(GO) run ./cmd/panel write

fmt:
	gofumpt -l -w .

help:
	@echo "Available targets:"
	@echo "  help          - Show Build targets"
	@echo "  clean         - Remove built binary"
	@echo "  lint          - Run linter on panel package"
	@echo "  lint-all      - Run linter on all packages"
	@echo "  fmt           - Format code with gofumpt"
	@echo "  test          - Run panel tests"
	@echo "  test-all      - Run all tests"
	@echo "  build         - Build the panel binary into bin/"
	@echo "  version       - Save version: make version 2.0.2.001"
	@echo "  install       - Install panel to /usr/bin"
	@echo "  uninstall     - Remove panel from /usr/bin"
	@echo "  run-panel     - Run a panel example"
	@echo "  run-input     - Run the input prompt"
	@echo "  run-write     - Run the write prompt"
