mod_version := `sed -n 's/.*"version"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' mod/factorio_mcp/info.json | head -1`
mod_name := "factorio_mcp"

# Build the MCP server binary into bin/.
build:
    go build -o bin/factorio-mcp ./cmd/factorio-mcp

# Run all Go tests.
test:
    go test ./...

# Fuzz the Lua escaping.
fuzz:
    go test ./internal/factorio -run=NONE -fuzz=FuzzEscapeRoundTrip -fuzztime=20s

# go vet.
vet:
    go vet ./...

# Format with gofmt -s.
fmt:
    gofmt -s -w .

# Fail if any file needs formatting.
fmt-check:
    #!/usr/bin/env sh
    out="$(gofmt -s -l .)"
    if [ -n "$out" ]; then echo "gofmt needed:"; echo "$out"; exit 1; fi

# Lint (requires golangci-lint).
lint:
    golangci-lint run

# Package the mod into dist/<name>_<version>.zip.
mod-zip:
    #!/usr/bin/env sh
    set -e
    dir="dist/{{mod_name}}_{{mod_version}}"
    rm -rf "$dir"
    mkdir -p "$dir"
    cp -r mod/{{mod_name}}/. "$dir/"
    (cd dist && zip -r -q "{{mod_name}}_{{mod_version}}.zip" "{{mod_name}}_{{mod_version}}")
    rm -rf "$dir"
    echo "built dist/{{mod_name}}_{{mod_version}}.zip"

# fmt-check + vet + test.
ci: fmt-check vet test

clean:
    rm -rf bin dist
