package server

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"time"
    "strings"

	"github.com/extism/go-sdk"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/openagentsinc/autodev/agent"
	"github.com/openagentsinc/autodev/config"
	"github.com/openagentsinc/autodev/pkg/wanix/githubfs"
	"github.com/openagentsinc/autodev/plugins"
	"github.com/openagentsinc/autodev/views"
    "github.com/openagentsinc/autodev/llm"
)

func SetupServer(cfg *config.Config, extismPlugin *extism.Plugin) *echo.Echo {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Renderer = &TemplRenderer{}
	e.Static("/static", "static")

	cssVersion := fmt.Sprintf("v=%d", time.Now().Unix())

    llmClient := llm.NewLLM("", "")

	// Create a new agent with a hardcoded plan
	initialPlan := agent.NewPlan(
		"We are cloning OpenDevin, a web UI for managing semi-autonomous AI coding agents that implements the CodeAct paper. Their codebase is in Python and we are converting it to Golang.",
		[]*agent.Task{},
	)
	myAgent := agent.NewAgent(initialPlan)

	e.GET("/", func(c echo.Context) error {
		return c.Render(http.StatusOK, "index", map[string]interface{}{
			"CssVersion": cssVersion,
			"Agent":      myAgent,
		})
	})

    e.POST("/submit-message", func(c echo.Context) error {
        message := c.FormValue("message")

		return c.HTML(http.StatusOK, "<h1>NICE</h1>")
    })

    e.POST("/submit-message2", func(c echo.Context) error {
		message := c.FormValue("message")

		// Get conversation history from the agent
		conversationHistory := myAgent.GetConversationHistory()

		// Add the new user message
		conversationHistory = append(conversationHistory, llm.Message{
			Role:    "user",
			Content: message,
		})

		// Generate response using LLM
		response, err := cfg.LLM.GenerateResponse(conversationHistory, 1024)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		// Add the assistant's response to the conversation history
		conversationHistory = append(conversationHistory, llm.Message{
			Role:    "assistant",
			Content: response,
		})

		// Update the agent's conversation history
		myAgent.SetConversationHistory(conversationHistory)

		// Update the plan based on the LLM response
		// This is a placeholder - you might want to implement more sophisticated plan updating logic
		newTask := &agent.Task{
			ID:    fmt.Sprintf("%d", len(myAgent.GetPlan().Tasks) + 1),
			Goal:  response,
			State: "open",
		}
		myAgent.GetPlan().Tasks = append(myAgent.GetPlan().Tasks, newTask)

		// Generate the updated plan HTML
		planHTML := generatePlanHTML(myAgent.GetPlan())

		// Return both the message, response, and the updated plan
		htmlResponse := fmt.Sprintf(`
			<div class="bg-zinc-900 rounded p-3 inline-block mb-2">%s</div>
			<div class="bg-zinc-800 rounded p-3 inline-block">%s</div>
			<div id="plan-display" hx-swap-oob="true">%s</div>
		`, message, response, planHTML)

		return c.HTML(http.StatusOK, htmlResponse)
	});

	e.POST("/replay", func(c echo.Context) error {
		// Clear existing tasks and generate new plan
		myAgent.ResetPlan()
		return c.NoContent(http.StatusOK)
	})

	e.GET("/plan-updates", func(c echo.Context) error {
		c.Response().Header().Set(echo.HeaderContentType, "text/event-stream")
		c.Response().Header().Set(echo.HeaderCacheControl, "no-cache")
		c.Response().Header().Set(echo.HeaderConnection, "keep-alive")
		c.Response().WriteHeader(http.StatusOK)

		updates := myAgent.GenerateDemoPlan()
		for update := range updates {
			taskListHTML := generateTaskListHTML(update)

			if _, err := c.Response().Write([]byte(fmt.Sprintf("data: %s\n\n", taskListHTML))); err != nil {
				return err
			}
			c.Response().Flush()
		}

		return nil
	})

	e.GET("/repos", func(c echo.Context) error {
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

		return c.Render(http.StatusOK, "repos", data)
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

	if !ok {
		return fmt.Errorf("invalid data type for rendering")
	}

	cssVersion, _ := viewContext["CssVersion"].(string)
	myAgent, _ := viewContext["Agent"].(*agent.Agent)

	switch name {
	case "index":
		return views.Index(cssVersion, myAgent).Render(context.Background(), w)
	case "repos":
		return views.Repos(cssVersion, viewContext).Render(context.Background(), w)
	case "greptile":
		return views.Greptile(cssVersion).Render(context.Background(), w)
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

func generateTaskListHTML(update agent.PlanUpdate) string {
	return fmt.Sprintf(`<li class="mb-2"><span class="text-blue-400 mr-2">%s.</span><span>%s</span><span class="ml-2 text-yellow-400">(%s)</span>
</li>`, update.TaskID, update.Goal, update.State)
}

func generatePlanHTML(plan *agent.Plan) string {
	var taskListHTML strings.Builder
	for _, task := range plan.Tasks {
		stateClass := "text-yellow-400"
		switch task.State {
		case "completed":
			stateClass = "text-green-400"
		case "in_progress":
			stateClass = "text-blue-400"
		}
		taskListHTML.WriteString(fmt.Sprintf(`
			<li class="flex justify-between items-center mb-2">
				<span class="text-white">%s</span>
				<span class="%s">%s</span>
			</li>
		`, task.Goal, stateClass, task.State))
	}
	
	return fmt.Sprintf(`
		<h3 class="text-lg font-bold mb-2">Main Goal:</h3>
		<p class="mb-4 text-white">%s</p>
		<h3 class="text-lg font-bold mb-2">Tasks:</h3>
		<ul class="list-none pl-0 space-y-2">
			%s
		</ul>
	`, plan.MainGoal, taskListHTML.String())
}
