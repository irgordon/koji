.PHONY: fmt test build build-web build-kojid build-agent

fmt:
	go fmt ./...

test:
	go test ./...

build: build-web build-kojid build-agent

build-web:
	npm --prefix web run build

build-kojid:
	go build -o dist/kojid ./cmd/kojid

build-agent:
	go build -o dist/koji-agent ./cmd/koji-agent
