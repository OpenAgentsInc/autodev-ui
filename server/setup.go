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
	"github.com/openagentsinc/autodev/pkg/wanix/githubfs"
	"github.com/openagentsinc/autodev/plugins"
	"github.com/openagentsinc/autodev/views"
)

func SetupServer(cfg *config.Config, extismPlugin *extism.Plugin) *echo.Echo {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Renderer = &TemplRenderer{}
	e.Static("/static", "static")

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

	e.GET("/explorer/:repo", func(c echo.Context) error {
		repo := c.Param("repo")
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

	e.GET("/explorer/:repo/list", func(c echo.Context) error {
		repo := c.Param("repo")
		branch := c.QueryParam("branch")
		path := c.QueryParam("path")
		service, err := githubfs.NewGitHubFSService(repo)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		entries, err := service.ListDirectory(branch, path)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return c.Render(http.StatusOK, "directory_list", map[string]interface{}{
			"Entries": entries,
			"Path":    path,
			"Branch":  branch,
			"Repo":    repo,
		})
	})

	e.GET("/explorer/:repo/file", func(c echo.Context) error {
		repo := c.Param("repo")
		branch := c.QueryParam("branch")
		path := c.QueryParam("path")
		service, err := githubfs.NewGitHubFSService(repo)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		content, err := service.GetFileContent(branch, path)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return c.Render(http.StatusOK, "file_content", map[string]interface{}{
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

	switch name {
	case "index":
		return views.Index(cssVersion, viewContext).Render(context.Background(), w)
	case "greptile":
		return views.Greptile(cssVersion).Render(context.Background(), w)
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
