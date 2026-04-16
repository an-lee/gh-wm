# gh-wm — local development shortcuts (CI parity: make ci)

BINARY ?= gh-wm

# Build flags: strip debug symbols and DWARF info to shrink binary (~40-50% smaller)
BUILD_FLAGS ?= -trimpath -ldflags="-s -w"

.PHONY: build test bench vet fmt fmt-check ci clean install docs

# Default: produce ./gh-wm at repo root with size-optimised flags
build:
	go build $(BUILD_FLAGS) -o $(BINARY) .

test:
	go test ./...

bench:
	go test -bench=. -benchmem ./...

vet:
	go vet ./...

fmt:
	gofmt -w .

fmt-check:
	test -z "$$(gofmt -l .)"

ci: fmt-check vet test
	go build -v $(BUILD_FLAGS) ./...

clean:
	rm -f $(BINARY)

install:
	go install

# Hugo docs site (requires hugo on PATH); see docs/content/development.md
docs:
	cd docs && hugo mod get -u && hugo server
