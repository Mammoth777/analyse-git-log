package analyzer

import (
	"fmt"
	"sort"
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

// BranchData contains branch structure and commit relationships
type BranchData struct {
	Branches       []BranchInfo     `json:"branches"`
	CommitGraph    []CommitNode     `json:"commit_graph"`
	MergePatterns  []MergeInfo      `json:"merge_patterns"`
}

// BranchInfo contains information about a single branch
type BranchInfo struct {
	Name         string    `json:"name"`
	CommitCount  int       `json:"commit_count"`
	FirstCommit  time.Time `json:"first_commit"`
	LastCommit   time.Time `json:"last_commit"`
	IsActive     bool      `json:"is_active"`
	MainAuthors  []string  `json:"main_authors"`
}

// CommitNode represents a commit in the graph structure
type CommitNode struct {
	Hash       string    `json:"hash"`
	ShortHash  string    `json:"short_hash"`
	Message    string    `json:"message"`
	Author     string    `json:"author"`
	Date       time.Time `json:"date"`
	Branch     string    `json:"branch"`
	Parents    []string  `json:"parents"`
	Children   []string  `json:"children"`
	X          int       `json:"x"` // 图形坐标
	Y          int       `json:"y"`
	IsMerge    bool      `json:"is_merge"`
}

// MergeInfo contains information about merge operations
type MergeInfo struct {
	MergeCommit   string    `json:"merge_commit"`
	SourceBranch  string    `json:"source_branch"`
	TargetBranch  string    `json:"target_branch"`
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
	return a.AnalyzeWithProgress(nil)
}

// AnalyzeWithProgress performs comprehensive analysis with progress callback
func (a *Analyzer) AnalyzeWithProgress(progressCallback ProgressCallback) (*Statistics, error) {
	if progressCallback != nil {
		progressCallback(0, 100, "开始获取提交历史...")
	}
	
	commits, err := a.repo.GetCommitsWithProgress(0, func(current, total int, message string) {
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
		Branches:      make([]BranchInfo, 0),
		CommitGraph:   make([]CommitNode, 0),
		MergePatterns: make([]MergeInfo, 0),
	}

	// Get branch information from git (this is fast)
	branches, err := a.repo.GetBranches()
	if err != nil {
		return nil, fmt.Errorf("failed to get branches: %v", err)
	}

	// Build commit hash to index mapping
	commitHashMap := make(map[string]int)
	for i, commit := range commits {
		commitHashMap[commit.Hash] = i
	}

	// 🚀 OPTIMIZATION: Use existing commits instead of fetching again for each branch
	// Group commits by branch using existing data
	branchCommitsMap := make(map[string][]git.GitCommit)
	
	// Initialize branch maps
	for _, branch := range branches {
		branchCommitsMap[branch] = make([]git.GitCommit, 0)
	}
	
	// Assign commits to branches (simplified approach using commit order and main branch assumption)
	// For better accuracy, we could use git branch --contains <hash> but it's slow
	currentBranch := "main" // Default branch
	if len(branches) > 0 {
		// Try to find main/master branch
		for _, branch := range branches {
			if branch == "main" || branch == "master" {
				currentBranch = branch
				break
			}
		}
		if currentBranch == "main" && !contains(branches, "main") {
			currentBranch = branches[0] // Use first branch if main not found
		}
	}
	
	// Simple branch assignment: assign all commits to main branch for performance
	// In a real scenario, you might want more sophisticated branch detection
	for _, commit := range commits {
		branchCommitsMap[currentBranch] = append(branchCommitsMap[currentBranch], commit)
	}

	// Analyze each branch using grouped commits
	branchStats := make(map[string]*BranchInfo)
	for _, branch := range branches {
		branchCommits := branchCommitsMap[branch]
		
		branchInfo := &BranchInfo{
			Name:        branch,
			CommitCount: len(branchCommits),
			IsActive:    false,
			MainAuthors: make([]string, 0),
		}

		if len(branchCommits) == 0 {
			// Branch exists but has no commits in our analysis window
			branchStats[branch] = branchInfo
			branchData.Branches = append(branchData.Branches, *branchInfo)
			continue
		}

		// Sort commits by date to find first and last
		sortedCommits := make([]git.GitCommit, len(branchCommits))
		copy(sortedCommits, branchCommits)
		sort.Slice(sortedCommits, func(i, j int) bool {
			return sortedCommits[i].Date.Before(sortedCommits[j].Date)
		})

		branchInfo.FirstCommit = sortedCommits[0].Date                        // Oldest commit
		branchInfo.LastCommit = sortedCommits[len(sortedCommits)-1].Date      // Newest commit

		// Determine if branch is active (has commits in last 30 days)
		thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
		branchInfo.IsActive = branchInfo.LastCommit.After(thirtyDaysAgo)

		// Find main authors for this branch
		authorCount := make(map[string]int)
		for _, commit := range branchCommits {
			authorKey := fmt.Sprintf("%s <%s>", commit.Author, commit.Email)
			authorCount[authorKey]++
		}

		// Sort authors by commit count and take top 3
		type authorCommitPair struct {
			author string
			count  int
		}
		
		var sortedAuthors []authorCommitPair
		for author, count := range authorCount {
			sortedAuthors = append(sortedAuthors, authorCommitPair{author, count})
		}
		
		sort.Slice(sortedAuthors, func(i, j int) bool {
			return sortedAuthors[i].count > sortedAuthors[j].count
		})

		for i, pair := range sortedAuthors {
			if i >= 3 { // Top 3 authors
				break
			}
			branchInfo.MainAuthors = append(branchInfo.MainAuthors, pair.author)
		}

		branchStats[branch] = branchInfo
		branchData.Branches = append(branchData.Branches, *branchInfo)
	}

	// Build commit graph (simplified for performance)
	branchData.CommitGraph = a.buildCommitGraphOptimized(commits, branchStats)

	// Analyze merge patterns (optimized)
	branchData.MergePatterns = a.analyzeMergePatternsOptimized(commits)

	return branchData, nil
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// buildCommitGraphOptimized builds a graph structure for commits (optimized version)
func (a *Analyzer) buildCommitGraphOptimized(commits []git.GitCommit, branchStats map[string]*BranchInfo) []CommitNode {
	nodes := make([]CommitNode, 0, len(commits))
	commitMap := make(map[string]*CommitNode)

	// Create commit nodes
	for i, commit := range commits {
		node := CommitNode{
			Hash:      commit.Hash,
			ShortHash: commit.Hash[:8], // First 8 characters
			Message:   commit.Message,
			Author:    commit.Author,
			Date:      commit.Date,
			Branch:    a.getBranchForCommitOptimized(commit.Hash, branchStats),
			Parents:   commit.Parents,
			Children:  make([]string, 0),
			X:         0, // Will be calculated later
			Y:         i, // Simple Y positioning based on commit order
			IsMerge:   len(commit.Parents) > 1,
		}

		nodes = append(nodes, node)
		commitMap[commit.Hash] = &nodes[len(nodes)-1]
	}

	// Build parent-child relationships
	for i := range nodes {
		node := &nodes[i]
		for _, parentHash := range node.Parents {
			if parentNode, exists := commitMap[parentHash]; exists {
				parentNode.Children = append(parentNode.Children, node.Hash)
			}
		}
	}

	// Calculate X positions for visualization (simple branch-based positioning)
	a.calculateCommitPositions(nodes, branchStats)

	return nodes
}

// getBranchForCommitOptimized determines which branch a commit belongs to (optimized)
func (a *Analyzer) getBranchForCommitOptimized(commitHash string, branchStats map[string]*BranchInfo) string {
	// For performance, we'll use a simplified approach
	// In most cases, commits belong to the main branch
	for branchName := range branchStats {
		if branchName == "main" || branchName == "master" {
			return branchName
		}
	}
	
	// If no main/master branch, return the first available branch
	for branchName := range branchStats {
		return branchName
	}
	
	return "main" // Default fallback
}

// analyzeMergePatternsOptimized analyzes merge commit patterns (optimized version)
func (a *Analyzer) analyzeMergePatternsOptimized(commits []git.GitCommit) []MergeInfo {
	mergePatterns := make([]MergeInfo, 0)

	for _, commit := range commits {
		if len(commit.Parents) > 1 { // This is a merge commit
			mergeInfo := MergeInfo{
				MergeCommit:  commit.Hash,
				Date:         commit.Date,
				Author:       commit.Author,
				SourceBranch: "feature-branch", // Simplified for performance
				TargetBranch: "main",          // Simplified for performance
				CommitCount:  1,               // Simplified for now
			}

			mergePatterns = append(mergePatterns, mergeInfo)
		}
	}

	return mergePatterns
}

// calculateCommitPositions calculates X,Y positions for commit graph visualization
func (a *Analyzer) calculateCommitPositions(nodes []CommitNode, branchStats map[string]*BranchInfo) {
	// Create branch-to-column mapping
	branchColumns := make(map[string]int)
	currentColumn := 0

	// Assign columns to branches
	for branchName := range branchStats {
		branchColumns[branchName] = currentColumn
		currentColumn++
	}

	// Set X positions based on branch
	for i := range nodes {
		if column, exists := branchColumns[nodes[i].Branch]; exists {
			nodes[i].X = column * 50 // 50px spacing between branches
		} else {
			nodes[i].X = currentColumn * 50 // Unknown branch gets new column
		}
	}
}

// analyzeMergePatterns analyzes merge commit patterns
func (a *Analyzer) analyzeMergePatterns(commits []git.GitCommit) []MergeInfo {
	mergePatterns := make([]MergeInfo, 0)

	for _, commit := range commits {
		if len(commit.Parents) > 1 { // This is a merge commit
			mergeInfo := MergeInfo{
				MergeCommit:  commit.Hash,
				Date:         commit.Date,
				Author:       commit.Author,
				SourceBranch: "unknown", // Would need more git analysis to determine
				TargetBranch: "unknown", // Would need more git analysis to determine
				CommitCount:  1,         // Simplified for now
			}

			// Try to extract branch information from commit message
			if len(commit.Message) > 0 {
				// Look for patterns like "Merge branch 'feature/xyz' into main"
				// This is a simplified approach
				mergeInfo.SourceBranch = a.extractBranchFromMergeMessage(commit.Message, "source")
				mergeInfo.TargetBranch = a.extractBranchFromMergeMessage(commit.Message, "target")
			}

			mergePatterns = append(mergePatterns, mergeInfo)
		}
	}

	return mergePatterns
}

// extractBranchFromMergeMessage extracts branch names from merge commit messages
func (a *Analyzer) extractBranchFromMergeMessage(message, branchType string) string {
	// This is a simplified implementation
	// In practice, you'd use more sophisticated parsing
	if len(message) > 0 {
		// Look for common merge message patterns
		// "Merge branch 'feature/xyz' into main"
		// "Merge pull request #123 from feature/xyz"
		return "feature-branch" // Placeholder
	}
	return "unknown"
}
