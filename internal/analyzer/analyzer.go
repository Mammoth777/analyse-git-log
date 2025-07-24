package analyzer

import (
	"fmt"
	"sort"
	"time"

	"git-log-analyzer/internal/git"
	"git-log-analyzer/internal/i18n"
)

// Statistics contains analysis results
type Statistics struct {
	TotalCommits     int
	AuthorStats      map[string]*AuthorStat
	TimeStats        *TimeStat
	FileStats        map[string]int
	CommitFrequency  map[string]int // date -> count
}

// AuthorStat contains statistics for a single author
type AuthorStat struct {
	Name         string
	Email        string
	CommitCount  int
	Additions    int
	Deletions    int
	FirstCommit  time.Time
	LastCommit   time.Time
	Files        map[string]int
}

// TimeStat contains time-based statistics
type TimeStat struct {
	FirstCommit   time.Time
	LastCommit    time.Time
	ActiveDays    int
	ActiveWeeks   int
	ActiveMonths  int
	HourlyPattern map[int]int // hour -> count
	DailyPattern  map[time.Weekday]int
}

// Analyzer analyzes git commits
type Analyzer struct {
	repo *git.Repository
}

// NewAnalyzer creates a new analyzer instance
func NewAnalyzer(repoPath string) *Analyzer {
	return &Analyzer{
		repo: git.NewRepository(repoPath),
	}
}

// Analyze performs comprehensive analysis of the git repository
func (a *Analyzer) Analyze() (*Statistics, error) {
	commits, err := a.repo.GetCommits(0) // Get all commits
	if err != nil {
		return nil, err
	}

	if len(commits) == 0 {
		return nil, fmt.Errorf("no commits found in repository")
	}

	stats := &Statistics{
		AuthorStats:     make(map[string]*AuthorStat),
		FileStats:       make(map[string]int),
		CommitFrequency: make(map[string]int),
		TimeStats: &TimeStat{
			HourlyPattern: make(map[int]int),
			DailyPattern:  make(map[time.Weekday]int),
		},
	}

	stats.TotalCommits = len(commits)

	// Process each commit
	for _, commit := range commits {
		a.processCommit(&commit, stats)
	}

	// Calculate time statistics
	a.calculateTimeStats(commits, stats.TimeStats)

	return stats, nil
}

// processCommit processes a single commit and updates statistics
func (a *Analyzer) processCommit(commit *git.GitCommit, stats *Statistics) {
	authorKey := fmt.Sprintf("%s <%s>", commit.Author, commit.Email)
	
	// Update author statistics
	if _, exists := stats.AuthorStats[authorKey]; !exists {
		stats.AuthorStats[authorKey] = &AuthorStat{
			Name:        commit.Author,
			Email:       commit.Email,
			FirstCommit: commit.Date,
			LastCommit:  commit.Date,
			Files:       make(map[string]int),
		}
	}

	authorStat := stats.AuthorStats[authorKey]
	authorStat.CommitCount++

	if commit.Date.Before(authorStat.FirstCommit) {
		authorStat.FirstCommit = commit.Date
	}
	if commit.Date.After(authorStat.LastCommit) {
		authorStat.LastCommit = commit.Date
	}

	// Get detailed commit statistics
	additions, deletions, files, err := a.repo.GetCommitStats(commit.Hash)
	if err == nil {
		authorStat.Additions += additions
		authorStat.Deletions += deletions
		commit.Additions = additions
		commit.Deletions = deletions
		commit.Files = files

		// Update file statistics
		for _, file := range files {
			stats.FileStats[file]++
			authorStat.Files[file]++
		}
	}

	// Update time-based statistics
	dateKey := commit.Date.Format("2006-01-02")
	stats.CommitFrequency[dateKey]++
	
	stats.TimeStats.HourlyPattern[commit.Date.Hour()]++
	stats.TimeStats.DailyPattern[commit.Date.Weekday()]++
}

// calculateTimeStats calculates time-related statistics
func (a *Analyzer) calculateTimeStats(commits []git.GitCommit, timeStats *TimeStat) {
	if len(commits) == 0 {
		return
	}

	// Sort commits by date
	sort.Slice(commits, func(i, j int) bool {
		return commits[i].Date.Before(commits[j].Date)
	})

	timeStats.FirstCommit = commits[0].Date
	timeStats.LastCommit = commits[len(commits)-1].Date

	// Calculate active periods
	uniqueDays := make(map[string]bool)
	uniqueWeeks := make(map[string]bool)
	uniqueMonths := make(map[string]bool)

	for _, commit := range commits {
		day := commit.Date.Format("2006-01-02")
		week := fmt.Sprintf("%d-W%02d", commit.Date.Year(), getWeekNumber(commit.Date))
		month := commit.Date.Format("2006-01")

		uniqueDays[day] = true
		uniqueWeeks[week] = true
		uniqueMonths[month] = true
	}

	timeStats.ActiveDays = len(uniqueDays)
	timeStats.ActiveWeeks = len(uniqueWeeks)
	timeStats.ActiveMonths = len(uniqueMonths)
}

// getWeekNumber returns the week number of the year
func getWeekNumber(t time.Time) int {
	_, week := t.ISOWeek()
	return week
}

// GenerateReport generates a text report from statistics
func (stats *Statistics) GenerateReport() string {
	msg := i18n.T()
	report := fmt.Sprintf("%s\n\n", msg.ReportTitle)
	
	report += fmt.Sprintf("%s: %d\n", msg.TotalCommits, stats.TotalCommits)
	report += fmt.Sprintf("%s: %s to %s\n", msg.ActivePeriod,
		stats.TimeStats.FirstCommit.Format("2006-01-02"),
		stats.TimeStats.LastCommit.Format("2006-01-02"))
	report += fmt.Sprintf("%s: %d\n", msg.ActiveDays, stats.TimeStats.ActiveDays)
	report += fmt.Sprintf("%s: %d\n", msg.ActiveWeeks, stats.TimeStats.ActiveWeeks)
	report += fmt.Sprintf("%s: %d\n\n", msg.ActiveMonths, stats.TimeStats.ActiveMonths)

	// Top authors by commit count
	report += msg.TopContributors + "\n"
	type authorPair struct {
		key   string
		stats *AuthorStat
	}
	
	var authors []authorPair
	for key, stat := range stats.AuthorStats {
		authors = append(authors, authorPair{key, stat})
	}
	
	sort.Slice(authors, func(i, j int) bool {
		return authors[i].stats.CommitCount > authors[j].stats.CommitCount
	})

	for i, author := range authors {
		if i >= 10 { // Top 10 authors
			break
		}
		report += fmt.Sprintf("%d. %s: %d %s (+%d/-%d %s)\n",
			i+1, author.stats.Name, author.stats.CommitCount, msg.Commits,
			author.stats.Additions, author.stats.Deletions, msg.Lines)
	}

	// Most active hours
	report += "\n" + msg.MostActiveHours + "\n"
	type hourPair struct {
		hour  int
		count int
	}
	
	var hours []hourPair
	for hour, count := range stats.TimeStats.HourlyPattern {
		hours = append(hours, hourPair{hour, count})
	}
	
	sort.Slice(hours, func(i, j int) bool {
		return hours[i].count > hours[j].count
	})

	for i, h := range hours {
		if i >= 5 { // Top 5 hours
			break
		}
		report += fmt.Sprintf("%02d:00 - %d %s\n", h.hour, h.count, msg.Commits)
	}

	// Most modified files
	report += "\n" + msg.MostModifiedFiles + "\n"
	type filePair struct {
		file  string
		count int
	}
	
	var files []filePair
	for file, count := range stats.FileStats {
		files = append(files, filePair{file, count})
	}
	
	sort.Slice(files, func(i, j int) bool {
		return files[i].count > files[j].count
	})

	for i, f := range files {
		if i >= 10 { // Top 10 files
			break
		}
		report += fmt.Sprintf("%s: %d %s\n", f.file, f.count, msg.Modifications)
	}

	return report
}
