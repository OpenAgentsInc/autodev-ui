package main

import (
	"github.com/openagentsinc/autodev/config"
	"github.com/openagentsinc/autodev/plugins"
	"github.com/openagentsinc/autodev/server"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}

	plugin, err := plugins.InitializePlugin()
	if err != nil {
		panic(err)
	}
	defer plugin.Close()

	e := server.SetupServer(cfg, plugin)
	e.Logger.Fatal(e.Start(":8080"))
}
