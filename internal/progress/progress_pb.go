package progress

import (
	"fmt"
	"time"

	"github.com/cheggaaa/pb/v3"
)

// ProgressTrackerPB uses cheggaaa/pb library
type ProgressTrackerPB struct {
	totalSteps    int
	currentStep   int
	startTime     time.Time
	stepStartTime time.Time
	stepName      string
	verbose       bool
	progressBar   *pb.ProgressBar
}

// NewProgressTrackerPB creates a new progress tracker using pb library
func NewProgressTrackerPB(totalSteps int, verbose bool) *ProgressTrackerPB {
	return &ProgressTrackerPB{
		totalSteps:  totalSteps,
		currentStep: 0,
		startTime:   time.Now(),
		verbose:     verbose,
	}
}

// StartStep begins a new step in the process
func (pt *ProgressTrackerPB) StartStep(stepName string) {
	pt.currentStep++
	pt.stepName = stepName
	pt.stepStartTime = time.Now()
	
	fmt.Printf("\n🚀 [%d/%d] %s\n", pt.currentStep, pt.totalSteps, stepName)
	
	// 创建PB进度条
	pt.progressBar = pb.Full.Start(100)
	pt.progressBar.SetTemplateString(`   进度 {{.Bar}} {{.Percent}}% {{.Speed}} {{.TimeElapsed}}/{{.TimeRemaining}}`)
	
	// 设置初始进度
	baseProgress := (pt.currentStep - 1) * 100 / pt.totalSteps
	pt.progressBar.SetCurrent(int64(baseProgress))
}

// UpdateStepProgress updates the current step's progress
func (pt *ProgressTrackerPB) UpdateStepProgress(message string) {
	if pt.verbose {
		stepElapsed := time.Since(pt.stepStartTime)
		fmt.Printf("\n   📋 %s (耗时: %v)\n", message, stepElapsed.Round(time.Millisecond))
	} else {
		fmt.Printf("\n   📋 %s\n", message)
	}
	
	// 更新进度
	if pt.progressBar != nil {
		baseProgress := (pt.currentStep - 1) * 100 / pt.totalSteps
		stepProgress := int(time.Since(pt.stepStartTime).Seconds() * 2) // 每秒增加2%
		if stepProgress > 10 {
			stepProgress = 10
		}
		currentProgress := baseProgress + stepProgress/pt.totalSteps
		if currentProgress > 100 {
			currentProgress = 100
		}
		pt.progressBar.SetCurrent(int64(currentProgress))
	}
}

// CompleteStep marks the current step as completed
func (pt *ProgressTrackerPB) CompleteStep(result string) {
	stepElapsed := time.Since(pt.stepStartTime)
	
	// 完成进度条
	if pt.progressBar != nil {
		percentage := pt.currentStep * 100 / pt.totalSteps
		pt.progressBar.SetCurrent(int64(percentage))
		pt.progressBar.Finish()
	}
	
	fmt.Printf("\n   ✅ %s (完成，耗时: %v)\n", result, stepElapsed.Round(time.Millisecond))
}

// Complete marks the entire process as completed
func (pt *ProgressTrackerPB) Complete() {
	totalElapsed := time.Since(pt.startTime)
	
	// 创建最终进度条
	finalBar := pb.Full.Start(100)
	finalBar.SetTemplateString(`🎉 分析完成 {{.Bar}} {{.Percent}}%`)
	finalBar.SetCurrent(100)
	finalBar.Finish()
	
	fmt.Printf("\n   ⏱️  总耗时: %v\n", totalElapsed.Round(time.Millisecond))
	fmt.Printf("   📊 共完成 %d 个分析步骤\n", pt.totalSteps)
}
