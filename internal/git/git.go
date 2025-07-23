package git

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// GitCommit represents a single git commit
type GitCommit struct {
	Hash      string
	Author    string
	Email     string
	Date      time.Time
	Subject   string
	Body      string
	Files     []string
	Additions int
	Deletions int
}

// Repository represents a git repository
type Repository struct {
	Path string
}

// NewRepository creates a new Repository instance
func NewRepository(path string) *Repository {
	return &Repository{Path: path}
}

// IsGitRepository checks if the given path is a valid git repository
func (r *Repository) IsGitRepository() bool {
	gitDir := filepath.Join(r.Path, ".git")
	if _, err := os.Stat(gitDir); err == nil {
		return true
	}

	// Check if it's inside a git repository
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Dir = r.Path
	err := cmd.Run()
	return err == nil
}

// IsGitInstalled checks if git command is available
func IsGitInstalled() bool {
	cmd := exec.Command("git", "--version")
	err := cmd.Run()
	return err == nil
}

// GetCommits retrieves git commits from the repository
func (r *Repository) GetCommits(limit int) ([]GitCommit, error) {
	if !IsGitInstalled() {
		return nil, fmt.Errorf("git is not installed or not available in PATH")
	}

	if !r.IsGitRepository() {
		return nil, fmt.Errorf("not a git repository: %s", r.Path)
	}

	// Git log format: hash|author|email|date|subject|body
	format := "--pretty=format:%H|%an|%ae|%ai|%s|%b"
	args := []string{"log", format}
	if limit > 0 {
		args = append(args, fmt.Sprintf("-%d", limit))
	}

	cmd := exec.Command("git", args...)
	cmd.Dir = r.Path
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get git log: %v", err)
	}

	return parseCommits(string(output))
}

// parseCommits parses git log output into GitCommit structs
func parseCommits(output string) ([]GitCommit, error) {
	var commits []GitCommit
	lines := strings.Split(output, "\n")
	
	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, "|", 6)
		if len(parts) < 5 {
			continue
		}

		date, err := time.Parse("2006-01-02 15:04:05 -0700", parts[3])
		if err != nil {
			// Try alternative format
			date, err = time.Parse("2006-01-02T15:04:05-07:00", parts[3])
			if err != nil {
				continue
			}
		}

		body := ""
		if len(parts) > 5 {
			body = parts[5]
		}

		commit := GitCommit{
			Hash:    parts[0],
			Author:  parts[1],
			Email:   parts[2],
			Date:    date,
			Subject: parts[4],
			Body:    body,
		}

		commits = append(commits, commit)
	}

	return commits, nil
}

// GetCommitStats gets detailed statistics for a commit
func (r *Repository) GetCommitStats(hash string) (int, int, []string, error) {
	cmd := exec.Command("git", "show", "--stat", "--format=", hash)
	cmd.Dir = r.Path
	output, err := cmd.Output()
	if err != nil {
		return 0, 0, nil, err
	}

	var files []string
	additions := 0
	deletions := 0

	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Parse file changes
		if strings.Contains(line, "|") {
			parts := strings.Split(line, "|")
			if len(parts) >= 2 {
				filename := strings.TrimSpace(parts[0])
				files = append(files, filename)
			}
		}

		// Parse summary line (e.g., "2 files changed, 15 insertions(+), 3 deletions(-)")
		if strings.Contains(line, "insertion") || strings.Contains(line, "deletion") {
			// Extract numbers from the summary line
			words := strings.Fields(line)
			for i, word := range words {
				if strings.Contains(word, "insertion") && i > 0 {
					fmt.Sscanf(words[i-1], "%d", &additions)
				}
				if strings.Contains(word, "deletion") && i > 0 {
					fmt.Sscanf(words[i-1], "%d", &deletions)
				}
			}
		}
	}

	return additions, deletions, files, nil
}
