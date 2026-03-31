BINARY    := ddctl
MODULE    := github.com/futuregerald/ddctl
VERSION   := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT    := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE      := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
LDFLAGS   := -ldflags "-X $(MODULE)/cmd.Version=$(VERSION) -X $(MODULE)/cmd.CommitSHA=$(COMMIT) -X $(MODULE)/cmd.BuildDate=$(DATE)"

.PHONY: build test lint clean install

build:
	go build $(LDFLAGS) -o $(BINARY) .

test:
	go test ./...

lint:
	go vet ./...

clean:
	rm -f $(BINARY)

install:
	go install $(LDFLAGS) .
