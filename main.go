package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/OpenAgentsInc/autodev/views"
	"github.com/extism/go-sdk"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type TemplRenderer struct{}

func (t *TemplRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	if viewContext, ok := data.(map[string]interface{}); ok {
		if cssVersion, ok := viewContext["CssVersion"].(string); ok {
			return views.Index(cssVersion).Render(context.Background(), w)
		}
	}
	return views.Index("").Render(context.Background(), w)
}

var cssVersion string

func init() {
	cssVersion = fmt.Sprintf("v=%d", time.Now().Unix())
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
		return c.Render(http.StatusOK, "index", map[string]interface{}{
			"CssVersion": cssVersion,
		})
	})

	e.POST("/run-plugin", func(c echo.Context) error {
		operation := c.FormValue("operation")
		repositories := c.Request().Form["repositories"]
		branch := c.FormValue("branch")
		if branch == "" {
			branch = "main"
		}
		query := c.FormValue("query")

		apiKey := os.Getenv("GREPTILE_API_KEY")
		githubToken := os.Getenv("GITHUB_TOKEN")
		if apiKey == "" || githubToken == "" {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "API key or GitHub token not set"})
		}

		var results []string
		for _, repository := range repositories {
			var pluginInput map[string]interface{}
			switch operation {
			case "index":
				pluginInput = map[string]interface{}{
					"operation":    operation,
					"repository":   repository,
					"remote":       "github",
					"branch":       branch,
					"api_key":      apiKey,
					"github_token": githubToken,
				}
			case "query":
				pluginInput = map[string]interface{}{
					"operation":    operation,
					"repository":   repository,
					"remote":       "github",
					"branch":       branch,
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
					"branch":       branch,
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
				results = append(results, fmt.Sprintf("Error for %s: %s", repository, err.Error()))
				continue
			}
			if exitCode != 0 {
				results = append(results, fmt.Sprintf("Plugin exited with code %d for %s", exitCode, repository))
				continue
			}

			bodyIndex := strings.Index(string(out), "Body: ")
			if bodyIndex == -1 {
				results = append(results, fmt.Sprintf("Unexpected response format for %s", repository))
				continue
			}
			jsonContent := string(out)[bodyIndex+6:]

			var response map[string]interface{}
			if err := json.Unmarshal([]byte(jsonContent), &response); err != nil {
				results = append(results, fmt.Sprintf("Failed to parse response for %s: %s", repository, err.Error()))
				continue
			}

			if message, ok := response["message"].(string); ok {
				results = append(results, fmt.Sprintf("Summary for %s:\n%s", repository, message))
			} else {
				results = append(results, fmt.Sprintf("No summary available for %s", repository))
			}
		}

		combinedResult := strings.Join(results, "\n\n")
		return c.HTML(http.StatusOK, "<pre>"+combinedResult+"</pre>")
	})

	e.Logger.Fatal(e.Start(":8080"))
}

