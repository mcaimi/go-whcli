package main

//go:generate goyacc -o cmd/whcli/grammar.go cmd/whcli/grammar.y

import (
  "codeberg.org/mcaimi/go-whcli/cmd/whcli"
)

func main() {
  whcli.Execute();
}
