package progress

import (
	"fmt"
	"strings"
	"time"
)

// ProgressTracker tracks and displays progress for different stages
type ProgressTracker struct {
	totalSteps          int
	currentStep         int
	startTime           time.Time
	stepStartTime       time.Time
	stepName            string
	verbose             bool
	isProgressBarActive bool     // 标记进度条是否激活
	progressBarLine     string   // 存储当前进度条内容
	hasProgressBar      bool     // 标记是否已经显示了进度条
}

// NewProgressTracker creates a new progress tracker
func NewProgressTracker(totalSteps int, verbose bool) *ProgressTracker {
	return &ProgressTracker{
		totalSteps:          totalSteps,
		currentStep:         0,
		startTime:           time.Now(),
		verbose:             verbose,
		isProgressBarActive: false,
		hasProgressBar:      false,
	}
}

// clearProgressBar clears the current progress bar line
func (pt *ProgressTracker) clearProgressBar() {
	if pt.hasProgressBar {
		// 移动光标到进度条行并清除
		fmt.Print("\r" + strings.Repeat(" ", len(pt.progressBarLine)) + "\r")
	}
}

// updateProgressBar updates the progress bar at the bottom
func (pt *ProgressTracker) updateProgressBar(newProgressLine string) {
	pt.clearProgressBar()
	pt.progressBarLine = newProgressLine
	fmt.Print(newProgressLine)
	pt.hasProgressBar = true
}

// printMessage prints a message above the progress bar
func (pt *ProgressTracker) printMessage(message string) {
	pt.clearProgressBar()
	fmt.Println(message)
	// 重新显示进度条
	if pt.hasProgressBar && pt.progressBarLine != "" {
		fmt.Print(pt.progressBarLine)
	}
}

// StartStep begins a new step in the process
func (pt *ProgressTracker) StartStep(stepName string) {
	pt.currentStep++
	pt.stepName = stepName
	pt.stepStartTime = time.Now()
	
	// 打印步骤信息（会自动在进度条上方显示）
	message := fmt.Sprintf("\n🚀 [%d/%d] %s", pt.currentStep, pt.totalSteps, stepName)
	if pt.verbose {
		elapsed := time.Since(pt.startTime)
		message += fmt.Sprintf("\n   ⏱️  总耗时: %v", elapsed.Round(time.Millisecond))
	}
	pt.printMessage(message)
	
	// 开始显示动态进度条
	pt.startDynamicProgress()
}

// startDynamicProgress starts the dynamic progress bar for current step
func (pt *ProgressTracker) startDynamicProgress() {
	pt.isProgressBarActive = true
	go func() {
		ticker := time.NewTicker(200 * time.Millisecond) // 每200ms更新一次
		defer ticker.Stop()
		
		animFrame := 0
		spinChars := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		
		for pt.isProgressBarActive {
			select {
			case <-ticker.C:
				if pt.isProgressBarActive {
					// 计算基础进度
					baseProgress := float64(pt.currentStep-1) / float64(pt.totalSteps) * 100
					
					// 在当前步骤内添加模拟进度（0-10%的范围）
					stepProgress := time.Since(pt.stepStartTime).Seconds() * 2 // 每秒增加2%
					if stepProgress > 10 {
						stepProgress = 10
					}
					
					currentProgress := baseProgress + (stepProgress / float64(pt.totalSteps))
					if currentProgress > 100 {
						currentProgress = 100
					}
					
					// 旋转动画
					spinner := spinChars[animFrame%len(spinChars)]
					animFrame++
					
					// 创建进度条
					progressBar := pt.createProgressBar(currentProgress, 30)
					
					progressText := fmt.Sprintf("   进度: %s %.1f%% %s", 
						progressBar, currentProgress, spinner)
					
					pt.updateProgressBar(progressText)
				}
			default:
				if !pt.isProgressBarActive {
					return
				}
				time.Sleep(10 * time.Millisecond)
			}
		}
	}()
}

// UpdateStepProgress updates the current step's progress
func (pt *ProgressTracker) UpdateStepProgress(message string) {
	// 停止动态进度条，显示消息，然后重新开始
	pt.isProgressBarActive = false
	time.Sleep(50 * time.Millisecond) // 等待动画停止
	
	var msg string
	if pt.verbose {
		stepElapsed := time.Since(pt.stepStartTime)
		msg = fmt.Sprintf("   📋 %s (耗时: %v)", message, stepElapsed.Round(time.Millisecond))
	} else {
		msg = fmt.Sprintf("   📋 %s", message)
	}
	
	pt.printMessage(msg)
	
	// 重新开始动态进度条
	pt.startDynamicProgress()
}

// CompleteStep marks the current step as completed
func (pt *ProgressTracker) CompleteStep(result string) {
	// 停止动态进度条
	pt.isProgressBarActive = false
	time.Sleep(50 * time.Millisecond) // 等待动画停止
	
	stepElapsed := time.Since(pt.stepStartTime)
	percentage := float64(pt.currentStep) / float64(pt.totalSteps) * 100
	
	// 显示完成信息
	msg := fmt.Sprintf("   ✅ %s (完成，耗时: %v)", result, stepElapsed.Round(time.Millisecond))
	pt.printMessage(msg)
	
	// 只在非最后一步时显示进度条，最后一步在Complete()中统一处理
	if pt.currentStep < pt.totalSteps {
		progressBar := pt.createProgressBar(percentage, 30)
		finalProgressText := fmt.Sprintf("   进度: %s %.1f%%", progressBar, percentage)
		pt.updateProgressBar(finalProgressText)
	}
}

// CompleteStepWithWarning marks the current step as completed with warning
func (pt *ProgressTracker) CompleteStepWithWarning(result string, warning string) {
	// 停止动态进度条
	pt.isProgressBarActive = false
	time.Sleep(50 * time.Millisecond) // 等待动画停止
	
	stepElapsed := time.Since(pt.stepStartTime)
	percentage := float64(pt.currentStep) / float64(pt.totalSteps) * 100
	
	// 显示警告信息
	msg := fmt.Sprintf("   ⚠️  %s (警告: %s，耗时: %v)", result, warning, stepElapsed.Round(time.Millisecond))
	pt.printMessage(msg)
	
	// 只在非最后一步时显示进度条
	if pt.currentStep < pt.totalSteps {
		progressBar := pt.createProgressBar(percentage, 30)
		finalProgressText := fmt.Sprintf("   进度: %s %.1f%%", progressBar, percentage)
		pt.updateProgressBar(finalProgressText)
	}
}

// FailStep marks the current step as failed
func (pt *ProgressTracker) FailStep(errorMsg string) {
	// 停止动态进度条
	pt.isProgressBarActive = false
	time.Sleep(50 * time.Millisecond) // 等待动画停止
	
	stepElapsed := time.Since(pt.stepStartTime)
	msg := fmt.Sprintf("   ❌ 失败: %s (耗时: %v)", errorMsg, stepElapsed.Round(time.Millisecond))
	pt.printMessage(msg)
}

// Complete marks the entire process as completed
func (pt *ProgressTracker) Complete() {
	// 停止任何活跃的进度条
	pt.isProgressBarActive = false
	time.Sleep(50 * time.Millisecond) // 等待动画停止
	
	totalElapsed := time.Since(pt.startTime)
	
	// 清除任何现有的进度条
	pt.clearProgressBar()
	
	// 显示完成信息
	msg := fmt.Sprintf("\n🎉 分析完成! (总耗时: %v)", totalElapsed.Round(time.Millisecond))
	fmt.Println(msg)
	
	msg = fmt.Sprintf("   📊 共完成 %d 个分析步骤", pt.totalSteps)
	fmt.Println(msg)
	
	// 显示唯一的最终进度条
	progressBar := pt.createProgressBar(100, 30)
	finalProgressText := fmt.Sprintf("🎉 最终进度 %s 100%%", progressBar)
	fmt.Println(finalProgressText)
}

// createProgressBar creates a visual progress bar with gradient effect
func (pt *ProgressTracker) createProgressBar(percentage float64, width int) string {
	filled := int(percentage / 100 * float64(width))
	if filled > width {
		filled = width
	}
	
	// 使用不同字符创建渐变效果
	var bar strings.Builder
	bar.WriteString("[")
	
	for i := 0; i < width; i++ {
		if i < filled {
			if i == filled-1 && filled < width {
				// 进度条的前端使用特殊字符
				bar.WriteString("▶")
			} else {
				bar.WriteString("█")
			}
		} else if i == filled && filled < width {
			// 当前位置使用渐变字符
			bar.WriteString("▓")
		} else {
			bar.WriteString("░")
		}
	}
	
	bar.WriteString("]")
	return bar.String()
}

// ShowSummary displays a summary of the analysis
func (pt *ProgressTracker) ShowSummary(stats interface{}) {
	fmt.Printf("\n📈 分析摘要:\n")
	fmt.Printf("   ⏰ 开始时间: %s\n", pt.startTime.Format("2006-01-02 15:04:05"))
	fmt.Printf("   ⏰ 完成时间: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Printf("   ⌛ 总耗时: %v\n", time.Since(pt.startTime).Round(time.Millisecond))
}

// EstimatedTimeRemaining calculates estimated time remaining
func (pt *ProgressTracker) EstimatedTimeRemaining() time.Duration {
	if pt.currentStep == 0 {
		return 0
	}
	
	elapsed := time.Since(pt.startTime)
	avgTimePerStep := elapsed / time.Duration(pt.currentStep)
	remainingSteps := pt.totalSteps - pt.currentStep
	
	return avgTimePerStep * time.Duration(remainingSteps)
}

// ShowDetailedProgress shows detailed progress information
func (pt *ProgressTracker) ShowDetailedProgress() {
	if !pt.verbose {
		return
	}
	
	elapsed := time.Since(pt.startTime)
	remaining := pt.EstimatedTimeRemaining()
	percentage := float64(pt.currentStep) / float64(pt.totalSteps) * 100
	
	fmt.Printf("   📊 详细进度信息:\n")
	fmt.Printf("      • 已完成步骤: %d/%d (%.1f%%)\n", pt.currentStep, pt.totalSteps, percentage)
	fmt.Printf("      • 已用时间: %v\n", elapsed.Round(time.Millisecond))
	fmt.Printf("      • 预计剩余: %v\n", remaining.Round(time.Millisecond))
	fmt.Printf("      • 平均每步: %v\n", (elapsed / time.Duration(pt.currentStep)).Round(time.Millisecond))
}

// CreateSubTracker creates a sub-tracker for detailed operations
func (pt *ProgressTracker) CreateSubTracker(stepName string, subSteps int) *SubProgressTracker {
	return &SubProgressTracker{
		parent:      pt,
		stepName:    stepName,
		totalSubs:   subSteps,
		currentSub:  0,
		startTime:   time.Now(),
	}
}

// SubProgressTracker tracks progress within a main step
type SubProgressTracker struct {
	parent     *ProgressTracker
	stepName   string
	totalSubs  int
	currentSub int
	startTime  time.Time
}

// UpdateSub updates sub-step progress
func (spt *SubProgressTracker) UpdateSub(subStepName string) {
	spt.currentSub++
	percentage := float64(spt.currentSub) / float64(spt.totalSubs) * 100
	
	var msg string
	if spt.parent.verbose {
		elapsed := time.Since(spt.startTime)
		msg = fmt.Sprintf("      └─ [%d/%d] %s (%.1f%%, 耗时: %v)", 
			spt.currentSub, spt.totalSubs, subStepName, percentage, elapsed.Round(time.Millisecond))
	} else {
		msg = fmt.Sprintf("      └─ %s", subStepName)
	}
	
	// 使用父进度跟踪器的printMessage方法，确保进度条正确更新
	spt.parent.printMessage(msg)
}

// CompleteSub completes the sub-tracker
func (spt *SubProgressTracker) CompleteSub(message string) {
	elapsed := time.Since(spt.startTime)
	msg := fmt.Sprintf("      ✅ %s (完成，耗时: %v)", message, elapsed.Round(time.Millisecond))
	
	// 使用父进度跟踪器的printMessage方法
	spt.parent.printMessage(msg)
}
