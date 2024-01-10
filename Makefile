
all: whcli

whcli:
	go generate
	go build

clean:
	rm -f y.output cmd/whcli/grammar.go go-whcli
