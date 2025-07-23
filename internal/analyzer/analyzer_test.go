package analyzer

import (
	"testing"
	"time"

	"git-log-analyzer/internal/git"
)

func TestNewAnalyzer(t *testing.T) {
	analyzer := NewAnalyzer("/tmp")
	if analyzer == nil {
		t.Error("NewAnalyzer should not return nil")
	}
	if analyzer.repo.Path != "/tmp" {
		t.Errorf("Expected repo path '/tmp', got '%s'", analyzer.repo.Path)
	}
}

func TestProcessCommit(t *testing.T) {
	stats := &Statistics{
		AuthorStats:     make(map[string]*AuthorStat),
		FileStats:       make(map[string]int),
		CommitFrequency: make(map[string]int),
		TimeStats: &TimeStat{
			HourlyPattern: make(map[int]int),
			DailyPattern:  make(map[time.Weekday]int),
		},
	}

	commit := &git.GitCommit{
		Hash:    "abc123",
		Author:  "John Doe",
		Email:   "john@example.com",
		Date:    time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC),
		Subject: "Test commit",
		Files:   []string{"file1.go", "file2.go"},
	}

	analyzer := NewAnalyzer(".")
	analyzer.processCommit(commit, stats)

	// Check author stats
	authorKey := "John Doe <john@example.com>"
	if _, exists := stats.AuthorStats[authorKey]; !exists {
		t.Error("Author should be added to stats")
	}

	authorStat := stats.AuthorStats[authorKey]
	if authorStat.CommitCount != 1 {
		t.Errorf("Expected commit count 1, got %d", authorStat.CommitCount)
	}
	if authorStat.Name != "John Doe" {
		t.Errorf("Expected author name 'John Doe', got '%s'", authorStat.Name)
	}

	// Check time stats
	if stats.TimeStats.HourlyPattern[10] != 1 {
		t.Errorf("Expected 1 commit at hour 10, got %d", stats.TimeStats.HourlyPattern[10])
	}
	if stats.TimeStats.DailyPattern[time.Sunday] != 1 {
		t.Errorf("Expected 1 commit on Sunday, got %d", stats.TimeStats.DailyPattern[time.Sunday])
	}

	// Check commit frequency
	dateKey := "2023-01-01"
	if stats.CommitFrequency[dateKey] != 1 {
		t.Errorf("Expected 1 commit on %s, got %d", dateKey, stats.CommitFrequency[dateKey])
	}
}

func TestGetWeekNumber(t *testing.T) {
	// Test week number calculation
	testDate := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	week := getWeekNumber(testDate)
	if week < 1 || week > 53 {
		t.Errorf("Week number should be between 1 and 53, got %d", week)
	}
}

func TestGenerateReport(t *testing.T) {
	stats := &Statistics{
		TotalCommits:    10,
		AuthorStats:     make(map[string]*AuthorStat),
		FileStats:       make(map[string]int),
		CommitFrequency: make(map[string]int),
		TimeStats: &TimeStat{
			FirstCommit:   time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			LastCommit:    time.Date(2023, 1, 10, 0, 0, 0, 0, time.UTC),
			ActiveDays:    10,
			ActiveWeeks:   2,
			ActiveMonths:  1,
			HourlyPattern: make(map[int]int),
			DailyPattern:  make(map[time.Weekday]int),
		},
	}

	// Add some test data
	stats.AuthorStats["John Doe <john@example.com>"] = &AuthorStat{
		Name:        "John Doe",
		CommitCount: 5,
		Additions:   100,
		Deletions:   20,
	}

	stats.FileStats["main.go"] = 3
	stats.TimeStats.HourlyPattern[9] = 5

	report := stats.GenerateReport()
	if report == "" {
		t.Error("Report should not be empty")
	}

	// Check that report contains expected sections
	expectedSections := []string{
		"=== Git Repository Analysis Report ===",
		"Total Commits: 10",
		"=== Top Contributors ===",
		"=== Most Active Hours ===",
		"=== Most Modified Files ===",
	}

	for _, section := range expectedSections {
		if !contains(report, section) {
			t.Errorf("Report should contain section: %s", section)
		}
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		(len(s) > len(substr) && (s[:len(substr)] == substr || 
		s[len(s)-len(substr):] == substr || 
		containsMiddle(s, substr))))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
