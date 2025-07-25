package developer

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"git-log-analyzer/internal/analyzer"
)

// DeveloperProfile represents a developer's work style profile
type DeveloperProfile struct {
	Name                string                 `json:"name"`
	Email               string                 `json:"email"`
	WorkStyleMetrics    WorkStyleMetrics       `json:"work_style_metrics"`
	CodingPatterns      CodingPatterns         `json:"coding_patterns"`
	CollaborationStyle  CollaborationStyle     `json:"collaboration_style"`
	TimeManagement      TimeManagement         `json:"time_management"`
	QualityIndicators   QualityIndicators      `json:"quality_indicators"`
	TechnicalProfile    TechnicalProfile       `json:"technical_profile"`
	PersonalityTraits   PersonalityTraits      `json:"personality_traits"`
}

// WorkStyleMetrics contains metrics about work style
type WorkStyleMetrics struct {
	CommitFrequency     float64 `json:"commit_frequency"`     // commits per day
	AverageCommitSize   float64 `json:"average_commit_size"`  // lines changed per commit
	WorkSessionLength   float64 `json:"work_session_length"`  // average hours between commits
	ConsistencyScore    float64 `json:"consistency_score"`    // 0-100, how consistent is the work pattern
	BurstWorkRatio      float64 `json:"burst_work_ratio"`     // ratio of work done in concentrated bursts
}

// CodingPatterns contains coding behavior patterns
type CodingPatterns struct {
	PreferredCommitSize string  `json:"preferred_commit_size"` // "atomic", "moderate", "bulk"
	RefactoringTendency float64 `json:"refactoring_tendency"`  // ratio of refactoring commits
	BugFixRatio         float64 `json:"bug_fix_ratio"`         // ratio of bug fix commits
	FeatureFocusRatio   float64 `json:"feature_focus_ratio"`   // ratio of feature commits
	DocumentationRatio  float64 `json:"documentation_ratio"`  // ratio of documentation commits
	TestingEngagement   float64 `json:"testing_engagement"`   // ratio of test-related commits
}

// CollaborationStyle contains collaboration patterns
type CollaborationStyle struct {
	FilesOwnershipRatio float64  `json:"files_ownership_ratio"` // ratio of files primarily worked on
	CrossTeamWork       float64  `json:"cross_team_work"`       // ratio of work on others' files
	SpecializationLevel float64  `json:"specialization_level"`  // how specialized vs generalist
	MentorshipLevel     string   `json:"mentorship_level"`      // "mentor", "peer", "learner"
	PreferredFileTypes  []string `json:"preferred_file_types"`  // most worked file extensions
}

// TimeManagement contains time-related work patterns
type TimeManagement struct {
	PreferredWorkHours []int   `json:"preferred_work_hours"` // hours of day (0-23)
	WeekendWorker      bool    `json:"weekend_worker"`       // works on weekends
	NightOwl           bool    `json:"night_owl"`            // works late hours
	EarlyBird          bool    `json:"early_bird"`           // works early hours
	WorkLifeBalance    float64 `json:"work_life_balance"`    // 0-100, based on work time distribution
}

// QualityIndicators contains code quality related metrics
type QualityIndicators struct {
	CommitMessageQuality float64 `json:"commit_message_quality"` // 0-100, based on message informativeness
	CodeStabilityScore   float64 `json:"code_stability_score"`   // 0-100, based on how often code changes again
	TechnicalDebtRatio   float64 `json:"technical_debt_ratio"`   // ratio of commits that might introduce debt
	ReviewAttentiveness  float64 `json:"review_attentiveness"`   // estimated from commit patterns
}

// TechnicalProfile contains technical skill patterns
type TechnicalProfile struct {
	PrimaryLanguages    []string `json:"primary_languages"`     // most used programming languages
	TechnologyStack     []string `json:"technology_stack"`      // inferred tech stack
	ArchitecturalFocus  string   `json:"architectural_focus"`   // "frontend", "backend", "fullstack", "devops"
	LearningVelocity    float64  `json:"learning_velocity"`     // how quickly adopts new technologies
	InnovationTendency  float64  `json:"innovation_tendency"`   // tendency to try new approaches
}

// PersonalityTraits contains inferred personality traits
type PersonalityTraits struct {
	WorkStyleType       string  `json:"work_style_type"`       // "steady", "burst", "balanced"
	PlanningOrientation string  `json:"planning_orientation"`  // "planner", "adaptive", "reactive"
	RiskTolerance       string  `json:"risk_tolerance"`        // "conservative", "moderate", "aggressive"
	DetailOrientation   string  `json:"detail_orientation"`    // "high", "medium", "low"
	CollaborationStyle  string  `json:"collaboration_style"`   // "independent", "collaborative", "leader"
	PerfectionismLevel  float64 `json:"perfectionism_level"`   // 0-100, based on commit patterns
}

// ProfileAnalyzer analyzes developer profiles
type ProfileAnalyzer struct {
	stats *analyzer.Statistics
}

// NewProfileAnalyzer creates a new profile analyzer
func NewProfileAnalyzer(stats *analyzer.Statistics) *ProfileAnalyzer {
	return &ProfileAnalyzer{stats: stats}
}

// AnalyzeAllDevelopers analyzes all developers and returns their profiles
func (pa *ProfileAnalyzer) AnalyzeAllDevelopers() []*DeveloperProfile {
	var profiles []*DeveloperProfile
	
	for _, authorStat := range pa.stats.AuthorStats {
		profile := pa.AnalyzeDeveloper(authorStat)
		profiles = append(profiles, profile)
	}
	
	// Sort by commit count (most active first)
	sort.Slice(profiles, func(i, j int) bool {
		return profiles[i].WorkStyleMetrics.CommitFrequency > profiles[j].WorkStyleMetrics.CommitFrequency
	})
	
	return profiles
}

// AnalyzeDeveloper analyzes a single developer
func (pa *ProfileAnalyzer) AnalyzeDeveloper(authorStat *analyzer.AuthorStat) *DeveloperProfile {
	profile := &DeveloperProfile{
		Name:  authorStat.Name,
		Email: authorStat.Email,
	}
	
	// Analyze different aspects
	profile.WorkStyleMetrics = pa.analyzeWorkStyle(authorStat)
	profile.CodingPatterns = pa.analyzeCodingPatterns(authorStat)
	profile.CollaborationStyle = pa.analyzeCollaborationStyle(authorStat)
	profile.TimeManagement = pa.analyzeTimeManagement(authorStat)
	profile.QualityIndicators = pa.analyzeQualityIndicators(authorStat)
	profile.TechnicalProfile = pa.analyzeTechnicalProfile(authorStat)
	profile.PersonalityTraits = pa.analyzePersonalityTraits(profile)
	
	return profile
}

// GenerateReport generates a text report for the developer profile
func (dp *DeveloperProfile) GenerateReport() string {
	var report strings.Builder
	
	report.WriteString(fmt.Sprintf("=== 开发者风格画像: %s ===\n", dp.Name))
	report.WriteString(fmt.Sprintf("邮箱: %s\n", dp.Email))
	report.WriteString("\n")
	
	// Work Style Metrics
	report.WriteString("🎯 工作风格分析:\n")
	report.WriteString(fmt.Sprintf("  提交频率: %.2f (commits/day)\n", dp.WorkStyleMetrics.CommitFrequency))
	report.WriteString(fmt.Sprintf("  平均提交大小: %.1f 行\n", dp.WorkStyleMetrics.AverageCommitSize))
	report.WriteString(fmt.Sprintf("  工作一致性: %.1f/100\n", dp.WorkStyleMetrics.ConsistencyScore))
	report.WriteString(fmt.Sprintf("  爆发性工作比例: %.1f%%\n", dp.WorkStyleMetrics.BurstWorkRatio*100))
	report.WriteString("\n")
	
	// Coding Patterns
	report.WriteString("💻 编码模式:\n")
	report.WriteString(fmt.Sprintf("  偏好提交大小: %s\n", dp.CodingPatterns.PreferredCommitSize))
	report.WriteString(fmt.Sprintf("  重构倾向: %.1f%%\n", dp.CodingPatterns.RefactoringTendency*100))
	report.WriteString(fmt.Sprintf("  Bug修复比例: %.1f%%\n", dp.CodingPatterns.BugFixRatio*100))
	report.WriteString(fmt.Sprintf("  功能开发比例: %.1f%%\n", dp.CodingPatterns.FeatureFocusRatio*100))
	report.WriteString(fmt.Sprintf("  文档编写比例: %.1f%%\n", dp.CodingPatterns.DocumentationRatio*100))
	report.WriteString(fmt.Sprintf("  测试参与度: %.1f%%\n", dp.CodingPatterns.TestingEngagement*100))
	report.WriteString("\n")
	
	// Collaboration Style
	report.WriteString("🤝 协作风格:\n")
	report.WriteString(fmt.Sprintf("  文件所有权比例: %.1f%%\n", dp.CollaborationStyle.FilesOwnershipRatio*100))
	report.WriteString(fmt.Sprintf("  跨团队工作: %.1f%%\n", dp.CollaborationStyle.CrossTeamWork*100))
	report.WriteString(fmt.Sprintf("  专业化程度: %.1f/10\n", dp.CollaborationStyle.SpecializationLevel))
	report.WriteString(fmt.Sprintf("  指导水平: %s\n", dp.CollaborationStyle.MentorshipLevel))
	if len(dp.CollaborationStyle.PreferredFileTypes) > 0 {
		report.WriteString(fmt.Sprintf("  偏好文件类型: %s\n", strings.Join(dp.CollaborationStyle.PreferredFileTypes, ", ")))
	}
	report.WriteString("\n")
	
	// Technical Profile
	report.WriteString("🔧 技术特征:\n")
	if len(dp.TechnicalProfile.PrimaryLanguages) > 0 {
		report.WriteString(fmt.Sprintf("  主要编程语言: %s\n", strings.Join(dp.TechnicalProfile.PrimaryLanguages, ", ")))
	}
	report.WriteString(fmt.Sprintf("  架构专注: %s\n", dp.TechnicalProfile.ArchitecturalFocus))
	report.WriteString(fmt.Sprintf("  学习速度: %.1f/10\n", dp.TechnicalProfile.LearningVelocity))
	report.WriteString(fmt.Sprintf("  创新倾向: %.1f/10\n", dp.TechnicalProfile.InnovationTendency))
	report.WriteString("\n")
	
	// Personality Traits
	report.WriteString("🧠 个性特征:\n")
	report.WriteString(fmt.Sprintf("  工作风格类型: %s\n", dp.PersonalityTraits.WorkStyleType))
	report.WriteString(fmt.Sprintf("  计划导向: %s\n", dp.PersonalityTraits.PlanningOrientation))
	report.WriteString(fmt.Sprintf("  风险容忍度: %s\n", dp.PersonalityTraits.RiskTolerance))
	report.WriteString(fmt.Sprintf("  细节导向: %s\n", dp.PersonalityTraits.DetailOrientation))
	report.WriteString(fmt.Sprintf("  完美主义程度: %.1f/100\n", dp.PersonalityTraits.PerfectionismLevel))
	report.WriteString("\n")
	
	return report.String()
}

// analyzeWorkStyle analyzes work style metrics
func (pa *ProfileAnalyzer) analyzeWorkStyle(authorStat *analyzer.AuthorStat) WorkStyleMetrics {
	totalDays := pa.stats.TimeStats.LastCommit.Sub(pa.stats.TimeStats.FirstCommit).Hours() / 24
	if totalDays == 0 {
		totalDays = 1
	}
	
	commitFrequency := float64(authorStat.CommitCount) / totalDays
	
	// Calculate average commit size
	averageCommitSize := float64(authorStat.Additions+authorStat.Deletions) / float64(authorStat.CommitCount)
	
	// Analyze work session patterns (simplified)
	workSessionLength := pa.calculateWorkSessionLength(authorStat)
	consistencyScore := pa.calculateConsistencyScore(authorStat)
	burstWorkRatio := pa.calculateBurstWorkRatio(authorStat)
	
	return WorkStyleMetrics{
		CommitFrequency:   commitFrequency,
		AverageCommitSize: averageCommitSize,
		WorkSessionLength: workSessionLength,
		ConsistencyScore:  consistencyScore,
		BurstWorkRatio:    burstWorkRatio,
	}
}

// analyzeCodingPatterns analyzes coding behavior patterns
func (pa *ProfileAnalyzer) analyzeCodingPatterns(authorStat *analyzer.AuthorStat) CodingPatterns {
	// Determine preferred commit size
	averageSize := float64(authorStat.Additions+authorStat.Deletions) / float64(authorStat.CommitCount)
	var preferredSize string
	switch {
	case averageSize < 20:
		preferredSize = "atomic"
	case averageSize < 100:
		preferredSize = "moderate"
	default:
		preferredSize = "bulk"
	}
	
	// Analyze commit message patterns to infer ratios (simplified for now)
	refactoringTendency := pa.estimateCommitTypeRatio(authorStat, []string{"refactor", "refact", "clean", "improve"})
	bugFixRatio := pa.estimateCommitTypeRatio(authorStat, []string{"fix", "bug", "issue", "error"})
	featureFocusRatio := pa.estimateCommitTypeRatio(authorStat, []string{"add", "feat", "feature", "implement"})
	documentationRatio := pa.estimateCommitTypeRatio(authorStat, []string{"doc", "readme", "comment", "documentation"})
	testingEngagement := pa.estimateCommitTypeRatio(authorStat, []string{"test", "spec", "coverage"})
	
	return CodingPatterns{
		PreferredCommitSize: preferredSize,
		RefactoringTendency: refactoringTendency,
		BugFixRatio:         bugFixRatio,
		FeatureFocusRatio:   featureFocusRatio,
		DocumentationRatio:  documentationRatio,
		TestingEngagement:   testingEngagement,
	}
}

// analyzeCollaborationStyle analyzes collaboration patterns
func (pa *ProfileAnalyzer) analyzeCollaborationStyle(authorStat *analyzer.AuthorStat) CollaborationStyle {
	// Calculate file ownership ratio (simplified)
	filesOwnershipRatio := pa.calculateFileOwnershipRatio(authorStat)
	crossTeamWork := 1.0 - filesOwnershipRatio // Inverse relationship
	
	// Calculate specialization level based on file type diversity
	specializationLevel := pa.calculateSpecializationLevel(authorStat)
	
	// Determine mentorship level based on commit patterns and activity
	mentorshipLevel := pa.determineMentorshipLevel(authorStat)
	
	// Get preferred file types
	preferredFileTypes := pa.getPreferredFileTypes(authorStat)
	
	return CollaborationStyle{
		FilesOwnershipRatio: filesOwnershipRatio,
		CrossTeamWork:       crossTeamWork,
		SpecializationLevel: specializationLevel,
		MentorshipLevel:     mentorshipLevel,
		PreferredFileTypes:  preferredFileTypes,
	}
}

// analyzeTimeManagement analyzes time-related work patterns
func (pa *ProfileAnalyzer) analyzeTimeManagement(authorStat *analyzer.AuthorStat) TimeManagement {
	// This is a simplified implementation
	// In a real implementation, you'd analyze actual commit timestamps
	
	preferredWorkHours := []int{9, 10, 11, 14, 15, 16} // Default business hours
	weekendWorker := false
	nightOwl := false
	earlyBird := false
	workLifeBalance := 75.0 // Default reasonable balance
	
	return TimeManagement{
		PreferredWorkHours: preferredWorkHours,
		WeekendWorker:      weekendWorker,
		NightOwl:           nightOwl,
		EarlyBird:          earlyBird,
		WorkLifeBalance:    workLifeBalance,
	}
}

// analyzeQualityIndicators analyzes code quality indicators
func (pa *ProfileAnalyzer) analyzeQualityIndicators(authorStat *analyzer.AuthorStat) QualityIndicators {
	// Estimate commit message quality (simplified)
	commitMessageQuality := pa.estimateCommitMessageQuality(authorStat)
	
	// Calculate code stability score
	codeStabilityScore := pa.calculateCodeStabilityScore(authorStat)
	
	// Estimate technical debt ratio
	technicalDebtRatio := pa.estimateTechnicalDebtRatio(authorStat)
	
	// Estimate review attentiveness
	reviewAttentiveness := pa.estimateReviewAttentiveness(authorStat)
	
	return QualityIndicators{
		CommitMessageQuality: commitMessageQuality,
		CodeStabilityScore:   codeStabilityScore,
		TechnicalDebtRatio:   technicalDebtRatio,
		ReviewAttentiveness:  reviewAttentiveness,
	}
}

// analyzeTechnicalProfile analyzes technical skill patterns
func (pa *ProfileAnalyzer) analyzeTechnicalProfile(authorStat *analyzer.AuthorStat) TechnicalProfile {
	// This would be enhanced with actual file analysis
	primaryLanguages := []string{"Go", "JavaScript", "TypeScript"} // Default for this project
	technologyStack := []string{"Git", "Web", "Backend"}
	architecturalFocus := "fullstack"
	learningVelocity := 60.0
	innovationTendency := 50.0
	
	return TechnicalProfile{
		PrimaryLanguages:   primaryLanguages,
		TechnologyStack:    technologyStack,
		ArchitecturalFocus: architecturalFocus,
		LearningVelocity:   learningVelocity,
		InnovationTendency: innovationTendency,
	}
}

// analyzePersonalityTraits infers personality traits from other metrics
func (pa *ProfileAnalyzer) analyzePersonalityTraits(profile *DeveloperProfile) PersonalityTraits {
	// Determine work style type
	var workStyleType string
	if profile.WorkStyleMetrics.BurstWorkRatio > 0.7 {
		workStyleType = "burst"
	} else if profile.WorkStyleMetrics.ConsistencyScore > 70 {
		workStyleType = "steady"
	} else {
		workStyleType = "balanced"
	}
	
	// Determine planning orientation
	var planningOrientation string
	if profile.CodingPatterns.PreferredCommitSize == "atomic" {
		planningOrientation = "planner"
	} else if profile.CodingPatterns.PreferredCommitSize == "bulk" {
		planningOrientation = "reactive"
	} else {
		planningOrientation = "adaptive"
	}
	
	// Determine risk tolerance
	var riskTolerance string
	riskScore := profile.TechnicalProfile.InnovationTendency
	if riskScore > 70 {
		riskTolerance = "aggressive"
	} else if riskScore > 40 {
		riskTolerance = "moderate"
	} else {
		riskTolerance = "conservative"
	}
	
	// Determine detail orientation
	var detailOrientation string
	if profile.QualityIndicators.CommitMessageQuality > 80 {
		detailOrientation = "high"
	} else if profile.QualityIndicators.CommitMessageQuality > 50 {
		detailOrientation = "medium"
	} else {
		detailOrientation = "low"
	}
	
	// Determine collaboration style
	var collaborationStyle string
	if profile.CollaborationStyle.CrossTeamWork > 0.6 {
		collaborationStyle = "collaborative"
	} else if profile.CollaborationStyle.FilesOwnershipRatio > 0.7 {
		collaborationStyle = "independent"
	} else {
		collaborationStyle = "leader"
	}
	
	// Calculate perfectionism level
	perfectionismLevel := (profile.QualityIndicators.CommitMessageQuality + 
		profile.QualityIndicators.CodeStabilityScore + 
		(100 - profile.QualityIndicators.TechnicalDebtRatio)) / 3
	
	return PersonalityTraits{
		WorkStyleType:       workStyleType,
		PlanningOrientation: planningOrientation,
		RiskTolerance:       riskTolerance,
		DetailOrientation:   detailOrientation,
		CollaborationStyle:  collaborationStyle,
		PerfectionismLevel:  perfectionismLevel,
	}
}

// Helper methods for calculations

func (pa *ProfileAnalyzer) calculateWorkSessionLength(authorStat *analyzer.AuthorStat) float64 {
	// Simplified: assume average 4 hours per session
	return 4.0
}

func (pa *ProfileAnalyzer) calculateConsistencyScore(authorStat *analyzer.AuthorStat) float64 {
	// Simplified: base on commit frequency regularity
	if authorStat.CommitCount > 50 {
		return 80.0
	} else if authorStat.CommitCount > 20 {
		return 60.0
	}
	return 40.0
}

func (pa *ProfileAnalyzer) calculateBurstWorkRatio(authorStat *analyzer.AuthorStat) float64 {
	// Simplified: estimate based on commit size variation
	averageSize := float64(authorStat.Additions+authorStat.Deletions) / float64(authorStat.CommitCount)
	if averageSize > 100 {
		return 0.8 // High burst ratio for large commits
	}
	return 0.3 // Low burst ratio for small commits
}

func (pa *ProfileAnalyzer) estimateCommitTypeRatio(authorStat *analyzer.AuthorStat, keywords []string) float64 {
	// This is a placeholder - in real implementation, analyze actual commit messages
	// For now, return random-ish values based on author characteristics
	baseRatio := float64(len(keywords)) / 20.0 // Simple heuristic
	if baseRatio > 1.0 {
		baseRatio = 1.0
	}
	return baseRatio
}

func (pa *ProfileAnalyzer) calculateFileOwnershipRatio(authorStat *analyzer.AuthorStat) float64 {
	// Simplified: estimate based on commit concentration
	if authorStat.CommitCount > 100 {
		return 0.7 // High ownership for very active developers
	}
	return 0.4 // Lower ownership for less active developers
}

func (pa *ProfileAnalyzer) calculateSpecializationLevel(authorStat *analyzer.AuthorStat) float64 {
	// Simplified: assume moderate specialization
	return 0.6
}

func (pa *ProfileAnalyzer) determineMentorshipLevel(authorStat *analyzer.AuthorStat) string {
	if authorStat.CommitCount > 200 {
		return "mentor"
	} else if authorStat.CommitCount > 50 {
		return "peer"
	}
	return "learner"
}

func (pa *ProfileAnalyzer) getPreferredFileTypes(authorStat *analyzer.AuthorStat) []string {
	// Default file types for this project
	return []string{".go", ".md", ".js", ".html"}
}

func (pa *ProfileAnalyzer) estimateCommitMessageQuality(authorStat *analyzer.AuthorStat) float64 {
	// Simplified: estimate based on activity level
	if authorStat.CommitCount > 100 {
		return 75.0
	} else if authorStat.CommitCount > 20 {
		return 60.0
	}
	return 45.0
}

func (pa *ProfileAnalyzer) calculateCodeStabilityScore(authorStat *analyzer.AuthorStat) float64 {
	// Simplified: base on lines changed ratio
	ratio := float64(authorStat.Deletions) / float64(authorStat.Additions+1)
	return math.Max(20, 100-(ratio*100))
}

func (pa *ProfileAnalyzer) estimateTechnicalDebtRatio(authorStat *analyzer.AuthorStat) float64 {
	// Simplified: estimate based on change frequency
	changeRatio := float64(authorStat.Deletions) / float64(authorStat.Additions+1)
	return math.Min(50, changeRatio*100)
}

func (pa *ProfileAnalyzer) estimateReviewAttentiveness(authorStat *analyzer.AuthorStat) float64 {
	// Simplified: assume good review practices for active developers
	if authorStat.CommitCount > 50 {
		return 70.0
	}
	return 50.0
}

// GenerateProfileSummary generates a human-readable summary of a developer profile
func (profile *DeveloperProfile) GenerateProfileSummary() string {
	var summary strings.Builder
	
	summary.WriteString(fmt.Sprintf("## 👤 开发者风格画像: %s\n\n", profile.Name))
	
	// Work Style Summary
	summary.WriteString("### 🎯 工作风格特征\n")
	summary.WriteString(fmt.Sprintf("- **工作类型**: %s开发者\n", profile.PersonalityTraits.WorkStyleType))
	summary.WriteString(fmt.Sprintf("- **计划导向**: %s型\n", profile.PersonalityTraits.PlanningOrientation))
	summary.WriteString(fmt.Sprintf("- **协作风格**: %s\n", profile.PersonalityTraits.CollaborationStyle))
	summary.WriteString(fmt.Sprintf("- **提交频率**: %.2f 次/天\n", profile.WorkStyleMetrics.CommitFrequency))
	summary.WriteString(fmt.Sprintf("- **平均提交大小**: %.0f 行变更\n", profile.WorkStyleMetrics.AverageCommitSize))
	summary.WriteString(fmt.Sprintf("- **工作一致性**: %.0f%%\n", profile.WorkStyleMetrics.ConsistencyScore))
	summary.WriteString("\n")
	
	// Coding Patterns
	summary.WriteString("### 💻 编码模式\n")
	summary.WriteString(fmt.Sprintf("- **偏好提交大小**: %s\n", profile.CodingPatterns.PreferredCommitSize))
	summary.WriteString(fmt.Sprintf("- **重构倾向**: %.1f%%\n", profile.CodingPatterns.RefactoringTendency*100))
	summary.WriteString(fmt.Sprintf("- **Bug修复比例**: %.1f%%\n", profile.CodingPatterns.BugFixRatio*100))
	summary.WriteString(fmt.Sprintf("- **功能开发比例**: %.1f%%\n", profile.CodingPatterns.FeatureFocusRatio*100))
	summary.WriteString(fmt.Sprintf("- **测试参与度**: %.1f%%\n", profile.CodingPatterns.TestingEngagement*100))
	summary.WriteString("\n")
	
	// Quality Indicators
	summary.WriteString("### 📊 质量指标\n")
	summary.WriteString(fmt.Sprintf("- **提交消息质量**: %.0f/100\n", profile.QualityIndicators.CommitMessageQuality))
	summary.WriteString(fmt.Sprintf("- **代码稳定性**: %.0f/100\n", profile.QualityIndicators.CodeStabilityScore))
	summary.WriteString(fmt.Sprintf("- **完美主义倾向**: %.0f/100\n", profile.PersonalityTraits.PerfectionismLevel))
	summary.WriteString("\n")
	
	// Technical Profile
	summary.WriteString("### 🔧 技术特征\n")
	summary.WriteString(fmt.Sprintf("- **主要技术栈**: %s\n", strings.Join(profile.TechnicalProfile.PrimaryLanguages, ", ")))
	summary.WriteString(fmt.Sprintf("- **架构关注点**: %s\n", profile.TechnicalProfile.ArchitecturalFocus))
	summary.WriteString(fmt.Sprintf("- **专业化程度**: %.0f%%\n", profile.CollaborationStyle.SpecializationLevel*100))
	summary.WriteString(fmt.Sprintf("- **导师等级**: %s\n", profile.CollaborationStyle.MentorshipLevel))
	summary.WriteString("\n")
	
	// Work Style Insights
	summary.WriteString("### 💡 工作风格洞察\n")
	summary.WriteString(profile.generateWorkStyleInsights())
	
	return summary.String()
}

// generateWorkStyleInsights generates personalized insights
func (profile *DeveloperProfile) generateWorkStyleInsights() string {
	var insights []string
	
	// Analyze work style type
	switch profile.PersonalityTraits.WorkStyleType {
	case "burst":
		insights = append(insights, "💥 **爆发型开发者**: 喜欢集中时间完成大量工作，适合处理复杂项目和深度开发任务")
	case "steady":
		insights = append(insights, "🎯 **稳定型开发者**: 工作节奏稳定，适合长期项目维护和持续改进工作")
	case "balanced":
		insights = append(insights, "⚖️ **平衡型开发者**: 工作方式灵活，能够适应不同类型的项目需求")
	}
	
	// Analyze collaboration style
	if profile.CollaborationStyle.CrossTeamWork > 0.6 {
		insights = append(insights, "🤝 **团队协作者**: 经常跨团队工作，善于协作和知识分享")
	} else if profile.CollaborationStyle.FilesOwnershipRatio > 0.7 {
		insights = append(insights, "👨‍💻 **独立贡献者**: 专注于特定领域，深入掌握相关代码")
	}
	
	// Analyze quality focus
	if profile.QualityIndicators.CommitMessageQuality > 80 {
		insights = append(insights, "📝 **文档专家**: 注重代码文档和提交信息质量，有助于团队知识传承")
	}
	
	// Analyze risk tolerance
	switch profile.PersonalityTraits.RiskTolerance {
	case "aggressive":
		insights = append(insights, "🚀 **创新先锋**: 乐于尝试新技术和方法，推动团队技术发展")
	case "conservative":
		insights = append(insights, "🛡️ **稳健派**: 注重代码稳定性和可靠性，确保项目质量")
	}
	
	// Generate recommendations
	insights = append(insights, "\n**💡 发展建议:**")
	
	if profile.CodingPatterns.TestingEngagement < 0.3 {
		insights = append(insights, "- 建议增加测试相关工作，提升代码质量保障")
	}
	
	if profile.CodingPatterns.DocumentationRatio < 0.2 {
		insights = append(insights, "- 可以考虑增加文档贡献，提升团队知识共享")
	}
	
	if profile.CollaborationStyle.CrossTeamWork < 0.3 {
		insights = append(insights, "- 建议多参与跨团队协作，扩展技术视野")
	}
	
	return strings.Join(insights, "\n- ") + "\n"
}
