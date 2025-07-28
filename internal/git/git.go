package git

import (
	"bufio"
	"context"
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

// ProgressCallback is a function type for progress reporting
type ProgressCallback func(current, total int, message string)

// GetCommits retrieves git commits from the repository
func (r *Repository) GetCommits(limit int) ([]GitCommit, error) {
	return r.GetCommitsWithProgress(limit, nil)
}

// GetCommitsWithProgress retrieves git commits from the repository with progress callback
func (r *Repository) GetCommitsWithProgress(limit int, progressCallback ProgressCallback) ([]GitCommit, error) {
	return r.GetCommitsWithTimeRangeAndProgress(limit, 0, progressCallback)
}

// GetCommitsWithTimeRangeAndProgress retrieves git commits from the repository with time range and progress callback
func (r *Repository) GetCommitsWithTimeRangeAndProgress(limit int, months int, progressCallback ProgressCallback) ([]GitCommit, error) {
	if !IsGitInstalled() {
		return nil, fmt.Errorf("git is not installed or not available in PATH")
	}

	if !r.IsGitRepository() {
		return nil, fmt.Errorf("not a git repository: %s", r.Path)
	}

	// Git log format: hash|author|email|date|subject|body|parents
	format := "--pretty=format:%H|%an|%ae|%ai|%s|%b|%P"
	args := []string{"log", format}
	
	// Add time range if specified
	if months > 0 {
		since := time.Now().AddDate(0, -months, 0).Format("2006-01-02")
		args = append(args, fmt.Sprintf("--since=%s", since))
	}
	
	if limit > 0 {
		args = append(args, fmt.Sprintf("-%d", limit))
	}

	cmd := exec.Command("git", args...)
	cmd.Dir = r.Path
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get git log: %v", err)
	}

	return parseCommitsWithProgress(string(output), progressCallback)
}

// parseCommits parses git log output into GitCommit structs
func parseCommits(output string) ([]GitCommit, error) {
	return parseCommitsWithProgress(output, nil)
}

// parseCommitsWithProgress parses git log output into GitCommit structs with progress reporting
func parseCommitsWithProgress(output string, progressCallback ProgressCallback) ([]GitCommit, error) {
	var commits []GitCommit
	lines := strings.Split(output, "\n")
	
	// Filter out empty lines first to get accurate count
	var validLines []string
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			validLines = append(validLines, line)
		}
	}
	
	totalLines := len(validLines)
	processed := 0
	
	for i, line := range validLines {
		line = strings.TrimSpace(line)
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
		processed++
		
		// Only report progress at the very end
		if progressCallback != nil && i == totalLines-1 {
			progressCallback(processed, totalLines, fmt.Sprintf("已解析 %d 个提交", processed))
		}
	}

	return commits, nil
}

// GetCommitStats gets detailed statistics for a commit
func (r *Repository) GetCommitStats(hash string) (int, int, []string, error) {
	return r.GetCommitStatsWithProgress(hash, nil, 0, 0)
}

// GetCommitStatsBatch gets detailed statistics for multiple commits in batches
func (r *Repository) GetCommitStatsBatch(commits []GitCommit, progressCallback ProgressCallback) error {
	total := len(commits)
	
	// For large repositories (>1000 commits), use optimized batch method
	if total > 1000 {
		return r.getCommitStatsBatchOptimized(commits, progressCallback)
	}
	
	// For smaller repositories, use the original method
	const batchSize = 100
	
	for i := 0; i < total; i += batchSize {
		end := i + batchSize
		if end > total {
			end = total
		}
		
		// Process batch
		batch := commits[i:end]
		
		if progressCallback != nil {
			progressCallback(i, total, fmt.Sprintf("开始处理批次 %d-%d/%d", i+1, end, total))
		}
		
		// Process each commit in the batch
		for j, commit := range batch {
			additions, deletions, files, err := r.GetCommitStatsWithProgress(
				commit.Hash, 
				progressCallback, 
				i+j+1, 
				total,
			)
			if err == nil {
				// Update the commit with statistics
				commits[i+j].Additions = additions
				commits[i+j].Deletions = deletions
				commits[i+j].Files = files
			}
		}
		
		if progressCallback != nil {
			progressCallback(end, total, fmt.Sprintf("完成批次 %d-%d/%d", i+1, end, total))
		}
	}
	
	return nil
}

// getCommitStatsBatchOptimized uses single git log --stat command for better performance
func (r *Repository) getCommitStatsBatchOptimized(commits []GitCommit, progressCallback ProgressCallback) error {
	total := len(commits)
	const maxBatchSize = 5000 // Process max 5000 commits at once to avoid memory issues
	
	if progressCallback != nil {
		progressCallback(0, total, "开始优化批量处理...")
	}
	
	// Create a map for quick lookup of commits by hash
	commitMap := make(map[string]*GitCommit)
	for i := range commits {
		commitMap[commits[i].Hash] = &commits[i]
	}
	
	// Process commits in chunks to avoid command line length limits and memory issues
	for start := 0; start < total; start += maxBatchSize {
		end := start + maxBatchSize
		if end > total {
			end = total
		}
		
		// Create hash list for this chunk
		var hashes []string
		for i := start; i < end; i++ {
			hashes = append(hashes, commits[i].Hash)
		}
		
		if progressCallback != nil {
			progressCallback(start, total, fmt.Sprintf("处理提交块 %d-%d/%d...", start+1, end, total))
		}
		
		err := r.processCommitStatsChunk(hashes, commitMap, progressCallback, start, total)
		if err != nil {
			// Fallback to individual processing if batch fails
			if progressCallback != nil {
				progressCallback(start, total, fmt.Sprintf("批量处理失败，回退到逐个处理 %d-%d/%d", start+1, end, total))
			}
			
			for i := start; i < end; i++ {
				additions, deletions, files, err := r.GetCommitStatsWithProgress(
					commits[i].Hash,
					progressCallback,
					i+1,
					total,
				)
				if err == nil {
					commits[i].Additions = additions
					commits[i].Deletions = deletions
					commits[i].Files = files
				}
			}
		}
		
		if progressCallback != nil {
			progressCallback(end, total, fmt.Sprintf("完成提交块 %d-%d/%d", start+1, end, total))
		}
	}
	
	return nil
}

// processCommitStatsChunk processes a chunk of commits using single git log --stat command
func (r *Repository) processCommitStatsChunk(hashes []string, commitMap map[string]*GitCommit, progressCallback ProgressCallback, offset, total int) error {
	if len(hashes) == 0 {
		return nil
	}
	
	// Build git log command with --stat for all hashes
	args := []string{"log", "--stat", "--format=%H", "--no-merges"}
	
	// Add individual commit hashes
	for _, hash := range hashes {
		args = append(args, hash, "-1") // -1 limits to single commit
	}
	
	// Set timeout for large operations to prevent hanging
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = r.Path
	
	if progressCallback != nil {
		progressCallback(offset, total, fmt.Sprintf("执行Git命令获取 %d 个提交的统计信息...", len(hashes)))
	}
	
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get commit stats batch: %v", err)
	}
	
	if progressCallback != nil {
		progressCallback(offset, total, fmt.Sprintf("解析 %d 个提交的统计信息...", len(hashes)))
	}
	
	return r.parseStatsBatchOutput(string(output), commitMap, progressCallback, offset, total)
}

// parseStatsBatchOutput parses the output from git log --stat command
func (r *Repository) parseStatsBatchOutput(output string, commitMap map[string]*GitCommit, progressCallback ProgressCallback, offset, total int) error {
	lines := strings.Split(output, "\n")
	var currentHash string
	var currentFiles []string
	var currentAdditions, currentDeletions int
	processed := 0
	
	for i, line := range lines {
		line = strings.TrimSpace(line)
		
		// Check if this is a commit hash line
		if len(line) == 40 && isValidHash(line) {
			// Save previous commit stats if we have one
			if currentHash != "" {
				if commit, exists := commitMap[currentHash]; exists {
					commit.Files = currentFiles
					commit.Additions = currentAdditions
					commit.Deletions = currentDeletions
					processed++
				}
			}
			
			// Start new commit
			currentHash = line
			currentFiles = []string{}
			currentAdditions = 0
			currentDeletions = 0
			continue
		}
		
		// Skip empty lines
		if line == "" {
			continue
		}
		
		// Parse file statistics
		if strings.Contains(line, "|") && currentHash != "" {
			parts := strings.Split(line, "|")
			if len(parts) >= 2 {
				filename := strings.TrimSpace(parts[0])
				currentFiles = append(currentFiles, filename)
			}
		}
		
		// Parse summary line (e.g., "2 files changed, 15 insertions(+), 3 deletions(-)")
		if (strings.Contains(line, "insertion") || strings.Contains(line, "deletion")) && currentHash != "" {
			words := strings.Fields(line)
			for j, word := range words {
				if strings.Contains(word, "insertion") && j > 0 {
					fmt.Sscanf(words[j-1], "%d", &currentAdditions)
				}
				if strings.Contains(word, "deletion") && j > 0 {
					fmt.Sscanf(words[j-1], "%d", &currentDeletions)
				}
			}
		}
		
		// Report progress every 100 lines for better performance
		if progressCallback != nil && i%100 == 0 {
			progressCallback(offset+processed, total, fmt.Sprintf("解析统计信息 %d/%d 行...", i+1, len(lines)))
		}
	}
	
	// Handle the last commit
	if currentHash != "" {
		if commit, exists := commitMap[currentHash]; exists {
			commit.Files = currentFiles
			commit.Additions = currentAdditions
			commit.Deletions = currentDeletions
			processed++
		}
	}
	
	if progressCallback != nil {
		progressCallback(offset+processed, total, fmt.Sprintf("完成解析 %d 个提交的统计信息", processed))
	}
	
	return nil
}

// isValidHash checks if a string is a valid git hash
func isValidHash(s string) bool {
	if len(s) != 40 {
		return false
	}
	for _, r := range s {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')) {
			return false
		}
	}
	return true
}

// GetCommitStatsWithProgress gets detailed statistics for a commit with progress reporting
func (r *Repository) GetCommitStatsWithProgress(hash string, progressCallback ProgressCallback, current, total int) (int, int, []string, error) {
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

	// Report progress if callback is provided
	if progressCallback != nil && total > 0 {
		progressCallback(current, total, fmt.Sprintf("获取提交 %s 的统计信息 (%d/%d)", hash[:8], current, total))
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
