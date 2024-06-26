package main

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/OpenAgentsInc/autodev/views"
	"github.com/extism/go-sdk"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// Templ renderer
type TemplRenderer struct{}

func (t *TemplRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return views.Index().Render(context.Background(), w)
}

func main() {
	e := echo.New()

	// Add logging middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Set up Templ renderer
	e.Renderer = &TemplRenderer{}

	// Serve static files
	e.Static("/static", "static")

	// Load the Extism plugin
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
	plugin, err := extism.NewPlugin(ctx, manifest, config, []extism.HostFunction{})
	if err != nil {
		e.Logger.Fatal(err)
	}
	defer plugin.Close()

	// Define routes
	e.GET("/", func(c echo.Context) error {
		e.Logger.Info("Rendering index template")
		return c.Render(http.StatusOK, "index", nil)
	})

	e.POST("/run-plugin", func(c echo.Context) error {
		operation := c.FormValue("operation")
		input := c.FormValue("input")

		e.Logger.Infof("Running plugin with operation: %s, input: %s", operation, input)

		// Prepare input for the plugin
		pluginInput := fmt.Sprintf(`{"operation":"%s","repository":"%s"}`, operation, input)

		// Call the plugin
		exitCode, out, err := plugin.Call("run", []byte(pluginInput))
		if err != nil {
			e.Logger.Errorf("Plugin call error: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		if exitCode != 0 {
			e.Logger.Errorf("Plugin exited with non-zero code: %d", exitCode)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Plugin exited with code %d", exitCode)})
		}

		e.Logger.Infof("Plugin call successful, output: %s", string(out))
		return c.JSON(http.StatusOK, map[string]string{"result": string(out)})
	})

	// Start the server
	e.Logger.Fatal(e.Start(":8080"))
}
