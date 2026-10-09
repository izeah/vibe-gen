SHELL := /bin/sh
BINARY   := vibe-gen
BUILD_DIR := ./out
CMD      := ./cmd/vibe-gen

.PHONY: build install clean test smoke gen-js gen-js-check run-tui

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

## smoke: gen-js-check, then generate every ready template and build/lint/test/vuln-check it (FRAMEWORK=<id> for one)
smoke: gen-js-check
	FRAMEWORK=$(FRAMEWORK) go test -tags smoke -count=1 -timeout 60m -v ./internal/smoke

GEN_JS := cd scripts/gen-js && npm ci --ignore-scripts --no-audit --no-fund --silent && npm test --silent &&

## gen-js: Derive templates/javascript/<fw> from templates/typescript/<fw> (run after editing a TS template)
gen-js:
	$(GEN_JS) node gen-js.mjs

## gen-js-check: Fail if the committed JavaScript templates are out of sync with the TypeScript sources
gen-js-check:
	$(GEN_JS) node gen-js.mjs --check

## run-tui: Run the interactive TUI
run-tui:
	go run $(CMD)
