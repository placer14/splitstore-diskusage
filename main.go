package main

import (
	"context"

	flags "github.com/jessevdk/go-flags"
	"github.com/placer14/splitstore-diskusage/splitstore"
)

var o splitstore.AgentOptions

func init() {
	_, err := flags.Parse(&o)
	if err != nil {
		panic(err)
	}
}

func main() {
	agent := splitstore.NewDiskUsageAgent(o)
	agent.Start(context.Background())
}
