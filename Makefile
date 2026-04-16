# gh-wm — local development shortcuts (CI parity: make ci)

BINARY ?= gh-wm

.PHONY: build test vet fmt fmt-check ci clean install docs

# Default: produce ./gh-wm at repo root
build:
	go build -o $(BINARY) .

test:
	go test ./...

vet:
	go vet ./...

fmt:
	gofmt -w .

fmt-check:
	test -z "$$(gofmt -l .)"

ci: fmt-check vet test
	go build -v ./...

clean:
	rm -f $(BINARY)

install:
	go install

# Hugo docs site (requires hugo on PATH); see docs/content/development.md
docs:
	cd docs && hugo mod get -u && hugo server
