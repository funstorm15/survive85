# survive85 — common tasks. Run `make help` for the list.
BINARY  := survive85
PKG     := ./cmd/survive85
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE    ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)
PREFIX  ?= $(HOME)/.local

.DEFAULT_GOAL := help
.PHONY: help build run test race vet lint validate sim lifetables snapshot aur install uninstall clean

help: ## show this help
	@awk 'BEGIN{FS=":.*## "} /^[a-z-]+:.*## /{printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## build a static binary into ./survive85
	CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -o $(BINARY) $(PKG)

run: ## play (pass flags with ARGS="-lang ru -seed 42")
	go run $(PKG) $(ARGS)

test: ## run all tests
	go test ./...

race: ## run tests with the race detector
	go test -race ./...

vet: ## go vet
	go vet ./...

lint: ## golangci-lint (v2)
	golangci-lint run

validate: ## check the embedded game data
	go run $(PKG) validate

sim: ## balance report: careful vs random vs reckless
	go run $(PKG) sim -runs 3000 -region world -sex male
	go run $(PKG) sim -runs 3000 -region rus -sex male

lifetables: ## regenerate data/lifetables/who2019.yaml from the WHO APIs
	go run ./tools/genlifetables -out data/lifetables/who2019.yaml

snapshot: ## build release archives locally with GoReleaser (no publishing)
	goreleaser release --snapshot --clean

aur: ## print .SRCINFO for the AUR packages into dist/aur
	mkdir -p dist/aur/survive85 dist/aur/survive85-bin
	cp scripts/PKGBUILD dist/aur/survive85/PKGBUILD
	cp scripts/PKGBUILD-bin dist/aur/survive85-bin/PKGBUILD
	cd dist/aur/survive85 && makepkg --printsrcinfo > .SRCINFO
	cd dist/aur/survive85-bin && makepkg --printsrcinfo > .SRCINFO

install: build ## install into $(PREFIX)/bin
	install -Dm755 $(BINARY) $(PREFIX)/bin/$(BINARY)

uninstall: ## remove from $(PREFIX)/bin
	rm -f $(PREFIX)/bin/$(BINARY)

clean: ## remove build output
	rm -rf $(BINARY) dist
