export VERSION := `echo v$(cat VERSION)-dev`

default:
    @just --list

# Build wiki binary into ./bin
build:
    mkdir -p bin
    go build -ldflags "-X main.Version={{VERSION}}" -o bin/wiki ./cmd/
    chmod +x bin/wiki

# Install wiki to GOPATH/bin (copies after tests pass)
install: test
    cp bin/wiki $(go env GOPATH)/bin/wiki

# Clean built binaries
clean:
    rm -rf bin
    rm -f $(go env GOPATH)/bin/wiki

# Run tests (builds first)
test: build
    PATH="$(pwd)/bin:$PATH" go test -v ./tests/

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
