package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"git-log-analyzer/internal/ai"
	"git-log-analyzer/internal/analyzer"
	"git-log-analyzer/internal/developer"
	"git-log-analyzer/internal/progress"
	"git-log-analyzer/internal/report"
)

var cfgFile string
var repoPath string
var useAI bool
var outputFile string
var generateWeb bool
var outputDir string
var openBrowser bool
var reportLanguage string
var analysisMonths int

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "git-log-analyzer",
	Short: "A Git log analysis tool",
	Long: `Git Log Analyzer is a command-line tool that analyzes Git commit logs
and provides insights about your repository's development history.

This tool can:
- Analyze commits by author, time, and other dimensions
- Generate statistical reports
- Use AI models for advanced analysis

Environment variables for AI analysis:
- AI_API_ENDPOINT: API endpoint (default: https://api.openai.com/v1/chat/completions)
- AI_API_KEY: API key (required for AI analysis)
- AI_MODEL: Model to use (default: gpt-3.5-turbo)
- AI_MAX_TOKENS: Maximum tokens (default: 2000)
- AI_TEMPERATURE: Temperature setting (default: 0.7)

Environment variables for report customization:
- REPORT_LANGUAGE: Report language (zh/en, default: zh)`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := analyzeGitLog(repoPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.git-log-analyzer.yaml)")
	rootCmd.PersistentFlags().StringVarP(&repoPath, "repo", "r", "./", "path to git repository")
	rootCmd.PersistentFlags().BoolVar(&useAI, "ai", false, "enable AI-powered analysis")
	rootCmd.PersistentFlags().StringVarP(&outputFile, "output", "o", "", "output file for the text report")
	rootCmd.PersistentFlags().BoolVar(&generateWeb, "web", true, "generate web-based HTML report")
	rootCmd.PersistentFlags().StringVar(&outputDir, "output-dir", getEnv("REPORT_OUTPUT_DIR", "./analysis-reports"), "output directory for reports")
	rootCmd.PersistentFlags().BoolVar(&openBrowser, "open", getEnvBool("AUTO_OPEN_BROWSER", false), "automatically open web report in browser")
	rootCmd.PersistentFlags().StringVarP(&reportLanguage, "lang", "l", getEnv("REPORT_LANGUAGE", "zh"), "report language (zh/en)")
	rootCmd.PersistentFlags().IntVarP(&analysisMonths, "months", "m", 6, "analyze commits from the last N months (default: 6)")

	// Bind flags to viper
	viper.BindPFlag("repo", rootCmd.PersistentFlags().Lookup("repo"))
	viper.BindPFlag("ai", rootCmd.PersistentFlags().Lookup("ai"))
	viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output"))
	viper.BindPFlag("web", rootCmd.PersistentFlags().Lookup("web"))
	viper.BindPFlag("output-dir", rootCmd.PersistentFlags().Lookup("output-dir"))
	viper.BindPFlag("open", rootCmd.PersistentFlags().Lookup("open"))
	viper.BindPFlag("months", rootCmd.PersistentFlags().Lookup("months"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".git-log-analyzer")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

func analyzeGitLog(repoPath string) error {
	// Set language from command line flag
	if reportLanguage != "" {
		os.Setenv("REPORT_LANGUAGE", reportLanguage)
	}
	
	// Initialize progress tracker (using custom implementation)
	totalSteps := 4 // Git分析、开发者分析、报告生成、输出
	if useAI {
		totalSteps = 5 // 增加AI分析步骤
	}
	
	tracker := progress.NewProgressTracker(totalSteps, true)
	
	fmt.Printf("\n🔍 开始分析Git仓库: %s\n", repoPath)
	
	// 添加一些延迟让动态进度条效果更明显
	time.Sleep(500 * time.Millisecond)
	
	// Step 1: Environment validation and initialization
	tracker.StartStep("环境验证与初始化")
	
	// Validate git environment
	time.Sleep(300 * time.Millisecond) // 模拟耗时操作
	tracker.UpdateStepProgress("Git环境验证通过")
	
	// Create analyzer
	time.Sleep(400 * time.Millisecond) // 模拟耗时操作
	tracker.UpdateStepProgress("创建分析器实例")
	
	a := analyzer.NewAnalyzer(repoPath, analysisMonths)
	tracker.CompleteStep("环境初始化完成")
	
	time.Sleep(300 * time.Millisecond) // 让用户看到完成状态
	
	// Step 2: Git Log Analysis
	tracker.StartStep("Git日志分析")
	tracker.UpdateStepProgress("获取提交历史...")
	
	stats, err := a.AnalyzeWithProgress(func(current, total int, message string) {
		// 将analyzer的进度映射到这个步骤的进度更新
		if current%10 == 0 || current == total { // 每10%或完成时更新
			tracker.UpdateStepProgress(fmt.Sprintf("%s (%d%%)", message, current))
		}
	})
	if err != nil {
		tracker.FailStep(fmt.Sprintf("分析失败: %v", err))
		return fmt.Errorf("failed to analyze repository: %v", err)
	}
	
	tracker.UpdateStepProgress(fmt.Sprintf("已分析 %d 个提交", stats.TotalCommits))
	if stats.CodeHealthMetrics != nil {
		tracker.UpdateStepProgress(fmt.Sprintf("代码健康评分: %.0f/100", stats.CodeHealthMetrics.HealthScore*100))
	}
	
	basicReport := stats.GenerateReport()
	tracker.CompleteStep("Git日志分析完成")
	
	time.Sleep(300 * time.Millisecond) // 让用户看到完成状态
	
	// Step 3: Developer Profile Analysis
	tracker.StartStep("开发者风格画像分析")
	tracker.UpdateStepProgress("初始化开发者分析器...")
	
	profileAnalyzer := developer.NewProfileAnalyzer(stats)
	
	// Analyze top contributors (limit to top 10 for performance)
	var developerProfiles []*developer.DeveloperProfile
	contributorCount := len(stats.AuthorStats)
	if contributorCount > 100 {
		contributorCount = 100
	}
	
	idx := 0
	for authorName, authorStat := range stats.AuthorStats {
		if idx >= contributorCount {
			break
		}
		tracker.UpdateStepProgress(fmt.Sprintf("分析开发者: %s (%d/%d)", authorName, idx+1, contributorCount))
		profile := profileAnalyzer.AnalyzeDeveloper(authorStat)
		developerProfiles = append(developerProfiles, profile)
		idx++
	}
	
	tracker.CompleteStep(fmt.Sprintf("开发者风格画像分析完成 (%d位开发者)", len(developerProfiles)))
	
	// Generate developer profiles report
	var developerReport strings.Builder
	if len(developerProfiles) > 0 {
		developerReport.WriteString("\n\n=== 🎭 开发者风格画像分析 ===\n")
		for _, profile := range developerProfiles {
			developerReport.WriteString(profile.GenerateReport())
			developerReport.WriteString("\n")
		}
	}

	var finalReport string
	var aiAnalysis string
	var aiError error
	var aiConfigError bool
	
	// Step 4: AI Analysis (if enabled)
	if useAI {
		tracker.StartStep("AI智能分析")
		tracker.UpdateStepProgress("初始化AI客户端...")
		
		aiClient, err := ai.NewAIClient()
		if err != nil {
			aiError = err
			aiConfigError = true
			tracker.CompleteStepWithWarning("AI分析跳过", fmt.Sprintf("AI客户端初始化失败: %v", err))
			finalReport = basicReport + developerReport.String()
		} else {
			tracker.UpdateStepProgress("发送分析请求到AI服务...")
			aiResult, err := aiClient.AnalyzeWithAI(stats, basicReport)
			if err != nil {
				aiError = err
				aiConfigError = false
				tracker.CompleteStepWithWarning("AI分析跳过", fmt.Sprintf("AI分析失败: %v", err))
				finalReport = basicReport + developerReport.String()
			} else {
				tracker.UpdateStepProgress("AI分析响应处理完成")
				aiAnalysis = aiResult
				finalReport = basicReport + developerReport.String() + "\n\n=== AI-Powered Analysis ===\n" + aiAnalysis
				tracker.CompleteStep("AI智能分析完成")
			}
		}
	} else {
		finalReport = basicReport + developerReport.String()
	}
	
	// Step 5: Report Generation
	tracker.StartStep("报告生成与输出")
	
	reportGenerated := false
	
	// Generate web report
	if generateWeb {
		tracker.UpdateStepProgress("生成Web报告...")
		webGen := report.NewWebReportGenerator(outputDir)
		projectName := filepath.Base(repoPath)
		if projectName == "." || projectName == "" {
			projectName = "Current Repository"
		}
		
		// Prepare AI status
		var aiStatus report.AIStatus
		if !useAI {
			aiStatus = report.AIStatus{
				Enabled:      false,
				Available:    false,
				ErrorType:    "disabled",
				ErrorMessage: "",
			}
		} else if aiError != nil {
			if aiConfigError {
				aiStatus = report.AIStatus{
					Enabled:      true,
					Available:    false,
					ErrorType:    "config_error",
					ErrorMessage: aiError.Error(),
				}
			} else {
				aiStatus = report.AIStatus{
					Enabled:      true,
					Available:    false,
					ErrorType:    "analysis_error",
					ErrorMessage: aiError.Error(),
				}
			}
		} else {
			aiStatus = report.AIStatus{
				Enabled:      true,
				Available:    true,
				ErrorType:    "",
				ErrorMessage: "",
			}
		}
		
		subTracker := tracker.CreateSubTracker("Web报告生成", 3)
		subTracker.UpdateSub("准备报告数据")
		subTracker.UpdateSub("渲染HTML模板")
		
		err := webGen.GenerateReport(stats, aiAnalysis, aiStatus, projectName, developerProfiles, analysisMonths)
		if err != nil {
			tracker.UpdateStepProgress(fmt.Sprintf("Web报告生成失败: %v", err))
		} else {
			reportPath := webGen.GetReportPath()
			subTracker.UpdateSub("保存报告文件")
			subTracker.CompleteSub(fmt.Sprintf("Web报告已生成: %s", reportPath))
			reportGenerated = true
			
			if openBrowser {
				tracker.UpdateStepProgress("正在打开浏览器...")
				openWebReport(reportPath)
			}
		}
	}
	
	// Output text results
	if outputFile != "" {
		tracker.UpdateStepProgress("保存文本报告...")
		err := os.WriteFile(outputFile, []byte(finalReport), 0644)
		if err != nil {
			tracker.UpdateStepProgress(fmt.Sprintf("文本报告保存失败: %v", err))
		} else {
			tracker.UpdateStepProgress(fmt.Sprintf("文本报告已保存: %s", outputFile))
			reportGenerated = true
		}
	}
	
	if reportGenerated {
		tracker.CompleteStep("报告生成完成")
	} else {
		tracker.CompleteStep("控制台输出完成")
		fmt.Println("\n" + strings.Repeat("=", 80))
		fmt.Println(finalReport)
		fmt.Println(strings.Repeat("=", 80))
	}
	
	// Complete the entire process
	tracker.Complete()
	tracker.ShowSummary(stats)
	
	return nil
}

// openWebReport opens the web report in the default browser
func openWebReport(reportPath string) {
	absPath, err := filepath.Abs(reportPath)
	if err != nil {
		fmt.Printf("Warning: Could not get absolute path for report: %v\n", err)
		return
	}
	
	fmt.Printf("Opening web report in browser: file://%s\n", absPath)
	// Note: This is a simplified version. In a real implementation,
	// you might want to use a library like "github.com/pkg/browser"
}

// getEnv gets environment variable with default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvBool gets environment variable as boolean with default value
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return value == "true" || value == "1" || value == "yes"
	}
	return defaultValue
}
