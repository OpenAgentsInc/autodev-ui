package githubfs

import (
	"encoding/json"
	"fmt"
	"io"
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

func (s *GitHubFSService) ListDirectory(branch, path string) ([]fs.FileInfo, error) {
	fullPath := filepath.Join(branch, path)
	fmt.Printf("Attempting to open directory: %s\n", fullPath) // Debug print

	file, err := s.fs.Open(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open directory %s: %w", fullPath, err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file info for %s: %w", fullPath, err)
	}

	if !fileInfo.IsDir() {
		return nil, fmt.Errorf("path is not a directory: %s", fullPath)
	}

	dir, ok := file.(fs.ReadDirFile)
	if !ok {
		return nil, fmt.Errorf("file does not implement fs.ReadDirFile: %s", fullPath)
	}

	entries, err := dir.ReadDir(-1)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", fullPath, err)
	}

	var fileInfos []fs.FileInfo
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			return nil, fmt.Errorf("failed to get info for entry in %s: %w", fullPath, err)
		}
		fileInfos = append(fileInfos, info)
	}

	return fileInfos, nil
}

func (s *GitHubFSService) GetFileContent(branch, path string) (string, error) {
	fullPath := filepath.Join(branch, path)
	file, err := s.fs.Open(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return "", fmt.Errorf("failed to get file info: %w", err)
	}

	if fileInfo.IsDir() {
		return "", fmt.Errorf("path is a directory, not a file")
	}

	content, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("failed to read file content: %w", err)
	}

	return string(content), nil
}
