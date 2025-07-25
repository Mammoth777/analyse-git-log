package progress

import (
	"fmt"
	"time"

	"github.com/schollz/progressbar/v3"
)

// ProgressTrackerV2 uses professional progress bar library
type ProgressTrackerV2 struct {
	totalSteps    int
	currentStep   int
	startTime     time.Time
	stepStartTime time.Time
	stepName      string
	verbose       bool
	progressBar   *progressbar.ProgressBar
}

// NewProgressTrackerV2 creates a new progress tracker using professional library
func NewProgressTrackerV2(totalSteps int, verbose bool) *ProgressTrackerV2 {
	return &ProgressTrackerV2{
		totalSteps:  totalSteps,
		currentStep: 0,
		startTime:   time.Now(),
		verbose:     verbose,
	}
}

// StartStep begins a new step in the process
func (pt *ProgressTrackerV2) StartStep(stepName string) {
	pt.currentStep++
	pt.stepName = stepName
	pt.stepStartTime = time.Now()
	
	fmt.Printf("\n🚀 [%d/%d] %s\n", pt.currentStep, pt.totalSteps, stepName)
	
	if pt.verbose {
		elapsed := time.Since(pt.startTime)
		fmt.Printf("   ⏱️  总耗时: %v\n", elapsed.Round(time.Millisecond))
	}
	
	// 创建新的进度条
	pt.progressBar = progressbar.NewOptions(100,
		progressbar.OptionSetDescription("   进度"),
		progressbar.OptionSetWidth(30),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "█",
			SaucerHead:    "▶",
			SaucerPadding: "░",
			BarStart:      "[",
			BarEnd:        "]",
		}),
		progressbar.OptionSetPredictTime(true),
		progressbar.OptionSpinnerType(14), // 使用spinner动画
	)
	
	// 设置初始进度
	baseProgress := float64(pt.currentStep-1) / float64(pt.totalSteps) * 100
	pt.progressBar.Set(int(baseProgress))
}

// UpdateStepProgress updates the current step's progress
func (pt *ProgressTrackerV2) UpdateStepProgress(message string) {
	if pt.verbose {
		stepElapsed := time.Since(pt.stepStartTime)
		fmt.Printf("\n   📋 %s (耗时: %v)\n", message, stepElapsed.Round(time.Millisecond))
	} else {
		fmt.Printf("\n   📋 %s\n", message)
	}
	
	// 更新进度条
	if pt.progressBar != nil {
		// 在当前步骤内增加一些进度
		baseProgress := float64(pt.currentStep-1) / float64(pt.totalSteps) * 100
		stepProgress := time.Since(pt.stepStartTime).Seconds() * 2 // 每秒增加2%
		if stepProgress > 10 {
			stepProgress = 10
		}
		currentProgress := baseProgress + (stepProgress / float64(pt.totalSteps))
		if currentProgress > 100 {
			currentProgress = 100
		}
		pt.progressBar.Set(int(currentProgress))
	}
}

// CompleteStep marks the current step as completed
func (pt *ProgressTrackerV2) CompleteStep(result string) {
	stepElapsed := time.Since(pt.stepStartTime)
	percentage := float64(pt.currentStep) / float64(pt.totalSteps) * 100
	
	// 完成当前步骤的进度条
	if pt.progressBar != nil {
		pt.progressBar.Set(int(percentage))
		pt.progressBar.Finish()
	}
	
	fmt.Printf("\n   ✅ %s (完成，耗时: %v)\n", result, stepElapsed.Round(time.Millisecond))
}

// CompleteStepWithWarning marks the current step as completed with warning
func (pt *ProgressTrackerV2) CompleteStepWithWarning(result string, warning string) {
	stepElapsed := time.Since(pt.stepStartTime)
	percentage := float64(pt.currentStep) / float64(pt.totalSteps) * 100
	
	// 完成当前步骤的进度条
	if pt.progressBar != nil {
		pt.progressBar.Set(int(percentage))
		pt.progressBar.Finish()
	}
	
	fmt.Printf("\n   ⚠️  %s (警告: %s，耗时: %v)\n", result, warning, stepElapsed.Round(time.Millisecond))
}

// FailStep marks the current step as failed
func (pt *ProgressTrackerV2) FailStep(errorMsg string) {
	stepElapsed := time.Since(pt.stepStartTime)
	
	// 停止进度条
	if pt.progressBar != nil {
		pt.progressBar.Finish()
	}
	
	fmt.Printf("\n   ❌ 失败: %s (耗时: %v)\n", errorMsg, stepElapsed.Round(time.Millisecond))
}

// Complete marks the entire process as completed
func (pt *ProgressTrackerV2) Complete() {
	totalElapsed := time.Since(pt.startTime)
	
	// 创建最终完成进度条
	finalBar := progressbar.NewOptions(100,
		progressbar.OptionSetDescription("🎉 分析完成"),
		progressbar.OptionSetWidth(30),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "█",
			SaucerHead:    "█",
			SaucerPadding: "░",
			BarStart:      "[",
			BarEnd:        "]",
		}),
		progressbar.OptionShowCount(),
	)
	finalBar.Set(100)
	finalBar.Finish()
	
	fmt.Printf("\n   ⏱️  总耗时: %v\n", totalElapsed.Round(time.Millisecond))
	fmt.Printf("   📊 共完成 %d 个分析步骤\n", pt.totalSteps)
}

// ShowSummary displays a summary of the analysis
func (pt *ProgressTrackerV2) ShowSummary(stats interface{}) {
	fmt.Printf("\n📈 分析摘要:\n")
	fmt.Printf("   ⏰ 开始时间: %s\n", pt.startTime.Format("2006-01-02 15:04:05"))
	fmt.Printf("   ⏰ 完成时间: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Printf("   ⌛ 总耗时: %v\n", time.Since(pt.startTime).Round(time.Millisecond))
}

// CreateSubTracker creates a sub-tracker for detailed operations
func (pt *ProgressTrackerV2) CreateSubTracker(stepName string, subSteps int) *SubProgressTrackerV2 {
	return &SubProgressTrackerV2{
		parent:      pt,
		stepName:    stepName,
		totalSubs:   subSteps,
		currentSub:  0,
		startTime:   time.Now(),
		progressBar: progressbar.NewOptions(subSteps,
			progressbar.OptionSetDescription(fmt.Sprintf("      └─ %s", stepName)),
			progressbar.OptionSetWidth(20),
			progressbar.OptionShowCount(),
			progressbar.OptionSetTheme(progressbar.Theme{
				Saucer:        "▓",
				SaucerHead:    "▶",
				SaucerPadding: "░",
				BarStart:      "[",
				BarEnd:        "]",
			}),
		),
	}
}

// SubProgressTrackerV2 tracks progress within a main step
type SubProgressTrackerV2 struct {
	parent      *ProgressTrackerV2
	stepName    string
	totalSubs   int
	currentSub  int
	startTime   time.Time
	progressBar *progressbar.ProgressBar
}

// UpdateSub updates sub-step progress
func (spt *SubProgressTrackerV2) UpdateSub(subStepName string) {
	spt.currentSub++
	spt.progressBar.Add(1)
	
	if spt.parent.verbose {
		elapsed := time.Since(spt.startTime)
		percentage := float64(spt.currentSub) / float64(spt.totalSubs) * 100
		fmt.Printf("         %s (%.1f%%, 耗时: %v)\n", 
			subStepName, percentage, elapsed.Round(time.Millisecond))
	}
}

// CompleteSub completes the sub-tracker
func (spt *SubProgressTrackerV2) CompleteSub(message string) {
	elapsed := time.Since(spt.startTime)
	spt.progressBar.Finish()
	fmt.Printf("      ✅ %s (完成，耗时: %v)\n", message, elapsed.Round(time.Millisecond))
}
