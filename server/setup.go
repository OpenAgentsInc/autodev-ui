package server

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"time"

	"github.com/extism/go-sdk"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/openagentsinc/autodev/config"
	"github.com/openagentsinc/autodev/pkg/action"
	"github.com/openagentsinc/autodev/pkg/agent"
	"github.com/openagentsinc/autodev/pkg/plugin"
	"github.com/openagentsinc/autodev/pkg/state"
	"github.com/openagentsinc/autodev/pkg/wanix/githubfs"
	"github.com/openagentsinc/autodev/plugins"
	"github.com/openagentsinc/autodev/views"
)

type ConversationEntry struct {
	Role    string
	Content string
}

type AgentController struct {
	agent   func() agent.Agent
	Actions action.ActionManager
}

func (ac *AgentController) ActionManager() action.ActionManager {
	return ac.Actions
}

func (ac *AgentController) Agent() agent.Agent {
	return ac.agent()
}

func SetupServer(cfg *config.Config, extismPlugin *extism.Plugin) *echo.Echo {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Renderer = &TemplRenderer{}
	e.Static("/static", "static")

	conversation := make([]ConversationEntry, 0)
	pluginReqs := []plugin.PluginRequirement{
		// Add your plugin requirements here, for example:
		// {Name: "example-plugin", Version: "1.0.0"},
	}
	agentCtrl := &AgentController{
		agent: func() agent.Agent {
			return agent.NewBaseAgent(cfg.LLM, pluginReqs)
		},
		Actions: action.NewActionManager(), // Make sure this function exists in the action package
	}

	e.GET("/agent/test", func(c echo.Context) error {
		return c.Render(http.StatusOK, "agent_test", conversation)
	})

	e.POST("/agent/interact", func(c echo.Context) error {
		message := c.FormValue("message")
		conversation = append(conversation, ConversationEntry{Role: "user", Content: message})

		// Create a new state for each interaction
		state := state.NewState(nil)

		// Create a new state for each interaction
		// state := state.NewState(message)

		// Get the next action from the agent
		nextAction := agentCtrl.Agent.Step(state)

		// 0] server/setup.go:73:30: cannot use agentCtrl (variable of type *AgentController) as action.AgentController value in argument to nextAction.Run: *AgentController does not implement action.AgentController (*AgentController.Agent is a field, not a method)
		// Execute the action
		obs, err := nextAction.Run(agentCtrl)
		// if err != nil {
		// return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		// }

		// Add the agent's response to the conversation
		// agentResponse := obs.GetContent()
		// conversation = append(conversation, ConversationEntry{Role: "agent", Content: agentResponse})
		agentResponse := "Hello, I am an agent. How can I help you?"

		return c.Render(http.StatusOK, "conversation_update", ConversationEntry{Role: "agent", Content: agentResponse})
	})

	cssVersion := fmt.Sprintf("v=%d", time.Now().Unix())

	e.GET("/", func(c echo.Context) error {
		repo := c.QueryParam("repo")
		data := map[string]interface{}{
			"CssVersion": cssVersion,
			"Repo":       repo,
		}

		if repo != "" {
			service, err := githubfs.NewGitHubFSService(repo)
			if err != nil {
				data["Error"] = fmt.Sprintf("Failed to initialize GitHubFS: %v", err)
			} else {
				branches, err := service.GetBranches()
				if err != nil {
					data["Error"] = fmt.Sprintf("Failed to get branches: %v", err)
				} else {
					data["Branches"] = branches
					totalFiles := 0
					branchFileCounts := make(map[string]int)
					for _, branch := range branches {
						fileCount, err := service.GetFileCount(branch)
						if err != nil {
							data["Error"] = fmt.Sprintf("Failed to get file count for branch %s: %v", branch, err)
							break
						}
						totalFiles += fileCount
						branchFileCounts[branch] = fileCount
					}
					data["TotalFiles"] = totalFiles
					data["BranchFileCounts"] = branchFileCounts
				}
			}
		}

		return c.Render(http.StatusOK, "index", data)
	})

	e.GET("/explorer", func(c echo.Context) error {
		repo := c.QueryParam("repo")
		if repo == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Repository not specified"})
		}
		service, err := githubfs.NewGitHubFSService(repo)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		branches, err := service.GetBranches()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return c.Render(http.StatusOK, "explorer", map[string]interface{}{
			"CssVersion": cssVersion,
			"Repo":       repo,
			"Branches":   branches,
		})
	})

	e.GET("/explorer/list", func(c echo.Context) error {
		repo := c.QueryParam("repo")
		branch := c.QueryParam("branch")
		path := c.QueryParam("path")

		c.Logger().Infof("Listing directory: repo=%s, branch=%s, path=%s", repo, branch, path)

		service, err := githubfs.NewGitHubFSService(repo)
		if err != nil {
			c.Logger().Errorf("Failed to create GitHubFSService: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		entries, err := service.ListDirectory(branch, path)
		if err != nil {
			c.Logger().Errorf("Failed to list directory: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		c.Logger().Infof("Found %d entries", len(entries))

		return c.Render(http.StatusOK, "directory_list", map[string]interface{}{
			"Entries": entries,
			"Path":    path,
			"Branch":  branch,
			"Repo":    repo,
		})
	})

	e.GET("/explorer/file", func(c echo.Context) error {
		repo := c.QueryParam("repo")
		branch := c.QueryParam("branch")
		path := c.QueryParam("path")

		c.Logger().Infof("Fetching file content: repo=%s, branch=%s, path=%s", repo, branch, path)

		service, err := githubfs.NewGitHubFSService(repo)
		if err != nil {
			c.Logger().Errorf("Failed to create GitHubFSService: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		content, err := service.GetFileContent(branch, path)
		if err != nil {
			c.Logger().Errorf("Failed to get file content: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		c.Logger().Infof("Successfully fetched file content (length: %d)", len(content))

		return c.Render(http.StatusOK, "file_content", map[string]interface{}{
			"Content": content,
			"Path":    path,
		})
	})

	e.GET("/widget/explorer", func(c echo.Context) error {
		repo := c.QueryParam("repo")
		if repo == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Repository not specified"})
		}
		service, err := githubfs.NewGitHubFSService(repo)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		branches, err := service.GetBranches()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		currentBranch := ""

		// Try to find "main" branch
		for _, branch := range branches {
			if branch == "main" {
				currentBranch = "main"
				break
			}
		}

		// If "main" not found, try to find "master" branch
		if currentBranch == "" {
			for _, branch := range branches {
				if branch == "master" {
					currentBranch = "master"
					break
				}
			}
		}

		// If neither "main" nor "master" found, use the first branch
		if currentBranch == "" && len(branches) > 0 {
			currentBranch = branches[0]
		}

		return c.Render(http.StatusOK, "file_explorer_widget", map[string]interface{}{
			"Repo":          repo,
			"Branches":      branches,
			"CurrentBranch": currentBranch,
		})
	})

	e.GET("/widget/explorer/list", func(c echo.Context) error {
		repo := c.QueryParam("repo")
		branch := c.QueryParam("branch")
		path := c.QueryParam("path")

		c.Logger().Infof("Listing directory: repo=%s, branch=%s, path=%s", repo, branch, path)

		service, err := githubfs.NewGitHubFSService(repo)
		if err != nil {
			c.Logger().Errorf("Failed to create GitHubFSService: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		entries, err := service.ListDirectory(branch, path)
		if err != nil {
			c.Logger().Errorf("Failed to list directory: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		c.Logger().Infof("Found %d entries", len(entries))

		return c.Render(http.StatusOK, "widget_directory_list", map[string]interface{}{
			"Entries": entries,
			"Path":    path,
			"Branch":  branch,
			"Repo":    repo,
		})
	})

	e.GET("/widget/explorer/file", func(c echo.Context) error {
		repo := c.QueryParam("repo")
		branch := c.QueryParam("branch")
		path := c.QueryParam("path")

		c.Logger().Infof("Fetching file content: repo=%s, branch=%s, path=%s", repo, branch, path)

		service, err := githubfs.NewGitHubFSService(repo)
		if err != nil {
			c.Logger().Errorf("Failed to create GitHubFSService: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		content, err := service.GetFileContent(branch, path)
		if err != nil {
			c.Logger().Errorf("Failed to get file content: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		c.Logger().Infof("Successfully fetched file content (length: %d)", len(content))

		return c.Render(http.StatusOK, "widget_file_content", map[string]interface{}{
			"Content": content,
			"Path":    path,
		})
	})

	e.GET("/greptile", func(c echo.Context) error {
		return c.Render(http.StatusOK, "greptile", map[string]interface{}{
			"CssVersion": cssVersion,
		})
	})

	e.POST("/run-plugin", func(c echo.Context) error {
		input := plugins.PluginInput{
			Operation:   c.FormValue("operation"),
			Repository:  c.FormValue("repository"),
			Query:       c.FormValue("query"),
			ApiKey:      cfg.GreptileApiKey,
			GithubToken: cfg.GithubToken,
		}

		pluginInputJSON, err := plugins.PreparePluginInput(input)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}

		result, err := plugins.CallPlugin(extismPlugin, pluginInputJSON)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		return c.JSON(http.StatusOK, result)
	})

	return e
}

type TemplRenderer struct{}

func (t *TemplRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	viewContext, ok := data.(map[string]interface{})
	ctx := context.Background()

	if !ok {
		return fmt.Errorf("invalid data type for rendering")
	}

	cssVersion, _ := viewContext["CssVersion"].(string)

	switch name {
	case "index":
		return views.Index(cssVersion, viewContext).Render(context.Background(), w)
	case "greptile":
		return views.Greptile(cssVersion).Render(context.Background(), w)
	case "agent_test":
		conversation, ok := data.([]ConversationEntry)
		if !ok {
			return fmt.Errorf("unexpected data type for agent_test template")
		}
		return views.AgentTest(conversation).Render(ctx, w)
	case "conversation_update":
		entry, ok := data.(ConversationEntry)
		if !ok {
			return fmt.Errorf("unexpected data type for conversation_update template")
		}
		return views.ConversationUpdate(entry).Render(ctx, w)
	case "file_explorer_widget":
		repo, _ := viewContext["Repo"].(string)
		branches, _ := viewContext["Branches"].([]string)
		currentBranch, _ := viewContext["CurrentBranch"].(string)
		return views.FileExplorerWidget(repo, branches, currentBranch).Render(context.Background(), w)
	case "widget_directory_list":
		entries, _ := viewContext["Entries"].([]fs.FileInfo)
		path, _ := viewContext["Path"].(string)
		branch, _ := viewContext["Branch"].(string)
		repo, _ := viewContext["Repo"].(string)
		return views.WidgetDirectoryList(entries, path, branch, repo).Render(context.Background(), w)
	case "widget_file_content":
		content, _ := viewContext["Content"].(string)
		path, _ := viewContext["Path"].(string)
		return views.WidgetFileContent(content, path).Render(context.Background(), w)
	case "explorer":
		repo, _ := viewContext["Repo"].(string)
		branches, _ := viewContext["Branches"].([]string)
		return views.Explorer(cssVersion, repo, branches).Render(context.Background(), w)
	case "directory_list":
		entries, _ := viewContext["Entries"].([]fs.FileInfo)
		path, _ := viewContext["Path"].(string)
		branch, _ := viewContext["Branch"].(string)
		repo, _ := viewContext["Repo"].(string)
		return views.DirectoryList(entries, path, branch, repo).Render(context.Background(), w)
	case "file_content":
		content, _ := viewContext["Content"].(string)
		path, _ := viewContext["Path"].(string)
		return views.FileContent(content, path).Render(context.Background(), w)
	default:
		return fmt.Errorf("unknown template: %s", name)
	}
}
