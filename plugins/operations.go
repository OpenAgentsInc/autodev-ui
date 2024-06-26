package plugin

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/extism/go-sdk"
)

type PluginInput struct {
	Operation   string
	Repository  string
	Query       string
	ApiKey      string
	GithubToken string
}

func PreparePluginInput(input PluginInput) ([]byte, error) {
	var pluginInput map[string]interface{}

	switch input.Operation {
	case "index":
		pluginInput = map[string]interface{}{
			"operation":    input.Operation,
			"repository":   input.Repository,
			"remote":       "github",
			"branch":       "main",
			"api_key":      input.ApiKey,
			"github_token": input.GithubToken,
		}
	case "query":
		pluginInput = map[string]interface{}{
			"operation":    input.Operation,
			"repository":   input.Repository,
			"remote":       "github",
			"branch":       "main",
			"api_key":      input.ApiKey,
			"github_token": input.GithubToken,
			"messages": []map[string]string{
				{
					"id":      "1",
					"content": input.Query,
					"role":    "user",
				},
			},
			"session_id": fmt.Sprintf("session-%d", time.Now().Unix()),
			"stream":     false,
			"genius":     true,
		}
	case "search":
		pluginInput = map[string]interface{}{
			"operation":    input.Operation,
			"repository":   input.Repository,
			"remote":       "github",
			"branch":       "main",
			"api_key":      input.ApiKey,
			"github_token": input.GithubToken,
			"query":        input.Query,
			"session_id":   fmt.Sprintf("session-%d", time.Now().Unix()),
			"stream":       false,
		}
	default:
		return nil, fmt.Errorf("invalid operation: %s", input.Operation)
	}

	return json.Marshal(pluginInput)
}

func CallPlugin(plugin *extism.Plugin, input []byte) (map[string]interface{}, error) {
	exitCode, out, err := plugin.Call("run", input)
	if err != nil {
		return nil, fmt.Errorf("plugin call error: %v", err)
	}
	if exitCode != 0 {
		return nil, fmt.Errorf("plugin exited with code %d", exitCode)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(out, &result); err != nil {
		return nil, fmt.Errorf("failed to parse plugin output as JSON: %v", err)
	}

	return result, nil
}
