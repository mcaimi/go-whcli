GOBIN := $(shell go env GOBIN)
ifeq ($(GOBIN),)
GOBIN := $(shell go env GOPATH)/bin
endif

export PATH := $(GOBIN):$(PATH)

all: whcli

whcli: goyacc
	go generate
	go build

.PHONY: goyacc
goyacc:
	@command -v goyacc >/dev/null 2>&1 || go install golang.org/x/tools/cmd/goyacc@latest

clean:
	rm -f y.output cmd/whcli/grammar.go go-whcli
