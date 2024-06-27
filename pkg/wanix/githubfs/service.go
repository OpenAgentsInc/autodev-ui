package githubfs

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type GitHubFSService struct {
	fs *FS
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
	return &GitHubFSService{fs: fs}, nil
}

func (s *GitHubFSService) GetBranches() ([]string, error) {
	// The root directory contains all branches as directories
	rootInfo, err := s.fs.Stat(".")
	if err != nil {
		return nil, err
	}

	rootDir, ok := rootInfo.(fs.ReadDirFile)
	if !ok {
		return nil, fmt.Errorf("root is not a directory")
	}

	entries, err := rootDir.ReadDir(-1)
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
	count := 0
	err := s.countFiles(branch, &count)
	return count, err
}

func (s *GitHubFSService) countFiles(path string, count *int) error {
	entries, err := s.readDir(path)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			*count++
		} else {
			err = s.countFiles(filepath.Join(path, entry.Name()), count)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *GitHubFSService) readDir(path string) ([]fs.DirEntry, error) {
	file, err := s.fs.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	dir, ok := file.(fs.ReadDirFile)
	if !ok {
		return nil, fmt.Errorf("%s is not a directory", path)
	}

	return dir.ReadDir(-1)
}

