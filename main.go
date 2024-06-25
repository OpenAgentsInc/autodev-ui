package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/extism/go-sdk"
	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()

	// Serve static files
	e.Static("/static", "static")

	// Load the Extism plugin
	manifest := extism.Manifest{
		Wasm: []extism.Wasm{
			extism.WasmFile{
				Path: "plugins/wasm/greptile.wasm",
			},
		},
	}

	ctx := context.Background()
	config := extism.PluginConfig{}
	plugin, err := extism.NewPlugin(ctx, manifest, config, []extism.HostFunction{})
	if err != nil {
		e.Logger.Fatal(err)
	}

	// Define routes
	e.GET("/", func(c echo.Context) error {
		return c.Render(http.StatusOK, "index.html", nil)
	})

	e.POST("/run-plugin", func(c echo.Context) error {
		operation := c.FormValue("operation")
		input := c.FormValue("input")

		// Prepare input for the plugin
		pluginInput := fmt.Sprintf(`{"operation":"%s","repository":"%s"}`, operation, input)

		// Call the plugin
		exit, out, err := plugin.Call("run", []byte(pluginInput))
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		if exit != 0 {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Plugin exited with non-zero status"})
		}

		return c.JSON(http.StatusOK, map[string]string{"result": string(out)})
	})

	// Start the server
	e.Logger.Fatal(e.Start(":8080"))
}
