VERSION ?= $(or $(shell git describe --tags --always --dirty),dev)

# Static Linux binary for the Vagrant VMs (runs on musl and glibc alike).
build: export GOOS=linux
build: export GOARCH=amd64
build: export CGO_ENABLED=0
build:
	go build -ldflags "-s -w -X main.version=$(VERSION)" -o dist/boxagent ./cmd/boxagent

test:
	go test ./...

# Install or update the agent on every running VM.
deploy: build
	vagrant provision --provision-with agent-bin,agent-install,agent

.PHONY: build test deploy
