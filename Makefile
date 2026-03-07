.PHONY: build test e2e fmt vet lint check clean all

BIN := /tmp/kno

all: check test e2e

build:
	go build -ldflags "-X github.com/kno-ai/kno/internal.Version=dev" -o $(BIN) ./cmd/kno

test:
	go test ./...

e2e: build
	./test/e2e_test.sh

fmt:
	gofmt -w .
	@command -v goimports >/dev/null 2>&1 && goimports -w . || true

fmt-check:
	@test -z "$$(gofmt -l .)" || (echo "Files not formatted:"; gofmt -l .; exit 1)

vet:
	go vet ./...

lint: check
	@command -v staticcheck >/dev/null 2>&1 && staticcheck ./... || echo "staticcheck not installed — skipping (go install honnef.co/go/tools/cmd/staticcheck@latest)"

check: fmt-check vet

clean:
	rm -f $(BIN)
