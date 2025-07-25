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
	Message   string   // Full commit message (Subject + Body)
	Files     []string
	Additions int
	Deletions int
	Parents   []string // Parent commit hashes
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

	// Git log format: hash|author|email|date|subject|body|parents
	format := "--pretty=format:%H|%an|%ae|%ai|%s|%b|%P"
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

		parts := strings.SplitN(line, "|", 7)
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

		// Parse parent commits
		var parents []string
		if len(parts) > 6 && strings.TrimSpace(parts[6]) != "" {
			parentHashes := strings.Fields(strings.TrimSpace(parts[6]))
			parents = parentHashes
		}

		// Create full message
		message := parts[4]
		if body != "" {
			message = parts[4] + "\n\n" + body
		}

		commit := GitCommit{
			Hash:    parts[0],
			Author:  parts[1],
			Email:   parts[2],
			Date:    date,
			Subject: parts[4],
			Body:    body,
			Message: message,
			Parents: parents,
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

// GetBranches retrieves all branches in the repository
func (r *Repository) GetBranches() ([]string, error) {
	if !IsGitInstalled() {
		return nil, fmt.Errorf("git is not installed or not available in PATH")
	}

	if !r.IsGitRepository() {
		return nil, fmt.Errorf("not a git repository: %s", r.Path)
	}

	cmd := exec.Command("git", "branch", "-a", "--format=%(refname:short)")
	cmd.Dir = r.Path
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get branches: %v", err)
	}

	branches := make([]string, 0)
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	
	for _, line := range lines {
		branch := strings.TrimSpace(line)
		if branch != "" && !strings.HasPrefix(branch, "origin/HEAD") {
			// Remove "origin/" prefix for remote branches if needed
			if strings.HasPrefix(branch, "origin/") {
				branch = strings.TrimPrefix(branch, "origin/")
			}
			branches = append(branches, branch)
		}
	}

	// Remove duplicates
	uniqueBranches := make([]string, 0)
	seen := make(map[string]bool)
	for _, branch := range branches {
		if !seen[branch] {
			uniqueBranches = append(uniqueBranches, branch)
			seen[branch] = true
		}
	}

	return uniqueBranches, nil
}

// GetBranchCommits retrieves commits for a specific branch
func (r *Repository) GetBranchCommits(branch string) ([]GitCommit, error) {
	if !IsGitInstalled() {
		return nil, fmt.Errorf("git is not installed or not available in PATH")
	}

	if !r.IsGitRepository() {
		return nil, fmt.Errorf("not a git repository: %s", r.Path)
	}

	// Git log format: hash|author|email|date|subject|body|parents
	format := "--pretty=format:%H|%an|%ae|%ai|%s|%b|%P"
	args := []string{"log", format, branch}

	cmd := exec.Command("git", args...)
	cmd.Dir = r.Path
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get commits for branch %s: %v", branch, err)
	}

	return parseCommits(string(output))
}

// GetCommitBranch determines which branch a commit belongs to
func (r *Repository) GetCommitBranch(commitHash string) (string, error) {
	if !IsGitInstalled() {
		return "", fmt.Errorf("git is not installed or not available in PATH")
	}

	if !r.IsGitRepository() {
		return "", fmt.Errorf("not a git repository: %s", r.Path)
	}

	// Try to find which branch contains this commit
	cmd := exec.Command("git", "branch", "--contains", commitHash)
	cmd.Dir = r.Path
	output, err := cmd.Output()
	if err != nil {
		return "unknown", nil // Don't fail, just return unknown
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		branch := strings.TrimSpace(line)
		if branch != "" {
			// Remove leading * if present
			if strings.HasPrefix(branch, "* ") {
				branch = strings.TrimPrefix(branch, "* ")
			}
			return branch, nil
		}
	}

	return "main", nil // Default to main if not found
}
