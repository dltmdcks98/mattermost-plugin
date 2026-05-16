GO ?= go
PLUGIN_ID := dltmdcks98.hermes-slash-plugin
BUNDLE_DIR := dist/$(PLUGIN_ID)
BUNDLE_FILE := dist/$(PLUGIN_ID).tar.gz

.PHONY: test build bundle clean

test:
	$(GO) test ./...

build:
	mkdir -p server/dist
	GOOS=darwin GOARCH=amd64 $(GO) build -o server/dist/plugin-darwin-amd64 ./server
	GOOS=darwin GOARCH=arm64 $(GO) build -o server/dist/plugin-darwin-arm64 ./server
	GOOS=linux GOARCH=amd64 $(GO) build -o server/dist/plugin-linux-amd64 ./server

bundle: clean build
	mkdir -p $(BUNDLE_DIR)
	cp plugin.json $(BUNDLE_DIR)/plugin.json
	mkdir -p $(BUNDLE_DIR)/assets $(BUNDLE_DIR)/server/dist
	cp assets/icon.svg $(BUNDLE_DIR)/assets/icon.svg
	cp server/dist/plugin-* $(BUNDLE_DIR)/server/dist/
	tar -C $(BUNDLE_DIR) -czf $(BUNDLE_FILE) plugin.json assets server
	@echo $(BUNDLE_FILE)

clean:
	rm -rf dist server/dist
