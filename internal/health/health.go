package health

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"git-log-analyzer/internal/git"
)

// CodeHealthAnalyzer performs code health analysis
type CodeHealthAnalyzer struct {
	commits []git.GitCommit
}

// NewCodeHealthAnalyzer creates a new code health analyzer
func NewCodeHealthAnalyzer(commits []git.GitCommit) *CodeHealthAnalyzer {
	return &CodeHealthAnalyzer{
		commits: commits,
	}
}

// CodeHealthMetrics contains all code health analysis results
type CodeHealthMetrics struct {
	TechnicalDebtHotspots   []TechnicalDebtHotspot   `json:"technicalDebtHotspots"`
	StabilityIndicators     []StabilityIndicator     `json:"stabilityIndicators"`
	RefactoringSignals      []RefactoringSignal      `json:"refactoringSignals"`
	CodeConcentrationIssues []CodeConcentrationIssue `json:"codeConcentrationIssues"`
	HealthScore             float64                  `json:"healthScore"`
	HealthSummary           string                   `json:"healthSummary"`
}

// TechnicalDebtHotspot represents a file with potential technical debt
type TechnicalDebtHotspot struct {
	FilePath         string    `json:"filePath"`
	ModificationFreq int       `json:"modificationFreq"`
	UniqueAuthors    int       `json:"uniqueAuthors"`
	TotalChanges     int       `json:"totalChanges"`
	RiskScore        float64   `json:"riskScore"`
	LastModified     time.Time `json:"lastModified"`
	Reason           string    `json:"reason"`
}

// StabilityIndicator represents file stability metrics
type StabilityIndicator struct {
	FilePath        string  `json:"filePath"`
	ShakeIndex      float64 `json:"shakeIndex"`      // 震荡指数
	TimeSpread      float64 `json:"timeSpread"`      // 时间分布
	ModificationGap float64 `json:"modificationGap"` // 修改间隔方差
	StabilityLevel  string  `json:"stabilityLevel"`
}

// RefactoringSignal represents potential refactoring needs
type RefactoringSignal struct {
	FilePath          string    `json:"filePath"`
	IntensiveModDays  int       `json:"intensiveModDays"`  // 密集修改天数
	ShortTermChanges  int       `json:"shortTermChanges"`  // 短期内修改次数
	RefactoringSignal string    `json:"refactoringSignal"` // 重构信号强度
	TimeWindow        string    `json:"timeWindow"`
	FirstChange       time.Time `json:"firstChange"`
	LastChange        time.Time `json:"lastChange"`
}

// CodeConcentrationIssue represents "God File" issues
type CodeConcentrationIssue struct {
	FilePath       string  `json:"filePath"`
	TotalChanges   int     `json:"totalChanges"`
	AuthorCount    int     `json:"authorCount"`
	ChangeRatio    float64 `json:"changeRatio"`    // 占总变更的比例
	ConcentrationLevel string `json:"concentrationLevel"`
	ImpactLevel    string  `json:"impactLevel"`
}

// AnalyzeCodeHealth performs comprehensive code health analysis
func (cha *CodeHealthAnalyzer) AnalyzeCodeHealth() *CodeHealthMetrics {
	// 分析技术债务热点
	techDebtHotspots := cha.analyzeTechnicalDebtHotspots()
	
	// 分析代码稳定性指标
	stabilityIndicators := cha.analyzeStabilityIndicators()
	
	// 分析重构信号
	refactoringSignals := cha.analyzeRefactoringSignals()
	
	// 分析代码集中度问题
	concentrationIssues := cha.analyzeCodeConcentration()
	
	// 计算总体健康分数
	healthScore := cha.calculateHealthScore(techDebtHotspots, stabilityIndicators, refactoringSignals, concentrationIssues)
	
	// 生成健康总结
	healthSummary := cha.generateHealthSummary(healthScore, len(techDebtHotspots), len(refactoringSignals), len(concentrationIssues))

	return &CodeHealthMetrics{
		TechnicalDebtHotspots:   techDebtHotspots,
		StabilityIndicators:     stabilityIndicators,
		RefactoringSignals:      refactoringSignals,
		CodeConcentrationIssues: concentrationIssues,
		HealthScore:             healthScore,
		HealthSummary:           healthSummary,
	}
}

// analyzeTechnicalDebtHotspots identifies files with potential technical debt
func (cha *CodeHealthAnalyzer) analyzeTechnicalDebtHotspots() []TechnicalDebtHotspot {
	fileStats := make(map[string]*fileStatistic)
	
	// 统计每个文件的修改信息
	for _, commit := range cha.commits {
		for _, file := range commit.Files {
			if fileStats[file] == nil {
				fileStats[file] = &fileStatistic{
					FilePath:      file,
					Authors:       make(map[string]bool),
					Changes:       0,
					FirstModified: commit.Date,
					LastModified:  commit.Date,
				}
			}
			
			stat := fileStats[file]
			stat.Authors[commit.Author] = true
			stat.Changes++
			
			if commit.Date.Before(stat.FirstModified) {
				stat.FirstModified = commit.Date
			}
			if commit.Date.After(stat.LastModified) {
				stat.LastModified = commit.Date
			}
		}
	}
	
	var hotspots []TechnicalDebtHotspot
	
	// 计算技术债务风险分数
	for _, stat := range fileStats {
		if stat.Changes < 3 { // 忽略修改次数太少的文件
			continue
		}
		
		riskScore := cha.calculateTechDebtRisk(stat)
		reason := cha.getTechDebtReason(stat, riskScore)
		
		if riskScore > 0.3 { // 只包含风险分数较高的文件
			hotspots = append(hotspots, TechnicalDebtHotspot{
				FilePath:         stat.FilePath,
				ModificationFreq: stat.Changes,
				UniqueAuthors:    len(stat.Authors),
				TotalChanges:     stat.Changes,
				RiskScore:        riskScore,
				LastModified:     stat.LastModified,
				Reason:           reason,
			})
		}
	}
	
	// 按风险分数排序
	sort.Slice(hotspots, func(i, j int) bool {
		return hotspots[i].RiskScore > hotspots[j].RiskScore
	})
	
	// 限制返回数量
	if len(hotspots) > 10 {
		hotspots = hotspots[:10]
	}
	
	return hotspots
}

// analyzeStabilityIndicators calculates file stability metrics
func (cha *CodeHealthAnalyzer) analyzeStabilityIndicators() []StabilityIndicator {
	fileChanges := make(map[string][]time.Time)
	
	// 收集每个文件的修改时间
	for _, commit := range cha.commits {
		for _, file := range commit.Files {
			fileChanges[file] = append(fileChanges[file], commit.Date)
		}
	}
	
	var indicators []StabilityIndicator
	
	for filePath, changes := range fileChanges {
		if len(changes) < 2 {
			continue
		}
		
		// 排序时间
		sort.Slice(changes, func(i, j int) bool {
			return changes[i].Before(changes[j])
		})
		
		// 计算震荡指数
		shakeIndex := cha.calculateShakeIndex(changes)
		
		// 计算时间分布
		timeSpread := cha.calculateTimeSpread(changes)
		
		// 计算修改间隔方差
		modGap := cha.calculateModificationGap(changes)
		
		// 确定稳定性等级
		stabilityLevel := cha.getStabilityLevel(shakeIndex, timeSpread, modGap)
		
		indicators = append(indicators, StabilityIndicator{
			FilePath:        filePath,
			ShakeIndex:      shakeIndex,
			TimeSpread:      timeSpread,
			ModificationGap: modGap,
			StabilityLevel:  stabilityLevel,
		})
	}
	
	// 按震荡指数排序
	sort.Slice(indicators, func(i, j int) bool {
		return indicators[i].ShakeIndex > indicators[j].ShakeIndex
	})
	
	// 限制返回数量
	if len(indicators) > 15 {
		indicators = indicators[:15]
	}
	
	return indicators
}

// analyzeRefactoringSignals identifies files that may need refactoring
func (cha *CodeHealthAnalyzer) analyzeRefactoringSignals() []RefactoringSignal {
	var signals []RefactoringSignal
	
	// 分析短期内密集修改的文件
	recentWindow := 7 * 24 * time.Hour // 7天窗口
	
	fileRecentChanges := make(map[string][]time.Time)
	
	now := time.Now()
	cutoff := now.Add(-recentWindow)
	
	// 收集最近的修改
	for _, commit := range cha.commits {
		if commit.Date.After(cutoff) {
			for _, file := range commit.Files {
				fileRecentChanges[file] = append(fileRecentChanges[file], commit.Date)
			}
		}
	}
	
	for filePath, changes := range fileRecentChanges {
		if len(changes) >= 3 { // 7天内修改3次以上
			sort.Slice(changes, func(i, j int) bool {
				return changes[i].Before(changes[j])
			})
			
			// 计算密集修改天数
			intensiveDays := cha.calculateIntensiveModificationDays(changes)
			
			signalStrength := cha.getRefactoringSignalStrength(len(changes), intensiveDays)
			
			signals = append(signals, RefactoringSignal{
				FilePath:          filePath,
				IntensiveModDays:  intensiveDays,
				ShortTermChanges:  len(changes),
				RefactoringSignal: signalStrength,
				TimeWindow:        "7 days",
				FirstChange:       changes[0],
				LastChange:        changes[len(changes)-1],
			})
		}
	}
	
	// 按修改频率排序
	sort.Slice(signals, func(i, j int) bool {
		return signals[i].ShortTermChanges > signals[j].ShortTermChanges
	})
	
	return signals
}

// analyzeCodeConcentration identifies "God Files" with excessive changes
func (cha *CodeHealthAnalyzer) analyzeCodeConcentration() []CodeConcentrationIssue {
	fileStats := make(map[string]*fileStatistic)
	totalChanges := 0
	
	// 统计文件修改信息
	for _, commit := range cha.commits {
		for _, file := range commit.Files {
			if fileStats[file] == nil {
				fileStats[file] = &fileStatistic{
					FilePath: file,
					Authors:  make(map[string]bool),
					Changes:  0,
				}
			}
			
			fileStats[file].Authors[commit.Author] = true
			fileStats[file].Changes++
			totalChanges++
		}
	}
	
	var issues []CodeConcentrationIssue
	
	for _, stat := range fileStats {
		changeRatio := float64(stat.Changes) / float64(totalChanges)
		
		// 只关注占总变更10%以上的文件
		if changeRatio > 0.1 || stat.Changes > 20 {
			concentrationLevel := cha.getConcentrationLevel(changeRatio, stat.Changes)
			impactLevel := cha.getImpactLevel(stat.Changes, len(stat.Authors))
			
			issues = append(issues, CodeConcentrationIssue{
				FilePath:           stat.FilePath,
				TotalChanges:       stat.Changes,
				AuthorCount:        len(stat.Authors),
				ChangeRatio:        changeRatio,
				ConcentrationLevel: concentrationLevel,
				ImpactLevel:        impactLevel,
			})
		}
	}
	
	// 按变更比例排序
	sort.Slice(issues, func(i, j int) bool {
		return issues[i].ChangeRatio > issues[j].ChangeRatio
	})
	
	return issues
}

// Helper types and methods
type fileStatistic struct {
	FilePath      string
	Authors       map[string]bool
	Changes       int
	FirstModified time.Time
	LastModified  time.Time
}

// calculateTechDebtRisk calculates technical debt risk score
func (cha *CodeHealthAnalyzer) calculateTechDebtRisk(stat *fileStatistic) float64 {
	// 考虑多个因素
	changeFreqScore := math.Min(float64(stat.Changes)/20.0, 1.0)     // 修改频率
	authorDiversityScore := math.Min(float64(len(stat.Authors))/5.0, 1.0) // 作者多样性
	
	// 综合风险分数
	riskScore := (changeFreqScore*0.6 + authorDiversityScore*0.4)
	
	return math.Min(riskScore, 1.0)
}

// getTechDebtReason provides reason for technical debt classification
func (cha *CodeHealthAnalyzer) getTechDebtReason(stat *fileStatistic, riskScore float64) string {
	reasons := []string{}
	
	if stat.Changes > 15 {
		reasons = append(reasons, "频繁修改")
	}
	if len(stat.Authors) > 3 {
		reasons = append(reasons, "多人修改")
	}
	if riskScore > 0.7 {
		reasons = append(reasons, "高风险")
	}
	
	if len(reasons) == 0 {
		return "潜在技术债务"
	}
	
	return strings.Join(reasons, ", ")
}

// calculateShakeIndex calculates how frequently a file changes
func (cha *CodeHealthAnalyzer) calculateShakeIndex(changes []time.Time) float64 {
	if len(changes) < 2 {
		return 0
	}
	
	// 计算变更频率的变异性
	totalDuration := changes[len(changes)-1].Sub(changes[0]).Hours()
	if totalDuration == 0 {
		return float64(len(changes))
	}
	
	changeRate := float64(len(changes)) / (totalDuration / 24) // changes per day
	return math.Min(changeRate, 10.0) // 限制最大值
}

// calculateTimeSpread calculates the time distribution of changes
func (cha *CodeHealthAnalyzer) calculateTimeSpread(changes []time.Time) float64 {
	if len(changes) < 2 {
		return 0
	}
	
	totalSpan := changes[len(changes)-1].Sub(changes[0]).Hours()
	return totalSpan / 24 // 返回天数
}

// calculateModificationGap calculates variance in modification intervals
func (cha *CodeHealthAnalyzer) calculateModificationGap(changes []time.Time) float64 {
	if len(changes) < 3 {
		return 0
	}
	
	var gaps []float64
	for i := 1; i < len(changes); i++ {
		gap := changes[i].Sub(changes[i-1]).Hours()
		gaps = append(gaps, gap)
	}
	
	// 计算方差
	mean := 0.0
	for _, gap := range gaps {
		mean += gap
	}
	mean /= float64(len(gaps))
	
	variance := 0.0
	for _, gap := range gaps {
		variance += (gap - mean) * (gap - mean)
	}
	variance /= float64(len(gaps))
	
	return math.Sqrt(variance) / 24 // 转换为天数
}

// getStabilityLevel determines stability level based on metrics
func (cha *CodeHealthAnalyzer) getStabilityLevel(shakeIndex, timeSpread, modGap float64) string {
	if shakeIndex > 2.0 && modGap > 3.0 {
		return "极不稳定"
	} else if shakeIndex > 1.0 && modGap > 1.5 {
		return "不稳定"
	} else if shakeIndex > 0.5 {
		return "中等稳定"
	} else {
		return "稳定"
	}
}

// calculateIntensiveModificationDays calculates intensive modification days
func (cha *CodeHealthAnalyzer) calculateIntensiveModificationDays(changes []time.Time) int {
	dayMap := make(map[string]bool)
	
	for _, change := range changes {
		day := change.Format("2006-01-02")
		dayMap[day] = true
	}
	
	return len(dayMap)
}

// getRefactoringSignalStrength determines refactoring signal strength
func (cha *CodeHealthAnalyzer) getRefactoringSignalStrength(changes, days int) string {
	ratio := float64(changes) / float64(days)
	
	if ratio >= 3.0 {
		return "强烈"
	} else if ratio >= 2.0 {
		return "中等"
	} else {
		return "轻微"
	}
}

// getConcentrationLevel determines code concentration level
func (cha *CodeHealthAnalyzer) getConcentrationLevel(ratio float64, changes int) string {
	if ratio > 0.3 || changes > 50 {
		return "严重集中"
	} else if ratio > 0.2 || changes > 30 {
		return "高度集中"
	} else if ratio > 0.1 || changes > 20 {
		return "中度集中"
	} else {
		return "轻度集中"
	}
}

// getImpactLevel determines impact level based on changes and authors
func (cha *CodeHealthAnalyzer) getImpactLevel(changes, authors int) string {
	if changes > 30 && authors > 3 {
		return "高影响"
	} else if changes > 20 || authors > 2 {
		return "中影响"
	} else {
		return "低影响"
	}
}

// calculateHealthScore calculates overall code health score
func (cha *CodeHealthAnalyzer) calculateHealthScore(hotspots []TechnicalDebtHotspot, indicators []StabilityIndicator, signals []RefactoringSignal, issues []CodeConcentrationIssue) float64 {
	// 基础分数
	baseScore := 1.0
	
	// 根据各种问题降低分数
	baseScore -= float64(len(hotspots)) * 0.05       // 技术债务热点
	baseScore -= float64(len(signals)) * 0.08        // 重构信号
	baseScore -= float64(len(issues)) * 0.1          // 代码集中度问题
	
	// 根据稳定性指标调整
	unstableCount := 0
	for _, indicator := range indicators {
		if indicator.StabilityLevel == "极不稳定" || indicator.StabilityLevel == "不稳定" {
			unstableCount++
		}
	}
	baseScore -= float64(unstableCount) * 0.03
	
	// 确保分数在0-1之间
	if baseScore < 0 {
		baseScore = 0
	}
	
	return baseScore
}

// generateHealthSummary generates a health summary string
func (cha *CodeHealthAnalyzer) generateHealthSummary(score float64, hotspots, signals, issues int) string {
	var level string
	var description string
	
	if score >= 0.8 {
		level = "健康"
		description = "代码质量良好，维护性强"
	} else if score >= 0.6 {
		level = "中等"
		description = "存在一些质量问题，建议关注"
	} else if score >= 0.4 {
		level = "较差"
		description = "代码质量问题较多，需要改进"
	} else {
		level = "差"
		description = "代码质量堪忧，急需重构"
	}
	
	return fmt.Sprintf("代码健康等级：%s (%.2f分) - %s。发现%d个技术债务热点，%d个重构信号，%d个代码集中度问题。", 
		level, score*100, description, hotspots, signals, issues)
}
