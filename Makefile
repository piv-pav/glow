.PHONY: build install clean test version

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -X 'github.com/pavelpivovarov/glow/cmd.Version=$(VERSION)'

build:
	go build -ldflags "$(LDFLAGS)" -o wiki ./cmd/wiki

install:
	go install -ldflags "$(LDFLAGS)" ./cmd/wiki

clean:
	rm -f wiki
	rm -f $(shell go env GOPATH)/bin/wiki

test:
	go test ./...

version:
	@echo $(VERSION)

help:
	@echo "GLOW Wiki - Build Targets"
	@echo ""
	@echo "  make build    - Build wiki binary locally"
	@echo "  make install  - Install wiki to GOPATH/bin"
	@echo "  make clean    - Remove built binaries"
	@echo "  make test     - Run tests"
	@echo "  make version  - Show version"
	@echo ""
	@echo "VERSION=$(VERSION)"
