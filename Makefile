# ---- Config ----
APP      := sip-options-once
PKG      := .
LDFLAGS  := -s -w
GO       := go

# Desative CGO para binários estáticos e cross-compile fácil
export CGO_ENABLED=0

# ---- Alvos principais ----
.PHONY: all build clean linux-amd64 linux-arm64 macos-arm64 macos-amd64 tidy deps help

all: build linux-amd64 linux-arm64 macos-amd64

## build: compila para sua máquina atual (macOS arm64, no seu caso)
build: deps
	$(GO) build -trimpath -ldflags="$(LDFLAGS)" -o $(APP) $(PKG)

## linux-amd64: cross-compile para Linux x86_64
linux-amd64: deps
	GOOS=linux GOARCH=amd64 $(GO) build -trimpath -ldflags="$(LDFLAGS)" -o $(APP)-linux-amd64 $(PKG)

## linux-arm64: cross-compile para Linux arm64
linux-arm64: deps
	GOOS=linux GOARCH=arm64 $(GO) build -trimpath -ldflags="$(LDFLAGS)" -o $(APP)-linux-arm64 $(PKG)

## macos-arm64: compila para macOS Apple Silicon
macos-arm64: deps
	GOOS=darwin GOARCH=arm64 $(GO) build -trimpath -ldflags="$(LDFLAGS)" -o $(APP)-macos-arm64 $(PKG)

## macos-amd64: cross-compile para macOS Intel
macos-amd64: deps
	GOOS=darwin GOARCH=amd64 $(GO) build -trimpath -ldflags="$(LDFLAGS)" -o $(APP)-macos-amd64 $(PKG)

## deps: garante módulos atualizados
deps: tidy
	$(GO) mod download

## tidy: ajusta go.mod/go.sum
tidy:
	$(GO) mod tidy

## clean: remove binários gerados
clean:
	rm -f $(APP) $(APP)-linux-amd64 $(APP)-linux-arm64 $(APP)-macos-arm64 $(APP)-macos-amd64

## help: mostra esta ajuda
help:
	@grep -E '^##' Makefile | sed 's/^## //'
