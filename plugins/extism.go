package plugins

import (
	"context"

	"github.com/extism/go-sdk"
)

// InitializePlugin sets up and returns an Extism plugin for Greptile operations.
// It configures the plugin with the necessary WASM file and allowed hosts.
func InitializePlugin() (*extism.Plugin, error) {
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
