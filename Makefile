BIN := npass

.PHONY: all
all: $(BIN)

.PHONY: $(BIN)
$(BIN):
	go build -v -o $(BIN) .

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
	rm -f $(BIN)
