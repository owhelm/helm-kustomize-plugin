.PHONY: build clean test install uninstall

BINARY_NAME=helm-kustomize-plugin
BUILD_DIR=.

build:
	go build -o $(BUILD_DIR)/$(BINARY_NAME) .

clean:
	rm -f $(BUILD_DIR)/$(BINARY_NAME)
	go clean

test:
	go test -v ./...

install: build
	helm plugin install .

uninstall:
	helm plugin uninstall kustomize

# Development: uninstall, rebuild, and reinstall
reinstall: uninstall build install