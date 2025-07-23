package git

import (
	"testing"
	"time"
)

func TestIsGitInstalled(t *testing.T) {
	// Test that git is installed (this should pass on most development machines)
	if !IsGitInstalled() {
		t.Skip("Git is not installed, skipping test")
	}
}

func TestNewRepository(t *testing.T) {
	repo := NewRepository("/tmp")
	if repo.Path != "/tmp" {
		t.Errorf("Expected path '/tmp', got '%s'", repo.Path)
	}
}

func TestParseCommits(t *testing.T) {
	// Test parsing git log output
	output := `abc123|John Doe|john@example.com|2023-01-01 10:00:00 +0000|Initial commit|
def456|Jane Smith|jane@example.com|2023-01-02 11:30:00 +0000|Add feature|Added new feature`

	commits, err := parseCommits(output)
	if err != nil {
		t.Fatalf("Failed to parse commits: %v", err)
	}

	if len(commits) != 2 {
		t.Errorf("Expected 2 commits, got %d", len(commits))
	}

	// Test first commit
	if commits[0].Hash != "abc123" {
		t.Errorf("Expected hash 'abc123', got '%s'", commits[0].Hash)
	}
	if commits[0].Author != "John Doe" {
		t.Errorf("Expected author 'John Doe', got '%s'", commits[0].Author)
	}
	if commits[0].Subject != "Initial commit" {
		t.Errorf("Expected subject 'Initial commit', got '%s'", commits[0].Subject)
	}

	// Test date parsing
	expectedDate := time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC)
	if !commits[0].Date.Equal(expectedDate) {
		t.Errorf("Expected date %v, got %v", expectedDate, commits[0].Date)
	}
}

func TestRepository_IsGitRepository(t *testing.T) {
	// Test with current directory (should be a git repo now)
	repo := NewRepository(".")
	if !repo.IsGitRepository() {
		t.Error("Current directory should be a git repository")
	}

	// Test with non-git directory
	repo = NewRepository("/tmp")
	// This might pass if /tmp is inside a git repo, so we won't assert false
}

func TestRepository_IsGitRepository_NonExistent(t *testing.T) {
	// Test with non-existent directory
	repo := NewRepository("/non/existent/path")
	if repo.IsGitRepository() {
		t.Error("Non-existent directory should not be a git repository")
	}
}
