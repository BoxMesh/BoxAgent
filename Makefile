# Numeric like the releases (0.1.<n>): local builds count commits, CI uses the run number.
VERSION ?= 0.1.$(shell git rev-list --count HEAD)

# Static Linux binary for the Vagrant VMs (runs on musl and glibc alike).
build: export GOOS=linux
build: export GOARCH=amd64
build: export CGO_ENABLED=0
build:
	go build -ldflags "-s -w -X main.version=$(VERSION)" -o dist/boxagent ./cmd/boxagent

test:
	go test ./...

# Regenerate internal/pb after editing proto/device.proto (needs protoc, protoc-gen-go and protoc-gen-go-grpc).
proto:
	protoc --go_out=. --go_opt=module=boxagent --go-grpc_out=. --go-grpc_opt=module=boxagent proto/device.proto

# Install or update the agent on every running VM.
deploy: build
	vagrant provision --provision-with agent-bin,agent-install,agent

.PHONY: build test proto deploy
