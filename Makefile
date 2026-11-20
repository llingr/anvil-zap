# Built with a Go at least as new as go.mod's go directive, which it refuses otherwise
GOLANGCI_LINT_VERSION := v2.15.0

default: build test

all: build test lint mutation

build:
	go build ./...
	go vet ./...

test:
	go test -race -coverpkg=./zaplog/... -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

# same image and version as the CI lint job, nothing installed on the host
lint:
	docker run --rm -v "$(CURDIR)":/app -w /app \
		golangci/golangci-lint:$(GOLANGCI_LINT_VERSION) golangci-lint run ./...

mutation:
	gremlins unleash --coverpkg=./zaplog/... --workers 2 \
		--invert-assignments --invert-bitwise --invert-bwassign --invert-logical \
		--invert-loopctrl --remove-self-assignments --invert-negatives \
		--threshold-efficacy 95 --threshold-mcover 90

.PHONY: default all build test lint mutation
