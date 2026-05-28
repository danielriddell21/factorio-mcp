MOD_VERSION := $(shell sed -n 's/.*"version"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' mod/factorio_mcp/info.json | head -1)
MOD_NAME := factorio_mcp
DIST := dist

.PHONY: build test fuzz lint vet fmt fmt-check mod-zip luacheck ci clean

build:
	go build -o bin/factorio-mcp ./cmd/factorio-mcp

test:
	go test ./...

fuzz:
	go test ./internal/factorio -run=NONE -fuzz=FuzzEscapeRoundTrip -fuzztime=20s

vet:
	go vet ./...

fmt:
	gofmt -s -w .

fmt-check:
	@out="$$(gofmt -s -l .)"; if [ -n "$$out" ]; then echo "gofmt needed:"; echo "$$out"; exit 1; fi

lint:
	golangci-lint run

luacheck:
	@command -v luacheck >/dev/null 2>&1 && luacheck mod/ || echo "luacheck not installed; skipping"

mod-zip:
	@rm -rf "$(DIST)/$(MOD_NAME)_$(MOD_VERSION)"
	@mkdir -p "$(DIST)/$(MOD_NAME)_$(MOD_VERSION)"
	@cp -r mod/$(MOD_NAME)/. "$(DIST)/$(MOD_NAME)_$(MOD_VERSION)/"
	@cd "$(DIST)" && zip -r -q "$(MOD_NAME)_$(MOD_VERSION).zip" "$(MOD_NAME)_$(MOD_VERSION)"
	@rm -rf "$(DIST)/$(MOD_NAME)_$(MOD_VERSION)"
	@echo "built $(DIST)/$(MOD_NAME)_$(MOD_VERSION).zip"

ci: fmt-check vet test

clean:
	rm -rf bin $(DIST)
