// Command remoteclient exercises the source API used by runtime download clients.
package main

import (
	"context"

	"github.com/bedrock-mc/protocolgen/data/source"
)

func main() {
	inputs, err := source.Open("")
	if err != nil {
		panic(err)
	}
	if _, err := inputs.ReadRemote(context.Background(), "missing", ""); err == nil {
		panic("unlocked input accepted")
	}
}
