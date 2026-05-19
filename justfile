export VERSION := `git describe --tags --always --dirty 2>/dev/null || echo "dev"`

default:
    @just --list

# Build wiki binary locally
build:
    go build -ldflags "-X 'github.com/pavelpivovarov/glow/cmd.Version={{VERSION}}'" -o wiki ./cmd/wiki

# Install wiki to GOPATH/bin
install: build test
    go install -ldflags "-X 'github.com/pavelpivovarov/glow/cmd.Version={{VERSION}}'" ./cmd/wiki

# Clean built binaries
clean:
    rm -f wiki
    rm -f $(go env GOPATH)/bin/wiki

# Run tests (requires build first)
test: build
    PATH="$(pwd):$PATH" go test -v ./tests/

# Format code
fmt:
    go fmt ./...

# Show version
version:
    @echo {{VERSION}}

# Build and test
all: fmt test build

# Create a new release tag
tag version:
    git tag -a v{{version}} -m "Release v{{version}}"
    @echo "Created tag v{{version}}"
    @echo "Push with: git push origin v{{version}}"
