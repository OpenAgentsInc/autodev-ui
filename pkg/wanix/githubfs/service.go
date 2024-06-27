package githubfs

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type GitHubFSService struct {
	fs    *FS
	owner string
	repo  string
	token string
}

func NewGitHubFSService(repo string) (*GitHubFSService, error) {
	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid repository format: %s", repo)
	}

	owner, repoName := parts[0], parts[1]
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("GITHUB_TOKEN environment variable is not set")
	}

	fs := New(owner, repoName, token)
	return &GitHubFSService{
		fs:    fs,
		owner: owner,
		repo:  repoName,
		token: token,
	}, nil
}

func (s *GitHubFSService) GetBranches() ([]string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/branches", s.owner, s.repo)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "token "+s.token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API request failed with status: %s", resp.Status)
	}

	var branches []struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&branches); err != nil {
		return nil, err
	}

	var branchNames []string
	for _, branch := range branches {
		branchNames = append(branchNames, branch.Name)
	}

	return branchNames, nil
}

func (s *GitHubFSService) GetFileCount(branch string) (int, error) {
	count := 0
	err := s.countFiles(branch, &count)
	return count, err
}

func (s *GitHubFSService) countFiles(path string, count *int) error {
	file, err := s.fs.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}

	if !fileInfo.IsDir() {
		*count++
		return nil
	}

	// If it's a directory, we need to list its contents
	dir, ok := file.(fs.ReadDirFile)
	if !ok {
		return fmt.Errorf("file is not a directory")
	}

	entries, err := dir.ReadDir(-1)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		fullPath := filepath.Join(path, entry.Name())
		if entry.IsDir() {
			if err := s.countFiles(fullPath, count); err != nil {
				return err
			}
		} else {
			*count++
		}
	}

	return nil
}

