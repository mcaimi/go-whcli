GOBIN := $(shell go env GOBIN)
ifeq ($(GOBIN),)
GOBIN := $(shell go env GOPATH)/bin
endif

export PATH := $(GOBIN):$(PATH)

all: whcli build-gui

whcli: goyacc
	go generate
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

clean:
	rm -f y.output cmd/whcli/grammar.go go-whcli whcli-gui
