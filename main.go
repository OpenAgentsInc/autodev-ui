package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/OpenAgentsInc/autodev/views"
	"github.com/extism/go-sdk"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type TemplRenderer struct{}

func (t *TemplRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return views.Index().Render(context.Background(), w)
}

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Renderer = &TemplRenderer{}
	e.Static("/static", "static")

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

	e.GET("/", func(c echo.Context) error {
		return c.Render(http.StatusOK, "index", nil)
	})

	e.POST("/run-plugin", func(c echo.Context) error {
		operation := c.FormValue("operation")
		repository := c.FormValue("repository")
		query := c.FormValue("query")

		apiKey := os.Getenv("GREPTILE_API_KEY")
		githubToken := os.Getenv("GITHUB_TOKEN")

		if apiKey == "" || githubToken == "" {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "API key or GitHub token not set"})
		}

		var pluginInput map[string]interface{}

		switch operation {
		case "index":
			pluginInput = map[string]interface{}{
				"operation":    operation,
				"repository":   repository,
				"remote":       "github",
				"branch":       "main",
				"api_key":      apiKey,
				"github_token": githubToken,
			}
		case "query":
			pluginInput = map[string]interface{}{
				"operation":    operation,
				"repository":   repository,
				"remote":       "github",
				"branch":       "main",
				"api_key":      apiKey,
				"github_token": githubToken,
				"messages": []map[string]string{
					{
						"id":      "1",
						"content": query,
						"role":    "user",
					},
				},
				"session_id": fmt.Sprintf("session-%d", time.Now().Unix()),
				"stream":     false,
				"genius":     true,
			}
		case "search":
			pluginInput = map[string]interface{}{
				"operation":    operation,
				"repository":   repository,
				"remote":       "github",
				"branch":       "main",
				"api_key":      apiKey,
				"github_token": githubToken,
				"query":        query,
				"session_id":   fmt.Sprintf("session-%d", time.Now().Unix()),
				"stream":       false,
			}
		default:
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid operation"})
		}

		pluginInputJSON, err := json.Marshal(pluginInput)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to prepare plugin input"})
		}

		exitCode, out, err := plugin.Call("run", pluginInputJSON)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		if exitCode != 0 {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Plugin exited with code %d", exitCode)})
		}

		return c.JSON(http.StatusOK, map[string]string{"result": string(out)})
	})

	e.Logger.Fatal(e.Start(":8080"))
}
