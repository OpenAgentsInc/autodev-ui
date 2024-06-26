package main

import (
	"context"

	"github.com/OpenAgentsInc/autodev/config"
	"github.com/OpenAgentsInc/autodev/server"
	"github.com/extism/go-sdk"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}

	plugin, err := initializePlugin()
	if err != nil {
		panic(err)
	}
	defer plugin.Close()

	e := server.SetupServer(cfg, plugin)
	e.Logger.Fatal(e.Start(":8080"))
}

func initializePlugin() (*extism.Plugin, error) {
	manifest := extism.Manifest{
		Wasm: []extism.Wasm{
			extism.WasmFile{
				Path: "plugins/wasm/greptile.wasm",
			},
		},
		AllowedHosts: []string{"api.greptile.com"},
	}

	config := extism.PluginConfig{
		EnableWasi: true,
	}

	ctx := context.Background()
	return extism.NewPlugin(ctx, manifest, config, []extism.HostFunction{})
}
