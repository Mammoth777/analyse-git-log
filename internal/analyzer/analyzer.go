package analyzer

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"git-log-analyzer/internal/git"
	"git-log-analyzer/internal/health"
	"git-log-analyzer/internal/i18n"
)

// Statistics contains analysis results
type Statistics struct {
	TotalCommits     int
	AuthorStats      map[string]*AuthorStat
	TimeStats        *TimeStat
	FileStats        map[string]int
	CommitFrequency  map[string]int // date -> count
	CodeHealthMetrics *health.CodeHealthMetrics // 代码健康分析
	BranchData       *BranchData // 分支数据
}

// BranchData contains simplified branch overview and lifecycle analysis
type BranchData struct {
	Branches        []BranchInfo     `json:"branches"`
	Summary         BranchSummary    `json:"summary"`
	LifecycleStats  LifecycleStats   `json:"lifecycle_stats"`
}

// BranchInfo contains essential information about a single branch
type BranchInfo struct {
	Name           string          `json:"name"`
	Type           string          `json:"type"`           // main, feature, release, hotfix, etc.
	Status         string          `json:"status"`         // active, merged, stale, abandoned
	CommitCount    int             `json:"commit_count"`
	FirstCommit    time.Time       `json:"first_commit"`
	LastCommit     time.Time       `json:"last_commit"`
	LastCommitHash string          `json:"last_commit_hash"` // 最后提交的hash
	LifespanDays   int             `json:"lifespan_days"`
	IsActive       bool            `json:"is_active"`
	MainAuthors    []BranchAuthor  `json:"main_authors"`
	AdditionsTotal int             `json:"additions_total"`
	DeletionsTotal int             `json:"deletions_total"`
	ActivityLevel  string          `json:"activity_level"` // high, medium, low
}

// BranchAuthor represents an author's contribution to a branch
type BranchAuthor struct {
	Name        string `json:"name"`
	Email       string `json:"email"`
	CommitCount int    `json:"commit_count"`
	Percentage  float64 `json:"percentage"`
}

// BranchSummary provides overall branch statistics
type BranchSummary struct {
	TotalBranches         int            `json:"total_branches"`          // 分支总数
	ActiveBranchesInRange int            `json:"active_branches_in_range"` // 指定时间内的活跃分支数量
	RecentActiveBranches  []BranchInfo   `json:"recent_active_branches"`   // 最近半个月的活跃分支列表
}

// LifecycleStats provides branch lifecycle analysis
type LifecycleStats struct {
	ShortLivedBranches  []BranchInfo `json:"short_lived_branches"`  // < 7 days
	LongLivedBranches   []BranchInfo `json:"long_lived_branches"`   // > 90 days
	MergePatterns       []MergePattern `json:"merge_patterns"`
	BranchingFrequency  map[string]int `json:"branching_frequency"`  // commits per time period
}

// MergePattern contains information about merge operations and patterns
type MergePattern struct {
	FromBranch    string    `json:"from_branch"`
	ToBranch      string    `json:"to_branch"`
	MergeCommit   string    `json:"merge_commit"`
	MergeDate     time.Time `json:"merge_date"`
	ConflictRisk  string    `json:"conflict_risk"` // low, medium, high
	Date          time.Time `json:"date"`
	Author        string    `json:"author"`
	CommitCount   int       `json:"commit_count"` // 合并的提交数量
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

// ProgressCallback defines the progress callback function type
type ProgressCallback func(current, total int, message string)

// Analyzer analyzes git commits
type Analyzer struct {
	repo         *git.Repository
	analysisMonths int
}

// NewAnalyzer creates a new analyzer instance
func NewAnalyzer(repoPath string, analysisMonths int) *Analyzer {
	return &Analyzer{
		repo:           git.NewRepository(repoPath),
		analysisMonths: analysisMonths,
	}
}

// Analyze performs comprehensive analysis of the git repository
func (a *Analyzer) Analyze() (*Statistics, error) {
	return a.AnalyzeWithProgress(nil)
}

// AnalyzeWithProgress performs comprehensive analysis with progress callback
func (a *Analyzer) AnalyzeWithProgress(progressCallback ProgressCallback) (*Statistics, error) {
	if progressCallback != nil {
		progressCallback(0, 100, "开始获取提交历史...")
	}
	
	commits, err := a.repo.GetCommitsWithTimeRangeAndProgress(0, a.analysisMonths, func(current, total int, message string) {
		if progressCallback != nil {
			// Map commit parsing to 0-20% of overall progress
			progress := int(float64(current) / float64(total) * 20)
			progressCallback(progress, 100, message)
		}
	})
	if err != nil {
		return nil, err
	}

	if len(commits) == 0 {
		return nil, fmt.Errorf("no commits found in repository")
	}

	if progressCallback != nil {
		progressCallback(20, 100, fmt.Sprintf("获取到 %d 个提交，开始详细分析...", len(commits)))
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

	// Use batch processing for better performance with large repositories
	if progressCallback != nil {
		progressCallback(20, 100, "开始批量处理提交统计...")
	}
	
	// Get commit statistics in batches using the optimized batch method
	batchErr := a.repo.GetCommitStatsBatch(commits, func(current, total int, message string) {
		if progressCallback != nil {
			// Map batch progress to 20-50% of overall progress
			progress := 20 + int(float64(current)/float64(total)*30)
			progressCallback(progress, 100, message)
		}
	})
	if batchErr != nil {
		fmt.Printf("Warning: Failed to get commit statistics in batch: %v\n", batchErr)
	}

	// Process each commit for author statistics and other metadata
	for i, commit := range commits {
		a.processCommitMetadata(&commit, stats)
		
		// Only report progress at the very end
		if progressCallback != nil && i == len(commits)-1 {
			currentProgress := 50 + int(float64(i+1)/float64(len(commits))*10)
			progressCallback(currentProgress, 100, fmt.Sprintf("处理提交元数据 %d 个", len(commits)))
		}
	}

	if progressCallback != nil {
		progressCallback(60, 100, "计算时间统计...")
	}

	// Calculate time statistics
	a.calculateTimeStats(commits, stats.TimeStats)

	if progressCallback != nil {
		progressCallback(80, 100, "分析分支结构...")
	}

	// Analyze branch structure
	branchData, err := a.analyzeBranchStructure(commits)
	if err != nil {
		// Branch analysis is optional, continue without it
		fmt.Printf("Warning: Failed to analyze branch structure: %v\n", err)
	} else {
		stats.BranchData = branchData
	}

	if progressCallback != nil {
		progressCallback(90, 100, "分析代码健康度...")
	}

	// Perform code health analysis
	healthAnalyzer := health.NewCodeHealthAnalyzer(commits)
	stats.CodeHealthMetrics = healthAnalyzer.AnalyzeCodeHealth()

	if progressCallback != nil {
		progressCallback(100, 100, "分析完成!")
	}

	return stats, nil
}

// processCommit processes a single commit and updates statistics
func (a *Analyzer) processCommit(commit *git.GitCommit, stats *Statistics) {
	a.processCommitWithProgress(commit, stats, nil, 0, 0)
}

// processCommitMetadata processes commit metadata (author stats, time stats) excluding file statistics
func (a *Analyzer) processCommitMetadata(commit *git.GitCommit, stats *Statistics) {
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

	// Use the file statistics that were already populated by the batch processing
	if len(commit.Files) > 0 {
		authorStat.Additions += commit.Additions
		authorStat.Deletions += commit.Deletions

		// Update file statistics
		for _, file := range commit.Files {
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

// processCommitWithProgress processes a single commit with progress reporting
func (a *Analyzer) processCommitWithProgress(commit *git.GitCommit, stats *Statistics, progressCallback ProgressCallback, current, total int) {
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

	// Get detailed commit statistics with progress reporting
	var gitProgressCallback git.ProgressCallback
	if progressCallback != nil {
		gitProgressCallback = func(current, total int, message string) {
			progressCallback(current, total, message)
		}
	}
	
	additions, deletions, files, err := a.repo.GetCommitStatsWithProgress(commit.Hash, gitProgressCallback, current, total)
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

	// Add code health analysis
	if stats.CodeHealthMetrics != nil {
		report += "\n\n=== 代码健康分析 ===\n"
		report += stats.CodeHealthMetrics.HealthSummary + "\n\n"
		
		// Technical debt hotspots
		if len(stats.CodeHealthMetrics.TechnicalDebtHotspots) > 0 {
			report += "技术债务热点:\n"
			for i, hotspot := range stats.CodeHealthMetrics.TechnicalDebtHotspots {
				if i >= 5 { // Top 5
					break
				}
				report += fmt.Sprintf("%d. %s (风险分数: %.2f, 修改次数: %d, 原因: %s)\n",
					i+1, hotspot.FilePath, hotspot.RiskScore, hotspot.ModificationFreq, hotspot.Reason)
			}
			report += "\n"
		}
		
		// Refactoring signals
		if len(stats.CodeHealthMetrics.RefactoringSignals) > 0 {
			report += "重构信号:\n"
			for i, signal := range stats.CodeHealthMetrics.RefactoringSignals {
				if i >= 5 { // Top 5
					break
				}
				report += fmt.Sprintf("%d. %s (%s信号, %d次修改在%d天内)\n",
					i+1, signal.FilePath, signal.RefactoringSignal, signal.ShortTermChanges, signal.IntensiveModDays)
			}
			report += "\n"
		}
		
		// Code concentration issues
		if len(stats.CodeHealthMetrics.CodeConcentrationIssues) > 0 {
			report += "代码集中度问题:\n"
			for i, issue := range stats.CodeHealthMetrics.CodeConcentrationIssues {
				if i >= 3 { // Top 3
					break
				}
				report += fmt.Sprintf("%d. %s (%s, 占总变更%.1f%%, %d次修改)\n",
					i+1, issue.FilePath, issue.ConcentrationLevel, issue.ChangeRatio*100, issue.TotalChanges)
			}
		}
	}

	return report
}

// analyzeBranchStructure analyzes git branch structure and commit relationships
func (a *Analyzer) analyzeBranchStructure(commits []git.GitCommit) (*BranchData, error) {
	branchData := &BranchData{
		Branches:       make([]BranchInfo, 0),
		Summary:        BranchSummary{},
		LifecycleStats: LifecycleStats{}, // 暂时清空，功能开发中
	}

	// Get branch information from git
	branches, err := a.repo.GetBranches()
	if err != nil {
		return nil, fmt.Errorf("failed to get branches: %v", err)
	}

	// Analyze each branch
	branchStats := make(map[string]*BranchInfo)
	
	for _, branch := range branches {
		branchInfo := a.analyzeSingleBranch(branch, commits)
		if branchInfo != nil {
			branchStats[branch] = branchInfo
			branchData.Branches = append(branchData.Branches, *branchInfo)
		}
	}

	// Calculate summary statistics
	branchData.Summary = a.calculateBranchSummary(branchStats)

	// 生命周期分析暂时清空（功能开发中）
	branchData.LifecycleStats = LifecycleStats{}

	return branchData, nil
}

// analyzeSingleBranch analyzes a single branch with simplified approach
func (a *Analyzer) analyzeSingleBranch(branchName string, allCommits []git.GitCommit) *BranchInfo {
	// Get actual last commit info for this branch
	lastCommitHash, lastCommitDate, err := a.repo.GetBranchLastCommit(branchName)
	if err != nil || lastCommitHash == "" {
		// Fallback to simplified approach if we can't get branch info
		return a.analyzeSingleBranchFallback(branchName, allCommits)
	}

	// For performance, we'll use a simplified approach for commit assignment:
	// - Main/master branch gets most commits
	// - Other branches get estimated based on naming patterns
	
	var branchCommits []git.GitCommit
	
	// Simple heuristic: if it's main/master, assign most commits
	if branchName == "main" || branchName == "master" {
		branchCommits = allCommits
	} else {
		// For other branches, we'll estimate based on commit messages and patterns
		branchCommits = a.estimateBranchCommits(branchName, allCommits)
	}

	if len(branchCommits) == 0 {
		return &BranchInfo{
			Name:           branchName,
			Type:           a.determineBranchType(branchName),
			Status:         "empty",
			CommitCount:    0,
			IsActive:       false,
			MainAuthors:    make([]BranchAuthor, 0),
			LastCommit:     lastCommitDate,
			LastCommitHash: lastCommitHash,
		}
	}

	// Sort commits by date
	sortedCommits := make([]git.GitCommit, len(branchCommits))
	copy(sortedCommits, branchCommits)
	sort.Slice(sortedCommits, func(i, j int) bool {
		return sortedCommits[i].Date.Before(sortedCommits[j].Date)
	})

	firstCommit := sortedCommits[0].Date
	// Use the actual last commit date from the branch
	lastCommit := lastCommitDate
	if lastCommit.IsZero() {
		// Fallback to estimated last commit
		lastCommit = sortedCommits[len(sortedCommits)-1].Date
	}
	lifespanDays := int(lastCommit.Sub(firstCommit).Hours() / 24)

	// Calculate totals
	totalAdditions := 0
	totalDeletions := 0
	for _, commit := range branchCommits {
		totalAdditions += commit.Additions
		totalDeletions += commit.Deletions
	}

	// Analyze authors
	mainAuthors := a.analyzeBranchAuthors(branchCommits)

	// Determine activity level
	activityLevel := a.determineActivityLevel(len(branchCommits), lifespanDays)

	// Determine status
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	isActive := lastCommit.After(thirtyDaysAgo)
	status := a.determineBranchStatus(branchName, isActive, lifespanDays)

	return &BranchInfo{
		Name:           branchName,
		Type:           a.determineBranchType(branchName),
		Status:         status,
		CommitCount:    len(branchCommits),
		FirstCommit:    firstCommit,
		LastCommit:     lastCommit,
		LastCommitHash: lastCommitHash,
		LifespanDays:   lifespanDays,
		IsActive:       isActive,
		MainAuthors:    mainAuthors,
		AdditionsTotal: totalAdditions,
		DeletionsTotal: totalDeletions,
		ActivityLevel:  activityLevel,
	}
}

// analyzeSingleBranchFallback is the fallback method when we can't get actual branch info
func (a *Analyzer) analyzeSingleBranchFallback(branchName string, allCommits []git.GitCommit) *BranchInfo {
	// For performance, we'll use a simplified approach:
	// - Main/master branch gets most commits
	// - Other branches get estimated based on naming patterns
	
	var branchCommits []git.GitCommit
	
	// Simple heuristic: if it's main/master, assign most commits
	if branchName == "main" || branchName == "master" {
		branchCommits = allCommits
	} else {
		// For other branches, we'll estimate based on commit messages and patterns
		branchCommits = a.estimateBranchCommits(branchName, allCommits)
	}

	if len(branchCommits) == 0 {
		return &BranchInfo{
			Name:         branchName,
			Type:         a.determineBranchType(branchName),
			Status:       "empty",
			CommitCount:  0,
			IsActive:     false,
			MainAuthors:  make([]BranchAuthor, 0),
		}
	}

	// Sort commits by date
	sortedCommits := make([]git.GitCommit, len(branchCommits))
	copy(sortedCommits, branchCommits)
	sort.Slice(sortedCommits, func(i, j int) bool {
		return sortedCommits[i].Date.Before(sortedCommits[j].Date)
	})

	firstCommit := sortedCommits[0].Date
	lastCommit := sortedCommits[len(sortedCommits)-1].Date
	lastCommitHash := sortedCommits[len(sortedCommits)-1].Hash
	lifespanDays := int(lastCommit.Sub(firstCommit).Hours() / 24)

	// Calculate totals
	totalAdditions := 0
	totalDeletions := 0
	for _, commit := range branchCommits {
		totalAdditions += commit.Additions
		totalDeletions += commit.Deletions
	}

	// Analyze authors
	mainAuthors := a.analyzeBranchAuthors(branchCommits)

	// Determine activity level
	activityLevel := a.determineActivityLevel(len(branchCommits), lifespanDays)

	// Determine status
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	isActive := lastCommit.After(thirtyDaysAgo)
	status := a.determineBranchStatus(branchName, isActive, lifespanDays)

	return &BranchInfo{
		Name:           branchName,
		Type:           a.determineBranchType(branchName),
		Status:         status,
		CommitCount:    len(branchCommits),
		FirstCommit:    firstCommit,
		LastCommit:     lastCommit,
		LastCommitHash: lastCommitHash,
		LifespanDays:   lifespanDays,
		IsActive:       isActive,
		MainAuthors:    mainAuthors,
		AdditionsTotal: totalAdditions,
		DeletionsTotal: totalDeletions,
		ActivityLevel:  activityLevel,
	}
}

// Helper methods for branch analysis
func (a *Analyzer) estimateBranchCommits(branchName string, allCommits []git.GitCommit) []git.GitCommit {
	var branchCommits []git.GitCommit
	
	// Simple pattern matching based on branch name and commit messages
	branchNameLower := strings.ToLower(branchName)
	
	for _, commit := range allCommits {
		messageLower := strings.ToLower(commit.Message)
		
		// Basic heuristics for branch assignment
		if strings.Contains(messageLower, branchNameLower) ||
		   strings.Contains(messageLower, "feat") && strings.Contains(branchNameLower, "feat") ||
		   strings.Contains(messageLower, "fix") && strings.Contains(branchNameLower, "fix") ||
		   strings.Contains(messageLower, "hotfix") && strings.Contains(branchNameLower, "hotfix") {
			branchCommits = append(branchCommits, commit)
		}
	}
	
	// If no matches found and it's a recent branch, give it some recent commits
	if len(branchCommits) == 0 && len(allCommits) > 0 {
		// Give recent branches some recent commits
		recentCount := minInt(5, len(allCommits)/10) // Up to 5 commits or 10% of total
		if recentCount > 0 {
			branchCommits = allCommits[:recentCount]
		}
	}
	
	return branchCommits
}

func (a *Analyzer) analyzeBranchAuthors(commits []git.GitCommit) []BranchAuthor {
	authorCount := make(map[string]map[string]int) // email -> name -> count
	totalCommits := len(commits)
	
	for _, commit := range commits {
		if authorCount[commit.Email] == nil {
			authorCount[commit.Email] = make(map[string]int)
		}
		authorCount[commit.Email][commit.Author]++
	}
	
	// Convert to slice and sort
	var authors []BranchAuthor
	for email, nameMap := range authorCount {
		var bestName string
		var commitCount int
		for name, count := range nameMap {
			commitCount += count
			if bestName == "" || len(name) > len(bestName) {
				bestName = name
			}
		}
		
		percentage := float64(commitCount) / float64(totalCommits) * 100
		authors = append(authors, BranchAuthor{
			Name:        bestName,
			Email:       email,
			CommitCount: commitCount,
			Percentage:  percentage,
		})
	}
	
	// Sort by commit count
	sort.Slice(authors, func(i, j int) bool {
		return authors[i].CommitCount > authors[j].CommitCount
	})
	
	// Return top 3 authors
	if len(authors) > 3 {
		authors = authors[:3]
	}
	
	return authors
}

func (a *Analyzer) determineBranchType(branchName string) string {
	lower := strings.ToLower(branchName)
	
	if lower == "main" || lower == "master" {
		return "main"
	}
	if strings.HasPrefix(lower, "feature/") || strings.HasPrefix(lower, "feat/") {
		return "feature"
	}
	if strings.HasPrefix(lower, "release/") || strings.HasPrefix(lower, "rel/") {
		return "release"
	}
	if strings.HasPrefix(lower, "hotfix/") || strings.HasPrefix(lower, "fix/") {
		return "hotfix"
	}
	if strings.HasPrefix(lower, "develop") || strings.HasPrefix(lower, "dev") {
		return "develop"
	}
	
	return "other"
}

func (a *Analyzer) determineBranchStatus(branchName string, isActive bool, lifespanDays int) string {
	if isActive {
		return "active"
	}
	
	if lifespanDays < 7 {
		return "short-lived"
	}
	
	if lifespanDays > 180 {
		return "stale"
	}
	
	return "merged"
}

func (a *Analyzer) determineActivityLevel(commitCount, lifespanDays int) string {
	if lifespanDays == 0 {
		lifespanDays = 1
	}
	
	commitsPerDay := float64(commitCount) / float64(lifespanDays)
	
	if commitsPerDay >= 1.0 {
		return "high"
	}
	if commitsPerDay >= 0.1 {
		return "medium"
	}
	
	return "low"
}

func (a *Analyzer) calculateBranchSummary(branchStats map[string]*BranchInfo) BranchSummary {
	summary := BranchSummary{}
	now := time.Now()
	
	// 计算时间边界
	analysisStartTime := now.AddDate(0, -a.analysisMonths, 0)
	recentStartTime := now.AddDate(0, 0, -15) // 最近半个月
	
	recentActiveBranches := make([]BranchInfo, 0)
	
	for _, branch := range branchStats {
		summary.TotalBranches++
		
		// 检查是否在指定时间范围内活跃
		if branch.LastCommit.After(analysisStartTime) {
			summary.ActiveBranchesInRange++
		}
		
		// 检查是否在最近半个月内活跃
		if branch.LastCommit.After(recentStartTime) {
			recentActiveBranches = append(recentActiveBranches, *branch)
		}
	}
	
	// 按最后提交时间排序最近活跃分支
	sort.Slice(recentActiveBranches, func(i, j int) bool {
		return recentActiveBranches[i].LastCommit.After(recentActiveBranches[j].LastCommit)
	})
	
	summary.RecentActiveBranches = recentActiveBranches
	
	return summary
}

func (a *Analyzer) analyzeLifecyclePatterns(branchStats map[string]*BranchInfo, commits []git.GitCommit) LifecycleStats {
	stats := LifecycleStats{
		ShortLivedBranches: make([]BranchInfo, 0),
		LongLivedBranches:  make([]BranchInfo, 0),
		MergePatterns:      make([]MergePattern, 0),
		BranchingFrequency: make(map[string]int),
	}
	
	// Categorize branches by lifespan
	for _, branch := range branchStats {
		if branch.LifespanDays <= 7 && branch.CommitCount > 0 {
			stats.ShortLivedBranches = append(stats.ShortLivedBranches, *branch)
		}
		if branch.LifespanDays >= 90 {
			stats.LongLivedBranches = append(stats.LongLivedBranches, *branch)
		}
	}
	
	// Analyze merge patterns from commit messages
	for _, commit := range commits {
		if strings.Contains(strings.ToLower(commit.Message), "merge") && len(commit.Parents) > 1 {
			// This is a merge commit
			pattern := MergePattern{
				FromBranch:   "feature",  // Simplified
				ToBranch:     "main",     // Simplified  
				MergeCommit:  commit.Hash,
				MergeDate:    commit.Date,
				ConflictRisk: "low",      // Simplified
			}
			stats.MergePatterns = append(stats.MergePatterns, pattern)
		}
	}
	
	// Calculate branching frequency by month
	for _, commit := range commits {
		monthKey := commit.Date.Format("2006-01")
		stats.BranchingFrequency[monthKey]++
	}
	
	return stats
}

// Helper function to get minimum of two integers
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
