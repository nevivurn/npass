BIN := npass

GIT_VERSION := $(shell git describe --always --dirty)
GO_LDFLAGS := -s -w -X 'main.version=$(GIT_VERSION)'
GO_FLAGS := -v -trimpath -ldflags '$(GO_LDFLAGS)'

.PHONY: all
all: $(BIN)

.PHONY: $(BIN)
$(BIN):
	go build $(GO_FLAGS) -o $(BIN) ./cmd/$(BIN)

.PHONY: check
check: lint test

.PHONY: lint
lint:
	golangci-lint run

.PHONY: test
test:
	go test -v ./...

.PHONY: clean
clean:
	go clean -x
