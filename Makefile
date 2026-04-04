SHELL := /bin/sh

APP_NAME := zatools
MAIN_PACKAGE := ./cmd/zatools
DIST_DIR := dist
BUILD_DIR := $(DIST_DIR)/bin
RELEASE_DIR := $(DIST_DIR)/release
CURRENT_OS := $(shell go env GOOS)
CURRENT_ARCH := $(shell go env GOARCH)
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
TARGETS := darwin/amd64 darwin/arm64 linux/amd64 linux/arm64 windows/amd64 windows/arm64

BIN_EXT :=
ifeq ($(CURRENT_OS),windows)
BIN_EXT := .exe
endif

.PHONY: help fmt test test-race vet lint clean build install-local build-all release checksums

help:
	@printf '%s\n' \
		'make build         Build the current platform binary' \
		'make install-local Install the current platform binary to PREFIX/bin' \
		'make build-all     Cross-build all supported release binaries' \
		'make release       Create release archives and checksums' \
		'make fmt           Run gofmt on tracked Go files' \
		'make test          Run unit tests' \
		'make test-race     Run tests with the race detector' \
		'make vet           Run go vet' \
		'make lint          Run golangci-lint if installed' \
		'make clean         Remove dist artifacts'

fmt:
	@gofmt -w $$(find . -name '*.go' -not -path './dist/*')

test:
	@go test ./...

test-race:
	@go test -race ./...

vet:
	@go vet ./...

lint:
	@golangci-lint run

clean:
	@rm -rf $(DIST_DIR)

build:
	@mkdir -p $(BUILD_DIR)/$(CURRENT_OS)-$(CURRENT_ARCH)
	@CGO_ENABLED=0 GOOS=$(CURRENT_OS) GOARCH=$(CURRENT_ARCH) \
		go build -trimpath -o $(BUILD_DIR)/$(CURRENT_OS)-$(CURRENT_ARCH)/$(APP_NAME)$(BIN_EXT) $(MAIN_PACKAGE)
	@printf 'built %s\n' "$(BUILD_DIR)/$(CURRENT_OS)-$(CURRENT_ARCH)/$(APP_NAME)$(BIN_EXT)"

install-local: build
	@set -eu; \
	prefix="$${PREFIX:-/usr/local}"; \
	install_dir="$$prefix/bin"; \
	src="$(BUILD_DIR)/$(CURRENT_OS)-$(CURRENT_ARCH)/$(APP_NAME)$(BIN_EXT)"; \
	mkdir -p "$$install_dir"; \
	install -m 0755 "$$src" "$$install_dir/$(APP_NAME)$(BIN_EXT)"; \
	printf 'installed %s\n' "$$install_dir/$(APP_NAME)$(BIN_EXT)"

build-all:
	@set -eu; \
	for target in $(TARGETS); do \
		os="$${target%/*}"; \
		arch="$${target#*/}"; \
		ext=""; \
		if [ "$$os" = "windows" ]; then ext=".exe"; fi; \
		out_dir="$(BUILD_DIR)/$${os}-$${arch}"; \
		mkdir -p "$$out_dir"; \
		printf 'building %s/%s\n' "$$os" "$$arch"; \
		CGO_ENABLED=0 GOOS="$$os" GOARCH="$$arch" \
			go build -trimpath -o "$$out_dir/$(APP_NAME)$$ext" $(MAIN_PACKAGE); \
	done

release: clean test vet build-all checksums
	@printf 'release artifacts written to %s\n' "$(RELEASE_DIR)"

checksums:
	@set -eu; \
	mkdir -p "$(RELEASE_DIR)"; \
	for target in $(TARGETS); do \
		os="$${target%/*}"; \
		arch="$${target#*/}"; \
		ext=""; \
		if [ "$$os" = "windows" ]; then ext=".exe"; fi; \
		asset="$(APP_NAME)_$(VERSION)_$${os}_$${arch}.tar.gz"; \
		printf 'packaging %s\n' "$$asset"; \
		LC_ALL=C tar -C "$(BUILD_DIR)/$${os}-$${arch}" -czf "$(RELEASE_DIR)/$$asset" "$(APP_NAME)$$ext"; \
	done; \
	cd "$(RELEASE_DIR)"; \
	if command -v sha256sum >/dev/null 2>&1; then \
		sha256sum ./*.tar.gz > checksums.txt; \
	elif command -v shasum >/dev/null 2>&1; then \
		shasum -a 256 ./*.tar.gz > checksums.txt; \
	else \
		printf 'missing sha256sum or shasum\n' >&2; \
		exit 1; \
	fi
