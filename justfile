set shell := ["bash", "-lc"]

default:
    just --list

init:
    just mod-tidy

mod-tidy:
    go mod tidy

fmt:
    go fmt ./...

fmt-check:
    test -z "$(gofmt -l $(find . -type f -name '*.go' -not -path './vendor/*'))"

lint:
    go vet ./...

lint-check:
    go vet ./...

test:
    go test ./...

run:
    go run ./cmd/api

build:
    go build ./cmd/api

ci:
    just mod-tidy && \
    just fmt-check && \
    just lint-check && \
    just test
