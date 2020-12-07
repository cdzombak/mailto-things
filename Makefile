SHELL:=/usr/bin/env bash
VERSION:=$(shell [ -z "$$(git tag)" ] && echo "$$(git describe --always --long --dirty)" || echo "$$(git tag | sed 's/^v//')")
GO_FILES:=$(shell find . -name '*.go' | grep -v /vendor/)
BIN_NAME:=mailto-things

default: help

# via https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
.PHONY: help
help: ## Print help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: all
all: clean build build-linux-amd64 build-darwin-amd64  ## Build for macOS and Linux on amd64

.PHONY: clean
clean: ## Remove built products in ./out
	rm -rf ./out

.PHONY: lint
lint: ## Lint all .go files
	@for file in ${GO_FILES} ;  do \
		echo "$$file" ; \
		golint $$file ; \
	done

.PHONY: build
build: lint ## Build (for the current platform & architecture) to ./out
	mkdir -p out
	go build -ldflags="-X main.version=${VERSION}" -o ./out/${BIN_NAME} .

.PHONY: build-linux-amd64
build-linux-amd64: ## Build for Linux/amd64 to ./out
	mkdir -p out/linux-amd64
	env GOOS=linux GOARCH=amd64 go build -ldflags="-X main.version=${VERSION}" -o ./out/linux-amd64/${BIN_NAME} .

.PHONY: build-darwin-amd64
build-darwin-amd64: ## Build for macOS (Darwin) / amd64 to ./out
	mkdir -p out/darwin-amd64
	env GOOS=darwin GOARCH=amd64 go build -ldflags="-X main.version=${VERSION}" -o ./out/darwin-amd64/${BIN_NAME} .

.PHONY: install
install: lint ## Build & install mailto-things to /usr/local/bin
	go build -ldflags="-X main.version=${VERSION}" -o /usr/local/bin/${BIN_NAME} .

.PHONY: cdzombak-deploy-burr
cdzombak-deploy-burr: clean build-linux-amd64
	scp ./out/linux-amd64/mailto-things burr:~/mailto-things
