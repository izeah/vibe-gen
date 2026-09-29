SHELL := /bin/sh
BINARY   := vibe-gen
BUILD_DIR := ./out
CMD      := ./cmd/vibe-gen

.PHONY: build install clean test run-tui

## build: Build the vibe-gen binary to ./out/
build:
	@mkdir -p $(BUILD_DIR)
	go build -v -o $(BUILD_DIR)/$(BINARY) $(CMD)
	@echo "Built: $(BUILD_DIR)/$(BINARY)"

## install: Install vibe-gen to $GOPATH/bin (makes it available system-wide)
install:
	go install $(CMD)
	@echo "Installed: $$(which $(BINARY))"

## clean: Remove build output
clean:
	@rm -rf $(BUILD_DIR)

## test: Run all tests
test:
	go test ./...

## run-tui: Run the interactive TUI
run-tui:
	go run $(CMD)
