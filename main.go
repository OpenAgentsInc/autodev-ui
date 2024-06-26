package main

import (
	"github.com/OpenAgentsInc/autodev/config"
	"github.com/OpenAgentsInc/autodev/plugins"
	"github.com/OpenAgentsInc/autodev/server"
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

