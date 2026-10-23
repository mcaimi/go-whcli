GOBIN := $(shell go env GOBIN)
ifeq ($(GOBIN),)
GOBIN := $(shell go env GOPATH)/bin
endif

export PATH := $(GOBIN):$(PATH)

all: whcli build-gui cross-compile

.PHONY: generate
generate: goyacc
	go generate

whcli: generate
	go build

.PHONY: goyacc
goyacc:
	@command -v goyacc >/dev/null 2>&1 || go install golang.org/x/tools/cmd/goyacc@latest

GUI_BIN := whcli-gui

.PHONY: build-gui run-gui
build-gui:
	go build -o $(GUI_BIN) ./cmd/whcli-gui

run-gui: build-gui
	./$(GUI_BIN)

DIST_DIR := dist
CLI_BIN := whcli

.PHONY: cross-compile cross-linux-amd64 cross-windows-amd64 cross-darwin-arm64
cross-compile: cross-linux-amd64 cross-windows-amd64 cross-darwin-arm64

cross-linux-amd64: generate
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o $(DIST_DIR)/$(CLI_BIN)-linux-amd64

cross-windows-amd64: generate
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o $(DIST_DIR)/$(CLI_BIN)-windows-amd64.exe

cross-darwin-arm64: generate
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -o $(DIST_DIR)/$(CLI_BIN)-darwin-arm64

clean:
	rm -f y.output cmd/whcli/grammar.go go-whcli whcli-gui
	rm -rf $(DIST_DIR)
