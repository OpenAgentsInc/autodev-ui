package services

import (
	"fmt"
	"github.com/openagentsinc/autodev/pkg/wanix/githubfs"
	"os"
	"strings"
)

type GitHubFSService struct {
	fs *githubfs.FS
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

	fs := githubfs.New(owner, repoName, token)
	return &GitHubFSService{fs: fs}, nil
}

func (s *GitHubFSService) GetBranches() ([]string, error) {
	entries, err := s.fs.ReadDir(".")
	if err != nil {
		return nil, err
	}

	var branches []string
	for _, entry := range entries {
		if entry.IsDir() {
			branches = append(branches, entry.Name())
		}
	}
	return branches, nil
}

func (s *GitHubFSService) GetFileCount(branch string) (int, error) {
	var count int
	err := s.countFiles(branch, &count)
	return count, err
}

func (s *GitHubFSService) countFiles(path string, count *int) error {
	entries, err := s.fs.ReadDir(path)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			*count++
		} else {
			err = s.countFiles(path+"/"+entry.Name(), count)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
