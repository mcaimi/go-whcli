
all: whcli

whcli:
	go generate
	go build

clean:
	rm y.output cmd/whcli/grammar.go go-whcli
